package department

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	deptuc "github.com/skryfon/employee360/backend/internal/usecase/interface/department"
)

func TestDeleteDepartmentUseCase(t *testing.T) {
	tenantID := uuid.New()
	deptID := uuid.New()
	bg := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		repo := &mockDepartmentRepo{
			getByIDFn: func(ctx context.Context, tID, id uuid.UUID) (*entity.Department, error) {
				assert.Equal(t, tenantID, tID)
				return &entity.Department{ID: deptID, TenantID: tenantID}, nil
			},
			isReferencedFn: func(ctx context.Context, tID, id uuid.UUID) (bool, error) {
				assert.Equal(t, tenantID, tID)
				return false, nil
			},
			deleteFn: func(ctx context.Context, tID, id uuid.UUID) error {
				assert.Equal(t, tenantID, tID)
				assert.Equal(t, deptID, id)
				return nil
			},
		}

		uc := NewDeleteDepartmentUseCase(repo)
		err := uc.Execute(bg, tenantID, deptuc.DeleteDepartmentInput{ID: deptID})
		require.NoError(t, err)
	})

	t.Run("nil uuid returns ErrDepartmentNotFound", func(t *testing.T) {
		uc := NewDeleteDepartmentUseCase(&mockDepartmentRepo{})
		err := uc.Execute(bg, tenantID, deptuc.DeleteDepartmentInput{ID: uuid.Nil})
		require.ErrorIs(t, err, domainerrors.ErrDepartmentNotFound)
	})

	t.Run("department not found returns ErrDepartmentNotFound", func(t *testing.T) {
		repo := &mockDepartmentRepo{
			getByIDFn: func(ctx context.Context, tID, id uuid.UUID) (*entity.Department, error) {
				return nil, domainerrors.ErrDepartmentNotFound
			},
		}
		uc := NewDeleteDepartmentUseCase(repo)
		err := uc.Execute(bg, tenantID, deptuc.DeleteDepartmentInput{ID: deptID})
		require.ErrorIs(t, err, domainerrors.ErrDepartmentNotFound)
	})

	t.Run("department in use returns ErrDepartmentInUse", func(t *testing.T) {
		repo := &mockDepartmentRepo{
			getByIDFn: func(ctx context.Context, tID, id uuid.UUID) (*entity.Department, error) {
				return &entity.Department{ID: deptID, TenantID: tenantID}, nil
			},
			isReferencedFn: func(ctx context.Context, tID, id uuid.UUID) (bool, error) {
				return true, nil
			},
		}
		uc := NewDeleteDepartmentUseCase(repo)
		err := uc.Execute(bg, tenantID, deptuc.DeleteDepartmentInput{ID: deptID})
		require.ErrorIs(t, err, domainerrors.ErrDepartmentInUse)
	})

	t.Run("nil tenant returns unauthorized", func(t *testing.T) {
		uc := NewDeleteDepartmentUseCase(&mockDepartmentRepo{})
		err := uc.Execute(bg, uuid.Nil, deptuc.DeleteDepartmentInput{ID: deptID})
		require.ErrorIs(t, err, domainerrors.ErrUnauthorized)
	})

	t.Run("repo error on delete propagated", func(t *testing.T) {
		boom := errors.New("db error")
		repo := &mockDepartmentRepo{
			getByIDFn: func(ctx context.Context, tID, id uuid.UUID) (*entity.Department, error) {
				return &entity.Department{ID: deptID, TenantID: tenantID}, nil
			},
			isReferencedFn: func(ctx context.Context, tID, id uuid.UUID) (bool, error) {
				return false, nil
			},
			deleteFn: func(ctx context.Context, tID, id uuid.UUID) error {
				return boom
			},
		}
		uc := NewDeleteDepartmentUseCase(repo)
		err := uc.Execute(bg, tenantID, deptuc.DeleteDepartmentInput{ID: deptID})
		require.ErrorIs(t, err, boom)
	})
}
