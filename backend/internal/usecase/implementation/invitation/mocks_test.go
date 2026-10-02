package invitation

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/event"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	domainservice "github.com/skryfon/employee360/backend/internal/domain/service"
	infraservice "github.com/skryfon/employee360/backend/internal/infrastructure/service"
)

// fakeStore is an in-memory stand-in for the DB; snapshot/restore emulate
// transaction rollback so we can assert the outbox guarantee.
type fakeStore struct {
	users       map[uuid.UUID]*entity.User
	userRoles   []*entity.UserRole
	invitations map[uuid.UUID]*entity.UserInvitation
	roles       map[uuid.UUID]*entity.Role
	events      []event.Event
	publishErr  error
	audits      []*entity.AuditLog
	// depts/positions map id -> owning tenant.
	depts     map[uuid.UUID]uuid.UUID
	positions map[uuid.UUID]uuid.UUID
	// domains maps lowercased domain -> owning tenant.
	domains map[string]uuid.UUID
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		users:       map[uuid.UUID]*entity.User{},
		invitations: map[uuid.UUID]*entity.UserInvitation{},
		roles:       map[uuid.UUID]*entity.Role{},
		depts:       map[uuid.UUID]uuid.UUID{},
		positions:   map[uuid.UUID]uuid.UUID{},
		domains:     map[string]uuid.UUID{},
	}
}

type snapshot struct {
	users       map[uuid.UUID]entity.User
	userRoles   int
	invitations map[uuid.UUID]entity.UserInvitation
	events      int
	audits      int
}

func (s *fakeStore) snap() snapshot {
	sn := snapshot{users: map[uuid.UUID]entity.User{}, invitations: map[uuid.UUID]entity.UserInvitation{}, userRoles: len(s.userRoles), events: len(s.events), audits: len(s.audits)}
	for k, v := range s.users {
		sn.users[k] = *v
	}
	for k, v := range s.invitations {
		sn.invitations[k] = *v
	}
	return sn
}

func (s *fakeStore) restore(sn snapshot) {
	s.users = map[uuid.UUID]*entity.User{}
	for k, v := range sn.users {
		v := v
		s.users[k] = &v
	}
	s.invitations = map[uuid.UUID]*entity.UserInvitation{}
	for k, v := range sn.invitations {
		v := v
		s.invitations[k] = &v
	}
	s.userRoles = s.userRoles[:sn.userRoles]
	s.events = s.events[:sn.events]
	s.audits = s.audits[:sn.audits]
}

// Transactor: rolls the store back if fn errors.
func (s *fakeStore) WithinTransaction(c context.Context, fn func(context.Context) error) error {
	sn := s.snap()
	if err := fn(c); err != nil {
		s.restore(sn)
		return err
	}
	return nil
}

func (s *fakeStore) Publish(_ context.Context, evs ...event.Event) error {
	s.events = append(s.events, evs...)
	return s.publishErr
}

// UserRepository
func (s *fakeStore) Create(_ context.Context, u *entity.User) error { s.users[u.ID] = u; return nil }
func (s *fakeStore) GetByID(_ context.Context, t, id uuid.UUID) (*entity.User, error) {
	if u, ok := s.users[id]; ok && u.TenantID == t {
		return u, nil
	}
	return nil, domainerrors.ErrUserNotFound
}
func (s *fakeStore) GetByTenantAndEmail(_ context.Context, t uuid.UUID, e string) (*entity.User, error) {
	for _, u := range s.users {
		if u.TenantID == t && u.Email == e {
			return u, nil
		}
	}
	return nil, domainerrors.ErrUserNotFound
}
func (s *fakeStore) GetByIDWithRoles(c context.Context, t, id uuid.UUID) (*entity.User, error) {
	return s.GetByID(c, t, id)
}
func (s *fakeStore) GetByTenantAndEmailWithRoles(c context.Context, t uuid.UUID, e string) (*entity.User, error) {
	return s.GetByTenantAndEmail(c, t, e)
}
func (s *fakeStore) Update(_ context.Context, t uuid.UUID, u *entity.User) error {
	s.users[u.ID] = u
	return nil
}
func (s *fakeStore) Delete(_ context.Context, t, id, _ uuid.UUID) error {
	if u, ok := s.users[id]; ok && u.TenantID == t {
		delete(s.users, id)
	}
	return nil
}
func (s *fakeStore) List(context.Context, uuid.UUID, int, int) ([]*entity.User, int64, error) {
	return nil, 0, nil
}

// RoleRepository (separate type: method names collide with UserRepository)
type fakeRoleRepo struct{ s *fakeStore }

func (r fakeRoleRepo) Create(context.Context, *entity.Role) error { return nil }
func (r fakeRoleRepo) GetByID(_ context.Context, t, id uuid.UUID) (*entity.Role, error) {
	if ro, ok := r.s.roles[id]; ok && ro.TenantID == t {
		return ro, nil
	}
	return nil, domainerrors.ErrRoleNotFound
}
func (r fakeRoleRepo) GetByName(context.Context, uuid.UUID, string) (*entity.Role, error) {
	return nil, errors.New("unused")
}
func (r fakeRoleRepo) List(context.Context, uuid.UUID) ([]*entity.Role, error) { return nil, nil }
func (r fakeRoleRepo) Update(context.Context, uuid.UUID, *entity.Role) error   { return nil }
func (r fakeRoleRepo) Delete(context.Context, uuid.UUID, uuid.UUID) error      { return nil }

// UserRoleRepository
type fakeUserRoleRepo struct{ s *fakeStore }

