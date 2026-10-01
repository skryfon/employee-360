package department

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	deptuc "github.com/skryfon/employee360/backend/internal/usecase/interface/department"
)

func TestUpdateDepartmentUseCase(t *testing.T) {
	tenantID := uuid.New()
	actorID := uuid.New()
	deptID := uuid.New()
	bg := context.Background()

	existing := &entity.Department{
		ID:          deptID,
		TenantID:    tenantID,
		Name:        "Old Name",
		Description: "Old Desc",
		CreatedAt:   time.Now().UTC().Add(-time.Hour),
		UpdatedAt:   time.Now().UTC().Add(-time.Hour),
	}

	t.Run("successful update with name change", func(t *testing.T) {
		repo := &mockDepartmentRepo{
			getByIDFn: func(ctx context.Context, tID, id uuid.UUID) (*entity.Department, error) {
				assert.Equal(t, tenantID, tID)
				return &entity.Department{
					ID:       deptID,
					TenantID: tenantID,
					Name:     "Old Name",
				}, nil
			},
			existsByNameFn: func(ctx context.Context, tID uuid.UUID, name string) (bool, error) {
				assert.Equal(t, tenantID, tID)
				assert.Equal(t, "New Name", name)
				return false, nil
			},
			updateFn: func(ctx context.Context, tID, aID uuid.UUID, d *entity.Department) error {
				assert.Equal(t, tenantID, tID)
				assert.Equal(t, actorID, aID)
				assert.Equal(t, "New Name", d.Name)
				assert.Equal(t, "New Desc", d.Description)
				return nil
			},
		}

		uc := NewUpdateDepartmentUseCase(repo)
		updated, err := uc.Execute(bg, tenantID, actorID, deptuc.UpdateDepartmentInput{
			ID:          deptID,
			Name:        "  New Name  ",
			Description: "  New Desc  ",
		})
		require.NoError(t, err)
		assert.Equal(t, "New Name", updated.Name)
		assert.Equal(t, "New Desc", updated.Description)
	})

	t.Run("successful update keeping same name (case-insensitive)", func(t *testing.T) {
		existsChecked := false
		repo := &mockDepartmentRepo{
			getByIDFn: func(ctx context.Context, tID, id uuid.UUID) (*entity.Department, error) {
				assert.Equal(t, tenantID, tID)
				return &entity.Department{
					ID:       deptID,
					TenantID: tenantID,
					Name:     "Old Name",
				}, nil
			},
			existsByNameFn: func(ctx context.Context, tID uuid.UUID, name string) (bool, error) {
				existsChecked = true
				return true, nil
			},
			updateFn: func(ctx context.Context, tID, aID uuid.UUID, d *entity.Department) error {
				assert.Equal(t, tenantID, tID)
				assert.Equal(t, actorID, aID)
				return nil
			},
		}

		uc := NewUpdateDepartmentUseCase(repo)
		updated, err := uc.Execute(bg, tenantID, actorID, deptuc.UpdateDepartmentInput{
			ID:          deptID,
			Name:        "old name",
			Description: "Updated desc only",
		})
		require.NoError(t, err)
		assert.False(t, existsChecked, "should not check ExistsByName if name unchanged")
		assert.Equal(t, "old name", updated.Name)
	})

	t.Run("name taken returns conflict", func(t *testing.T) {
		repo := &mockDepartmentRepo{
			getByIDFn: func(ctx context.Context, tID, id uuid.UUID) (*entity.Department, error) {
				return existing, nil
			},
			existsByNameFn: func(ctx context.Context, tID uuid.UUID, name string) (bool, error) {
				return true, nil
			},
		}

		uc := NewUpdateDepartmentUseCase(repo)
		_, err := uc.Execute(bg, tenantID, actorID, deptuc.UpdateDepartmentInput{
			ID:   deptID,
			Name: "Taken Name",
		})
		require.ErrorIs(t, err, domainerrors.ErrDepartmentNameTaken)
	})

	t.Run("nil uuid returns ErrDepartmentNotFound", func(t *testing.T) {
		uc := NewUpdateDepartmentUseCase(&mockDepartmentRepo{})
		_, err := uc.Execute(bg, tenantID, actorID, deptuc.UpdateDepartmentInput{
			ID:   uuid.Nil,
			Name: "Some Name",
		})
		require.ErrorIs(t, err, domainerrors.ErrDepartmentNotFound)
	})

	t.Run("not found in repo returns ErrDepartmentNotFound", func(t *testing.T) {
		repo := &mockDepartmentRepo{
			getByIDFn: func(ctx context.Context, tID, id uuid.UUID) (*entity.Department, error) {
				return nil, domainerrors.ErrDepartmentNotFound
			},
		}
		uc := NewUpdateDepartmentUseCase(repo)
		_, err := uc.Execute(bg, tenantID, actorID, deptuc.UpdateDepartmentInput{
			ID:   deptID,
			Name: "Some Name",
		})
		require.ErrorIs(t, err, domainerrors.ErrDepartmentNotFound)
	})

	t.Run("empty name returns validation error", func(t *testing.T) {
		uc := NewUpdateDepartmentUseCase(&mockDepartmentRepo{})
		_, err := uc.Execute(bg, tenantID, actorID, deptuc.UpdateDepartmentInput{
			ID:   deptID,
			Name: "  ",
		})
		require.ErrorIs(t, err, ErrDepartmentNameRequired)
	})

	t.Run("name too long returns validation error", func(t *testing.T) {
		uc := NewUpdateDepartmentUseCase(&mockDepartmentRepo{})
		_, err := uc.Execute(bg, tenantID, actorID, deptuc.UpdateDepartmentInput{
			ID:   deptID,
			Name: strings.Repeat("x", 101),
		})
		require.ErrorIs(t, err, ErrDepartmentNameTooLong)
	})

	t.Run("nil tenant or actor returns unauthorized", func(t *testing.T) {
		uc := NewUpdateDepartmentUseCase(&mockDepartmentRepo{})
		_, err := uc.Execute(bg, uuid.Nil, actorID, deptuc.UpdateDepartmentInput{
			ID:   deptID,
			Name: "Valid Name",
		})
		require.ErrorIs(t, err, domainerrors.ErrUnauthorized)

		_, err = uc.Execute(bg, tenantID, uuid.Nil, deptuc.UpdateDepartmentInput{
			ID:   deptID,
			Name: "Valid Name",
		})
		require.ErrorIs(t, err, domainerrors.ErrUnauthorized)
	})
}
