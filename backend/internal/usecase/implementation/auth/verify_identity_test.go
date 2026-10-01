package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
)

type viTenantReader struct {
	tenant *entity.Tenant
	err    error
}

func (r *viTenantReader) GetByID(_ context.Context, id uuid.UUID) (*entity.Tenant, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.tenant == nil || r.tenant.ID != id {
		return nil, domainerrors.ErrTenantNotFound
	}
	return r.tenant, nil
}

// viUserRepo embeds the interface so only GetByID needs implementing.
type viUserRepo struct {
	repository.UserRepository
	user      *entity.User
	err       error
	gotTenant uuid.UUID
	calls     int
}

func (r *viUserRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (*entity.User, error) {
	r.calls++
	r.gotTenant = tenantID
	if r.err != nil {
		return nil, r.err
	}
	// Mirror the real adapter: scoped by tenant, hides deleted users.
	if r.user == nil || r.user.ID != id || r.user.TenantID != tenantID || r.user.DeletedAt != nil {
		return nil, domainerrors.ErrUserNotFound
	}
	return r.user, nil
}

type viUserRoleRepo struct {
	repository.UserRoleRepository
	roles []*entity.Role
	err   error
}

func (r *viUserRoleRepo) GetRolesByUserID(context.Context, uuid.UUID, uuid.UUID) ([]*entity.Role, error) {
	return r.roles, r.err
}

func TestVerifyIdentityUseCase(t *testing.T) {
	tenantID, userID := uuid.New(), uuid.New()
	now := time.Now()
	dbErr := errors.New("connection refused")

	okTenant := func() *entity.Tenant { return &entity.Tenant{ID: tenantID, IsActive: true} }
	okUser := func() *entity.User { return &entity.User{ID: userID, TenantID: tenantID, IsActive: true} }

	tests := []struct {
		name      string
		tenant    *viTenantReader
		user      *viUserRepo
		roles     *viUserRoleRepo
		wantErr   error // matched with errors.Is
		wantRoles []string
	}{
		{"valid, current roles loaded", &viTenantReader{tenant: okTenant()}, &viUserRepo{user: okUser()},
			&viUserRoleRepo{roles: []*entity.Role{{Name: "admin"}, {Name: "employee"}}}, nil, []string{"admin", "employee"}},
		{"valid, no roles", &viTenantReader{tenant: okTenant()}, &viUserRepo{user: okUser()}, &viUserRoleRepo{}, nil, []string{}},
		{"tenant missing", &viTenantReader{}, &viUserRepo{user: okUser()}, &viUserRoleRepo{}, domainerrors.ErrUnauthorized, nil},
		{"tenant inactive", &viTenantReader{tenant: &entity.Tenant{ID: tenantID, IsActive: false}}, &viUserRepo{user: okUser()}, &viUserRoleRepo{}, domainerrors.ErrUnauthorized, nil},
		{"tenant soft-deleted", &viTenantReader{tenant: &entity.Tenant{ID: tenantID, IsActive: true, DeletedAt: &now}}, &viUserRepo{user: okUser()}, &viUserRoleRepo{}, domainerrors.ErrUnauthorized, nil},
		{"user unknown", &viTenantReader{tenant: okTenant()}, &viUserRepo{}, &viUserRoleRepo{}, domainerrors.ErrUnauthorized, nil},
		{"user in another tenant", &viTenantReader{tenant: okTenant()}, &viUserRepo{user: &entity.User{ID: userID, TenantID: uuid.New(), IsActive: true}}, &viUserRoleRepo{}, domainerrors.ErrUnauthorized, nil},
		{"user inactive", &viTenantReader{tenant: okTenant()}, &viUserRepo{user: &entity.User{ID: userID, TenantID: tenantID, IsActive: false}}, &viUserRoleRepo{}, domainerrors.ErrUnauthorized, nil},
		{"user soft-deleted", &viTenantReader{tenant: okTenant()}, &viUserRepo{user: &entity.User{ID: userID, TenantID: tenantID, IsActive: true, DeletedAt: &now}}, &viUserRoleRepo{}, domainerrors.ErrUnauthorized, nil},
		{"tenant repo error is not unauthorized", &viTenantReader{err: dbErr}, &viUserRepo{user: okUser()}, &viUserRoleRepo{}, dbErr, nil},
		{"user repo error is not unauthorized", &viTenantReader{tenant: okTenant()}, &viUserRepo{err: dbErr}, &viUserRoleRepo{}, dbErr, nil},
		{"roles repo error is not unauthorized", &viTenantReader{tenant: okTenant()}, &viUserRepo{user: okUser()}, &viUserRoleRepo{err: dbErr}, dbErr, nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewVerifyIdentityUseCase(tc.tenant, tc.user, tc.roles)
			got, err := uc.Execute(context.Background(), tenantID, userID)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				if tc.wantErr == dbErr && errors.Is(err, domainerrors.ErrUnauthorized) {
					t.Fatal("infrastructure error must not be reported as unauthorized")
				}
				if got != nil {
					t.Fatal("expected nil identity on error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.TenantID != tenantID || got.UserID != userID {
				t.Errorf("identity = %+v", got)
			}
			if len(got.Roles) != len(tc.wantRoles) {
				t.Fatalf("roles = %v, want %v", got.Roles, tc.wantRoles)
			}
			for i, r := range tc.wantRoles {
				if got.Roles[i] != r {
					t.Errorf("roles = %v, want %v", got.Roles, tc.wantRoles)
				}
			}
			if tc.user.gotTenant != tenantID {
				t.Errorf("user lookup scoped to %s, want %s", tc.user.gotTenant, tenantID)
			}
		})
	}
}

func TestVerifyIdentityUseCase_NilUUIDsRejectedWithoutLookup(t *testing.T) {
	users := &viUserRepo{}
	uc := NewVerifyIdentityUseCase(&viTenantReader{}, users, &viUserRoleRepo{})
	if _, err := uc.Execute(context.Background(), uuid.Nil, uuid.New()); !errors.Is(err, domainerrors.ErrUnauthorized) {
		t.Errorf("nil tenant: err = %v", err)
	}
	if _, err := uc.Execute(context.Background(), uuid.New(), uuid.Nil); !errors.Is(err, domainerrors.ErrUnauthorized) {
		t.Errorf("nil user: err = %v", err)
	}
	if users.calls != 0 {
		t.Error("repository must not be hit for nil UUIDs")
	}
}
