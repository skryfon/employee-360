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

func TestGormDepartmentRepository_TenantIsolationAndCRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormDepartmentRepository(db)

	tenantA := createTestTenant(t, db, "dept-tenant-a-"+uuid.NewString())
	tenantB := createTestTenant(t, db, "dept-tenant-b-"+uuid.NewString())

	actorA := uuid.New()
	actorB := uuid.New()

	ctxA := context.Background()
	ctxB := context.Background()

	// 1. Create in Tenant A with actorA
	deptA := &entity.Department{
		Name:        "Engineering",
		Description: "Product and platform engineering team",
	}
	err := repo.Create(ctxA, tenantA.ID, actorA, deptA)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, deptA.ID)
	require.Equal(t, tenantA.ID, deptA.TenantID)
	require.Equal(t, &actorA, deptA.CreatedBy)
	require.Equal(t, &actorA, deptA.UpdatedBy)

	// 2. Tenant A can read it and verify audit columns
	gotA, err := repo.GetByID(ctxA, tenantA.ID, deptA.ID)
	require.NoError(t, err)
	require.Equal(t, deptA.Name, gotA.Name)
	require.Equal(t, deptA.Description, gotA.Description)
	require.Equal(t, &actorA, gotA.CreatedBy)
	require.Equal(t, &actorA, gotA.UpdatedBy)

	// 3. Tenant B cannot read it (404 / ErrDepartmentNotFound)
	_, err = repo.GetByID(ctxB, tenantB.ID, deptA.ID)
	require.ErrorIs(t, err, domainerrors.ErrDepartmentNotFound)

	// 4. Same name in different tenant succeeds (multi-tenant uniqueness)
	deptB := &entity.Department{
		Name:        "Engineering",
		Description: "Engineering team in Tenant B",
	}
	err = repo.Create(ctxB, tenantB.ID, actorB, deptB)
	require.NoError(t, err)
	require.NotEqual(t, deptA.ID, deptB.ID)
	require.Equal(t, tenantB.ID, deptB.TenantID)
	require.Equal(t, &actorB, deptB.CreatedBy)
	require.Equal(t, &actorB, deptB.UpdatedBy)

	// 5. Duplicate name in same tenant fails with ErrDepartmentNameTaken
	deptADup := &entity.Department{
		Name:        "Engineering",
		Description: "Duplicate name in Tenant A",
	}
	err = repo.Create(ctxA, tenantA.ID, actorA, deptADup)
	require.ErrorIs(t, err, domainerrors.ErrDepartmentNameTaken)

	// 6. ExistsByName check
	existsA, err := repo.ExistsByName(ctxA, tenantA.ID, "engineering")
	require.NoError(t, err)
	require.True(t, existsA)

	existsOther, err := repo.ExistsByName(ctxA, tenantA.ID, "NonExistent")
	require.NoError(t, err)
	require.False(t, existsOther)

	// 7. Update in Tenant A with actorB
	deptA.Name = "Product Engineering"
	deptA.Description = "Updated description"
	err = repo.Update(ctxA, tenantA.ID, actorB, deptA)
	require.NoError(t, err)

	updatedA, err := repo.GetByID(ctxA, tenantA.ID, deptA.ID)
	require.NoError(t, err)
	require.Equal(t, "Product Engineering", updatedA.Name)
	require.Equal(t, "Updated description", updatedA.Description)
	require.Equal(t, &actorA, updatedA.CreatedBy)
	require.Equal(t, &actorB, updatedA.UpdatedBy)

	// 8. Tenant B cannot update Tenant A's department
	err = repo.Update(ctxB, tenantB.ID, actorB, deptA)
	require.ErrorIs(t, err, domainerrors.ErrDepartmentNotFound)

	// 9. List is tenant-scoped
	listA, totalA, err := repo.List(ctxA, tenantA.ID, 10, 0)
	require.NoError(t, err)
	require.Equal(t, int64(1), totalA)
	require.Len(t, listA, 1)
	require.Equal(t, deptA.ID, listA[0].ID)

	listB, totalB, err := repo.List(ctxB, tenantB.ID, 10, 0)
	require.NoError(t, err)
	require.Equal(t, int64(1), totalB)
	require.Len(t, listB, 1)
	require.Equal(t, deptB.ID, listB[0].ID)

	// 10. Tenant B cannot delete Tenant A's department
	err = repo.Delete(ctxB, tenantB.ID, deptA.ID)
	require.ErrorIs(t, err, domainerrors.ErrDepartmentNotFound)

	// 11. Delete in Tenant A succeeds
	err = repo.Delete(ctxA, tenantA.ID, deptA.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctxA, tenantA.ID, deptA.ID)
	require.ErrorIs(t, err, domainerrors.ErrDepartmentNotFound)
}

func TestGormDepartmentRepository_IsReferenced(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormDepartmentRepository(db)

	tenant := createTestTenant(t, db, "dept-ref-tenant-"+uuid.NewString())
	actor := uuid.New()
	c := context.Background()

	dept := &entity.Department{
		Name: "Operations",
	}
	require.NoError(t, repo.Create(c, tenant.ID, actor, dept))

	// Initially not referenced
	referenced, err := repo.IsReferenced(c, tenant.ID, dept.ID)
	require.NoError(t, err)
	require.False(t, referenced)

	// Reference in users table
	user := &entity.User{
		ID:           uuid.New(),
		TenantID:     tenant.ID,
		DepartmentID: &dept.ID,
		Email:        "dept-user-" + uuid.NewString() + "@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	require.NoError(t, db.Create(user).Error)

	referenced, err = repo.IsReferenced(c, tenant.ID, dept.ID)
	require.NoError(t, err)
	require.True(t, referenced)

	// Remove user reference
	require.NoError(t, db.Delete(user).Error)

	referenced, err = repo.IsReferenced(c, tenant.ID, dept.ID)
	require.NoError(t, err)
	require.False(t, referenced)

	// Reference in user_invitations table
	role := &entity.Role{
		ID:        uuid.New(),
		TenantID:  tenant.ID,
		Name:      "role-" + uuid.NewString(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	require.NoError(t, db.Create(role).Error)

	inv := &entity.UserInvitation{
		ID:           uuid.New(),
		TenantID:     tenant.ID,
		Email:        "invited-" + uuid.NewString() + "@example.com",
		RoleID:       role.ID,
		DepartmentID: &dept.ID,
		InvitedBy:    user.ID, // DB requires invited_by FK to users, let's create a valid inviter user
	}
	inviter := &entity.User{
		ID:        uuid.New(),
		TenantID:  tenant.ID,
		Email:     "inviter-" + uuid.NewString() + "@example.com",
		FirstName: "Admin",
		LastName:  "User",
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	require.NoError(t, db.Create(inviter).Error)
	inv.InvitedBy = inviter.ID
	inv.TokenHash = "mock-token-hash-" + uuid.NewString()
	inv.ExpiresAt = time.Now().UTC().Add(48 * time.Hour)
	inv.CreatedAt = time.Now().UTC()
	inv.UpdatedAt = time.Now().UTC()
	require.NoError(t, db.Create(inv).Error)

	referenced, err = repo.IsReferenced(c, tenant.ID, dept.ID)
	require.NoError(t, err)
	require.True(t, referenced)
}
