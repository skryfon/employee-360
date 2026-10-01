//go:build integration

package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
)

// Rows carrying deleted_at must disappear from role, user-role, invitation and
// org-reference reads.
func TestSoftDeletedRowsAreHiddenFromReads(t *testing.T) {
	f := newInvFixture(t)
	c := context.Background()
	db := f.db
	tid := f.tenantA.ID
	now := time.Now().UTC()

	// Roles (f.roleA is "admin"; add one and soft-delete it).
	roles := NewGormRoleRepository(db)
	gone := &entity.Role{ID: uuid.New(), TenantID: tid, Name: "gone-" + uuid.NewString(), CreatedAt: now, UpdatedAt: now}
	require.NoError(t, roles.Create(c, gone))
	require.NoError(t, db.Exec("UPDATE roles SET deleted_at = NOW() WHERE id = ?", gone.ID).Error)
	_, err := roles.GetByID(c, tid, gone.ID)
	require.ErrorIs(t, err, domainerrors.ErrRoleNotFound)
	_, err = roles.GetByName(c, tid, gone.Name)
	require.ErrorIs(t, err, domainerrors.ErrRoleNotFound)
	list, err := roles.List(c, tid)
	require.NoError(t, err)
	for _, r := range list {
		require.NotEqual(t, gone.ID, r.ID)
	}

	// User roles hide soft-deleted roles.
	urRepo := NewGormUserRoleRepository(db)
	require.NoError(t, urRepo.AssignRole(c, &entity.UserRole{ID: uuid.New(), TenantID: tid, UserID: f.inviterA, RoleID: gone.ID, CreatedAt: now, UpdatedAt: now}))
	got, err := urRepo.GetRolesByUserID(c, tid, f.inviterA)
	require.NoError(t, err)
	require.Empty(t, got)

	// Invitations.
	inv := f.mustCreate(t, now)
	require.NoError(t, db.Exec("UPDATE user_invitations SET deleted_at = NOW() WHERE id = ?", inv.ID).Error)
	_, err = f.repo.GetByID(c, tid, inv.ID)
	require.ErrorIs(t, err, domainerrors.ErrInvitationNotFound)
	_, err = f.repo.GetByTokenHash(c, inv.TokenHash)
	require.ErrorIs(t, err, domainerrors.ErrInvitationNotFound)
	invs, total, err := f.repo.List(c, tid, 0, 0)
	require.NoError(t, err)
	require.Zero(t, total)
	require.Empty(t, invs)
	require.ErrorIs(t, f.repo.MarkRevoked(c, tid, inv.ID, f.inviterA, now), domainerrors.ErrInvitationNotPending)

	// Departments / positions.
	org := NewGormOrgReferenceRepository(db)
	dept, pos := uuid.New(), uuid.New()
	require.NoError(t, db.Exec("INSERT INTO departments (id, tenant_id, name, deleted_at) VALUES (?, ?, 'D', NOW())", dept, tid).Error)
	require.NoError(t, db.Exec("INSERT INTO positions (id, tenant_id, name, deleted_at) VALUES (?, ?, 'P', NOW())", pos, tid).Error)
	ok, err := org.DepartmentExists(c, tid, dept)
	require.NoError(t, err)
	require.False(t, ok)
	ok, err = org.PositionExists(c, tid, pos)
	require.NoError(t, err)
	require.False(t, ok)
}
