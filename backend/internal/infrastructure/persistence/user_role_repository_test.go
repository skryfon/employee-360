//go:build integration

package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

func TestGormUserRoleRepository_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormUserRoleRepository(db)
	a := createTestTenant(t, db, "ur-a-"+uuid.NewString())
	b := createTestTenant(t, db, "ur-b-"+uuid.NewString())
	ca, cb, cn := tenantCtx(a.ID), tenantCtx(b.ID), context.Background()

	now := time.Now().UTC()
	role := &entity.Role{ID: uuid.New(), TenantID: a.ID, Name: "r-" + uuid.NewString(), CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(role).Error)
	user := createTestUser(t, db, a.ID, "ur+"+uuid.NewString()+"@example.com")
	ur := &entity.UserRole{ID: uuid.New(), TenantID: a.ID, UserID: user.ID, RoleID: role.ID, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, repo.AssignRole(ca, ur))

	// Happy path.
	roles, err := repo.GetRolesByUserID(ca, user.ID)
	require.NoError(t, err)
	require.Len(t, roles, 1)
	require.Equal(t, role.ID, roles[0].ID)
	urs, err := repo.GetUserRolesByUserID(ca, user.ID)
	require.NoError(t, err)
	require.Len(t, urs, 1)

	for name, c := range map[string]context.Context{"tenantB": cb, "noTenant": cn} {
		t.Run(name, func(t *testing.T) {
			roles, err := repo.GetRolesByUserID(c, user.ID)
			require.NoError(t, err)
			require.Empty(t, roles)
			urs, err := repo.GetUserRolesByUserID(c, user.ID)
			require.NoError(t, err)
			require.Empty(t, urs)
			require.NoError(t, repo.RemoveRole(c, user.ID, role.ID))
			require.NoError(t, repo.DeleteByUserID(c, user.ID))
		})
	}

	// Cross-tenant deletes were no-ops.
	urs, err = repo.GetUserRolesByUserID(ca, user.ID)
	require.NoError(t, err)
	require.Len(t, urs, 1)

	// Own-tenant removal works.
	require.NoError(t, repo.RemoveRole(ca, user.ID, role.ID))
	urs, err = repo.GetUserRolesByUserID(ca, user.ID)
	require.NoError(t, err)
	require.Empty(t, urs)

	require.NoError(t, repo.AssignRole(ca, &entity.UserRole{ID: uuid.New(), TenantID: a.ID, UserID: user.ID, RoleID: role.ID, CreatedAt: now, UpdatedAt: now}))
	require.NoError(t, repo.DeleteByUserID(ca, user.ID))
	urs, err = repo.GetUserRolesByUserID(ca, user.ID)
	require.NoError(t, err)
	require.Empty(t, urs)
}
