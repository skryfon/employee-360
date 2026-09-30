package invitation

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/ctx"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/event"
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
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		users:       map[uuid.UUID]*entity.User{},
		invitations: map[uuid.UUID]*entity.UserInvitation{},
		roles:       map[uuid.UUID]*entity.Role{},
	}
}

type snapshot struct {
	users       map[uuid.UUID]entity.User
	userRoles   int
	invitations map[uuid.UUID]entity.UserInvitation
	events      int
}

func (s *fakeStore) snap() snapshot {
	sn := snapshot{users: map[uuid.UUID]entity.User{}, invitations: map[uuid.UUID]entity.UserInvitation{}, userRoles: len(s.userRoles), events: len(s.events)}
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
func (s *fakeStore) Delete(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (s *fakeStore) List(context.Context, uuid.UUID, int, int) ([]*entity.User, int64, error) {
	return nil, 0, nil
}

// RoleRepository (separate type: method names collide with UserRepository)
type fakeRoleRepo struct{ s *fakeStore }

func (r fakeRoleRepo) Create(context.Context, *entity.Role) error { return nil }
func (r fakeRoleRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.Role, error) {
	if ro, ok := r.s.roles[id]; ok {
		return ro, nil
	}
	return nil, domainerrors.ErrRoleNotFound
}
func (r fakeRoleRepo) GetByName(context.Context, string) (*entity.Role, error) {
	return nil, errors.New("unused")
}
func (r fakeRoleRepo) List(context.Context) ([]*entity.Role, error) { return nil, nil }
func (r fakeRoleRepo) Update(context.Context, *entity.Role) error   { return nil }
func (r fakeRoleRepo) Delete(context.Context, uuid.UUID) error      { return nil }

// UserRoleRepository
type fakeUserRoleRepo struct{ s *fakeStore }

func (r fakeUserRoleRepo) AssignRole(_ context.Context, ur *entity.UserRole) error {
	r.s.userRoles = append(r.s.userRoles, ur)
	return nil
}
func (r fakeUserRoleRepo) RemoveRole(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (r fakeUserRoleRepo) GetRolesByUserID(context.Context, uuid.UUID) ([]*entity.Role, error) {
	return nil, nil
}
func (r fakeUserRoleRepo) GetUserRolesByUserID(context.Context, uuid.UUID) ([]*entity.UserRole, error) {
	return nil, nil
}
func (r fakeUserRoleRepo) DeleteByUserID(context.Context, uuid.UUID) error { return nil }

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
func (r fakeInvRepo) UpdateToken(_ context.Context, t, id uuid.UUID, h string, exp time.Time) error {
	i := r.s.invitations[id]
	i.TokenHash, i.ExpiresAt = h, exp
	return nil
}
func (r fakeInvRepo) MarkAccepted(_ context.Context, t, id uuid.UUID, at time.Time) error {
	r.s.invitations[id].AcceptedAt = &at
	return nil
}
func (r fakeInvRepo) MarkRevoked(_ context.Context, t, id uuid.UUID, at time.Time) error {
	r.s.invitations[id].RevokedAt = &at
	return nil
}
func (r fakeInvRepo) List(_ context.Context, t uuid.UUID, limit, offset int) ([]*entity.UserInvitation, int64, error) {
	var out []*entity.UserInvitation
	for _, i := range r.s.invitations {
		if i.TenantID == t {
			out = append(out, i)
		}
	}
	return out, int64(len(out)), nil
}

// fixture helpers
type fixture struct {
	s          *fakeStore
	tenantA    uuid.UUID
	tenantB    uuid.UUID
	adminID    uuid.UUID
	employeeRl *entity.Role
	adminCtx   context.Context
	hash       domainservice.HashService
}

func newFixture() *fixture {
	f := &fixture{s: newFakeStore(), tenantA: uuid.New(), tenantB: uuid.New(), adminID: uuid.New(), hash: infraservice.NewHashService(4)}
	f.employeeRl = &entity.Role{ID: uuid.New(), TenantID: f.tenantA, Name: "employee"}
	f.s.roles[f.employeeRl.ID] = f.employeeRl
	f.adminCtx = adminCtx(f.tenantA, f.adminID)
	return f
}

func adminCtx(tenant, user uuid.UUID) context.Context {
	c := ctx.WithTenantID(context.Background(), tenant.String())
	c = ctx.WithUserID(c, user.String())
	return ctx.WithRoles(c, []string{"admin"})
}
