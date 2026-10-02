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

func TestGormDepartmentRepository_CRUDAndTenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormDepartmentRepository(db)

	tenantA := createTestTenant(t, db, "dept-tenant-a-"+uuid.NewString())
	tenantB := createTestTenant(t, db, "dept-tenant-b-"+uuid.NewString())

	actorA := uuid.New()
	actorB := uuid.New()

	ctxA := context.Background()
	ctxB := context.Background()

	// 1. Create department in Tenant A
	deptA := &entity.Department{
		Name:        "Engineering",
		Description: "Core engineering team",
	}
	err := repo.Create(ctxA, tenantA.ID, actorA, deptA)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, deptA.ID)
	require.Equal(t, tenantA.ID, deptA.TenantID)
	require.Equal(t, &actorA, deptA.CreatedBy)
	require.Equal(t, &actorA, deptA.UpdatedBy)

	// 2. Tenant B can create a department with the same name (tenant-scoped uniqueness)
	deptB := &entity.Department{
		Name:        "Engineering",
		Description: "Tenant B engineering",
	}
	err = repo.Create(ctxB, tenantB.ID, actorB, deptB)
	require.NoError(t, err)
	require.NotEqual(t, deptA.ID, deptB.ID)

	// 3. Duplicate department name within the same tenant fails (case-insensitive check)
	deptADup := &entity.Department{
		Name: "engineering",
	}
	err = repo.Create(ctxA, tenantA.ID, actorA, deptADup)
	require.ErrorIs(t, err, domainerrors.ErrDepartmentNameTaken)

	// 4. ExistsByName check
	exists, err := repo.ExistsByName(ctxA, tenantA.ID, "Engineering")
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.ExistsByName(ctxA, tenantA.ID, "Marketing")
	require.NoError(t, err)
	require.False(t, exists)

	// 5. Tenant B cannot read Tenant A's department
	_, err = repo.GetByID(ctxB, tenantB.ID, deptA.ID)
	require.ErrorIs(t, err, domainerrors.ErrDepartmentNotFound)

	// 6. Tenant A reads own department
	foundA, err := repo.GetByID(ctxA, tenantA.ID, deptA.ID)
	require.NoError(t, err)
	require.Equal(t, "Engineering", foundA.Name)
	require.Equal(t, "Core engineering team", foundA.Description)

	// 7. Update department in Tenant A
	foundA.Name = "Product Engineering"
	foundA.Description = "Updated description"
	err = repo.Update(ctxA, tenantA.ID, actorB, foundA)
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
	err = repo.Delete(ctxB, tenantB.ID, deptA.ID, actorB)
	require.ErrorIs(t, err, domainerrors.ErrDepartmentNotFound)

	// 11. Delete in Tenant A succeeds (soft delete)
	err = repo.Delete(ctxA, tenantA.ID, deptA.ID, actorA)
	require.NoError(t, err)

	_, err = repo.GetByID(ctxA, tenantA.ID, deptA.ID)
	require.ErrorIs(t, err, domainerrors.ErrDepartmentNotFound)

	// Verify the department row is soft-deleted (deleted_at and deleted_by populated)
	var rawDept entity.Department
	err = db.Table("departments").Where("id = ?", deptA.ID).First(&rawDept).Error
	require.NoError(t, err)
	require.NotNil(t, rawDept.DeletedAt)
	require.Equal(t, &actorA, rawDept.DeletedBy)

	// 12. Soft-deleted department name can be reused in Tenant A (partial unique index test)
	reusedDept := &entity.Department{
		Name:        "Product Engineering", // same name as soft-deleted deptA
		Description: "Recreated department",
	}
	err = repo.Create(ctxA, tenantA.ID, actorA, reusedDept)
	require.NoError(t, err)
	require.NotEqual(t, deptA.ID, reusedDept.ID)

	// 13. Case-insensitive duplicate name within same tenant fails
	dupCaseDept := &entity.Department{
		Name: "product engineering", // same lower name
	}
	err = repo.Create(ctxA, tenantA.ID, actorA, dupCaseDept)
	require.ErrorIs(t, err, domainerrors.ErrDepartmentNameTaken)
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

func TestGormDepartmentRepository_SoftDeletePreservesSoftDeletedUserReferences(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormDepartmentRepository(db)

	tenant := createTestTenant(t, db, "dept-soft-pres-tenant-"+uuid.NewString())
	actor := uuid.New()
	c := context.Background()

	dept := &entity.Department{
		Name: "Finance",
	}
	require.NoError(t, repo.Create(c, tenant.ID, actor, dept))

	now := time.Now().UTC()
	// Soft-deleted user referencing this department
	user := &entity.User{
		ID:           uuid.New(),
		TenantID:     tenant.ID,
		DepartmentID: &dept.ID,
		Email:        "soft-deleted-user-" + uuid.NewString() + "@example.com",
		FirstName:    "Jane",
		LastName:     "Doe",
		IsActive:     false,
		DeletedAt:    &now,
		DeletedBy:    &actor,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	require.NoError(t, db.Create(user).Error)

	// Soft-deleted user does not block deletion because they are deleted_at IS NOT NULL
	referenced, err := repo.IsReferenced(c, tenant.ID, dept.ID)
	require.NoError(t, err)
	require.False(t, referenced)

	// Delete department (soft-delete)
	require.NoError(t, repo.Delete(c, tenant.ID, dept.ID, actor))

	// Verify user still references department_id (was NOT nulled out by FK cascade because it was not a hard delete)
	var reloadedUser entity.User
	require.NoError(t, db.Table("users").Where("id = ?", user.ID).First(&reloadedUser).Error)
	require.NotNil(t, reloadedUser.DepartmentID)
	require.Equal(t, dept.ID, *reloadedUser.DepartmentID)
}
