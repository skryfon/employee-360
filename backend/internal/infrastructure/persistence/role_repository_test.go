//go:build integration

package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/skryfon/employee360/backend/internal/ctx"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
)

func tenantCtx(id uuid.UUID) context.Context {
	return ctx.WithTenantID(context.Background(), id.String())
}

func TestGormRoleRepository_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRoleRepository(db)
	a := createTestTenant(t, db, "role-a-"+uuid.NewString())
	b := createTestTenant(t, db, "role-b-"+uuid.NewString())
	ca, cb, cn := tenantCtx(a.ID), tenantCtx(b.ID), context.Background()

	now := time.Now().UTC()
	role := &entity.Role{ID: uuid.New(), TenantID: a.ID, Name: "role-" + uuid.NewString(), CreatedAt: now, UpdatedAt: now}
	require.NoError(t, repo.Create(ca, role))

	// Happy path in own tenant.
	got, err := repo.GetByID(ca, role.ID)
	require.NoError(t, err)
	require.Equal(t, role.Name, got.Name)
	got, err = repo.GetByName(ca, role.Name)
	require.NoError(t, err)
	require.Equal(t, role.ID, got.ID)
	list, err := repo.List(ca)
	require.NoError(t, err)
	require.Len(t, list, 1)

	// Other tenant and no-tenant contexts see nothing.
	for name, c := range map[string]context.Context{"tenantB": cb, "noTenant": cn} {
		t.Run(name, func(t *testing.T) {
			_, err := repo.GetByID(c, role.ID)
			require.ErrorIs(t, err, domainerrors.ErrRoleNotFound)
			_, err = repo.GetByName(c, role.Name)
			require.ErrorIs(t, err, domainerrors.ErrRoleNotFound)
			l, err := repo.List(c)
			require.NoError(t, err)
			require.Empty(t, l)
			upd := &entity.Role{ID: role.ID, Name: "hijacked", UpdatedAt: time.Now().UTC()}
			require.ErrorIs(t, repo.Update(c, upd), domainerrors.ErrRoleNotFound)
			require.ErrorIs(t, repo.Delete(c, role.ID), domainerrors.ErrRoleNotFound)
		})
	}

	// Row untouched.
	got, err = repo.GetByID(ca, role.ID)
	require.NoError(t, err)
	require.Equal(t, role.Name, got.Name)

	// Own-tenant update and delete work.
	role.Name = "renamed-" + uuid.NewString()
	role.UpdatedAt = time.Now().UTC()
	require.NoError(t, repo.Update(ca, role))
	got, err = repo.GetByID(ca, role.ID)
	require.NoError(t, err)
	require.Equal(t, role.Name, got.Name)
	require.NoError(t, repo.Delete(ca, role.ID))
	_, err = repo.GetByID(ca, role.ID)
	require.ErrorIs(t, err, domainerrors.ErrRoleNotFound)
}