func (r fakeUserRoleRepo) AssignRole(_ context.Context, ur *entity.UserRole) error {
	r.s.userRoles = append(r.s.userRoles, ur)
	return nil
}
func (r fakeUserRoleRepo) RemoveRole(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	return nil
}
func (r fakeUserRoleRepo) GetRolesByUserID(context.Context, uuid.UUID, uuid.UUID) ([]*entity.Role, error) {
	return nil, nil
}
func (r fakeUserRoleRepo) GetUserRolesByUserID(context.Context, uuid.UUID, uuid.UUID) ([]*entity.UserRole, error) {
	return nil, nil
}
func (r fakeUserRoleRepo) DeleteByUserID(_ context.Context, _, uid uuid.UUID) error {
	kept := r.s.userRoles[:0:0]
	for _, ur := range r.s.userRoles {
		if ur.UserID != uid {
			kept = append(kept, ur)
		}
	}
	r.s.userRoles = kept
	return nil
}

// AuditRepository
type fakeAuditRepo struct{ s *fakeStore }

func (r fakeAuditRepo) Create(_ context.Context, l *entity.AuditLog) error {
	r.s.audits = append(r.s.audits, l)
	return nil
}
func (r fakeAuditRepo) ListByTenantID(context.Context, uuid.UUID, int, int) ([]*entity.AuditLog, int64, error) {
	return nil, 0, nil
}
func (r fakeAuditRepo) ListByEntity(context.Context, uuid.UUID, string, uuid.UUID, int, int) ([]*entity.AuditLog, int64, error) {
	return nil, 0, nil
}

// OrgReferenceRepository
type fakeOrgRepo struct{ s *fakeStore }

func (r fakeOrgRepo) DepartmentExists(_ context.Context, t, id uuid.UUID) (bool, error) {
	owner, ok := r.s.depts[id]
	return ok && owner == t, nil
}
func (r fakeOrgRepo) PositionExists(_ context.Context, t, id uuid.UUID) (bool, error) {
	owner, ok := r.s.positions[id]
	return ok && owner == t, nil
}

// UserInvitationRepository
type fakeInvRepo struct{ s *fakeStore }

func (r fakeInvRepo) Create(_ context.Context, i *entity.UserInvitation) error {
	r.s.invitations[i.ID] = i
	return nil
}
func (r fakeInvRepo) GetByID(_ context.Context, t, id uuid.UUID) (*entity.UserInvitation, error) {
	if i, ok := r.s.invitations[id]; ok && i.TenantID == t {
		cp := *i
		return &cp, nil
	}
	return nil, domainerrors.ErrNotFound
}
func (r fakeInvRepo) GetByTokenHash(_ context.Context, h string) (*entity.UserInvitation, error) {
	for _, i := range r.s.invitations {
		if i.TokenHash == h {
			cp := *i
			return &cp, nil
		}
	}
	return nil, domainerrors.ErrNotFound
}
func (r fakeInvRepo) pending(t, id uuid.UUID) (*entity.UserInvitation, error) {
	i, ok := r.s.invitations[id]
	if !ok || i.TenantID != t || !i.IsPending() {
		return nil, domainerrors.ErrInvitationNotPending
	}
	return i, nil
}
func (r fakeInvRepo) UpdateToken(_ context.Context, t, id, actor uuid.UUID, h string, exp time.Time) error {
	i, err := r.pending(t, id)
	if err != nil {
		return err
	}
	i.TokenHash, i.ExpiresAt = h, exp
	i.UpdatedBy = &actor
	return nil
}
func (r fakeInvRepo) MarkAccepted(_ context.Context, t, id uuid.UUID, at time.Time) error {
	i, err := r.pending(t, id)
	if err != nil {
		return err
	}
	i.AcceptedAt = &at
	return nil
}
func (r fakeInvRepo) MarkRevoked(_ context.Context, t, id, actor uuid.UUID, at time.Time) error {
	i, err := r.pending(t, id)
	if err != nil {
		return err
	}
	i.RevokedAt = &at
	i.UpdatedBy = &actor
	return nil
}
func (r fakeInvRepo) List(_ context.Context, t uuid.UUID, _ repository.InvitationListFilter) ([]*entity.InvitationListItem, int64, error) {
	var out []*entity.InvitationListItem
	for _, i := range r.s.invitations {
		if i.TenantID == t {
			out = append(out, &entity.InvitationListItem{UserInvitation: *i})
		}
	}
	return out, int64(len(out)), nil
}

type fakeTenantDomainRepo struct{ s *fakeStore }

func (r fakeTenantDomainRepo) FindTenantByDomain(context.Context, string) (*entity.Tenant, error) {
	return nil, domainerrors.ErrTenantNotFound
}
func (r fakeTenantDomainRepo) DomainBelongsToTenant(_ context.Context, t uuid.UUID, d string) (bool, error) {
	owner, ok := r.s.domains[d]
	return ok && owner == t, nil
}

// fixture helpers
type fixture struct {
	s          *fakeStore
	tenantA    uuid.UUID
	tenantB    uuid.UUID
	adminID    uuid.UUID
	employeeRl *entity.Role
	hash       domainservice.HashService
}

func newFixture() *fixture {
	f := &fixture{s: newFakeStore(), tenantA: uuid.New(), tenantB: uuid.New(), adminID: uuid.New(), hash: infraservice.NewHashService(4)}
	f.employeeRl = &entity.Role{ID: uuid.New(), TenantID: f.tenantA, Name: "employee"}
	f.s.roles[f.employeeRl.ID] = f.employeeRl
	f.s.domains["acme.com"] = f.tenantA
	f.s.domains["b.com"] = f.tenantB
	return f
}

// bg is the plain background context; identity is passed to usecases as explicit params.
var bg = context.Background()
