package role

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/ctx"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRoleRepo struct {
	roles []*entity.Role
	err   error
}

func (f fakeRoleRepo) Create(context.Context, *entity.Role) error { return nil }
func (f fakeRoleRepo) GetByID(context.Context, uuid.UUID) (*entity.Role, error) {
	return nil, errors.New("unused")
}
func (f fakeRoleRepo) GetByName(context.Context, string) (*entity.Role, error) {
	return nil, errors.New("unused")
}
func (f fakeRoleRepo) List(context.Context) ([]*entity.Role, error) { return f.roles, f.err }
func (f fakeRoleRepo) Update(context.Context, *entity.Role) error   { return nil }
func (f fakeRoleRepo) Delete(context.Context, uuid.UUID) error      { return nil }

func adminCtx(tenant uuid.UUID, roles ...string) context.Context {
	c := ctx.WithTenantID(context.Background(), tenant.String())
	return ctx.WithRoles(c, roles)
}

func TestListAssignableRoles_FiltersSuperAdminAndOtherTenants(t *testing.T) {
	tenant, other := uuid.New(), uuid.New()
	repo := fakeRoleRepo{roles: []*entity.Role{
		{ID: uuid.New(), TenantID: tenant, Name: "admin"},
		{ID: uuid.New(), TenantID: tenant, Name: "employee"},
		{ID: uuid.New(), TenantID: tenant, Name: "super_admin"},
		{ID: uuid.New(), TenantID: other, Name: "leaked"},
	}}
	uc := NewListAssignableRolesUseCase(repo)

	for _, role := range []string{"admin", "super_admin"} {
		got, err := uc.Execute(adminCtx(tenant, role))
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "admin", got[0].Name)
		assert.Equal(t, "employee", got[1].Name)
	}
}

func TestListAssignableRoles_Authorization(t *testing.T) {
	uc := NewListAssignableRolesUseCase(fakeRoleRepo{})

	_, err := uc.Execute(adminCtx(uuid.New(), "employee"))
	assert.ErrorIs(t, err, domainerrors.ErrForbidden)

	_, err = uc.Execute(ctx.WithRoles(context.Background(), []string{"admin"}))
	assert.ErrorIs(t, err, domainerrors.ErrUnauthorized)
}

func TestListAssignableRoles_RepoError(t *testing.T) {
	boom := errors.New("boom")
	uc := NewListAssignableRolesUseCase(fakeRoleRepo{err: boom})
	_, err := uc.Execute(adminCtx(uuid.New(), "admin"))
	assert.ErrorIs(t, err, boom)
}
