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

func TestListDepartmentsUseCase(t *testing.T) {
	tenantID := uuid.New()
	bg := context.Background()

	t.Run("default pagination", func(t *testing.T) {
		repo := &mockDepartmentRepo{
			listFn: func(ctx context.Context, tID uuid.UUID, limit, offset int) ([]*entity.Department, int64, error) {
				assert.Equal(t, tenantID, tID)
				assert.Equal(t, 20, limit)
				assert.Equal(t, 0, offset)
				return []*entity.Department{
					{ID: uuid.New(), TenantID: tenantID, Name: "Engineering"},
				}, 1, nil
			},
		}

		uc := NewListDepartmentsUseCase(repo)
		out, err := uc.Execute(bg, tenantID, deptuc.ListDepartmentsInput{})
		require.NoError(t, err)
		assert.Equal(t, int64(1), out.Total)
		assert.Equal(t, 1, out.Page)
		assert.Equal(t, 20, out.PageSize)
		assert.Len(t, out.Departments, 1)
	})

	t.Run("custom pagination", func(t *testing.T) {
		repo := &mockDepartmentRepo{
			listFn: func(ctx context.Context, tID uuid.UUID, limit, offset int) ([]*entity.Department, int64, error) {
				assert.Equal(t, tenantID, tID)
				assert.Equal(t, 10, limit)
				assert.Equal(t, 20, offset) // page 3 with size 10 -> offset 20
				return []*entity.Department{}, 25, nil
			},
		}

		uc := NewListDepartmentsUseCase(repo)
		out, err := uc.Execute(bg, tenantID, deptuc.ListDepartmentsInput{
			Page:     3,
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(25), out.Total)
		assert.Equal(t, 3, out.Page)
		assert.Equal(t, 10, out.PageSize)
	})

	t.Run("nil tenant returns unauthorized", func(t *testing.T) {
		uc := NewListDepartmentsUseCase(&mockDepartmentRepo{})
		_, err := uc.Execute(bg, uuid.Nil, deptuc.ListDepartmentsInput{})
		require.ErrorIs(t, err, domainerrors.ErrUnauthorized)
	})

	t.Run("repo error propagated", func(t *testing.T) {
		boom := errors.New("database error")
		repo := &mockDepartmentRepo{
			listFn: func(ctx context.Context, tID uuid.UUID, limit, offset int) ([]*entity.Department, int64, error) {
				return nil, 0, boom
			},
		}

		uc := NewListDepartmentsUseCase(repo)
		_, err := uc.Execute(bg, tenantID, deptuc.ListDepartmentsInput{})
		require.ErrorIs(t, err, boom)
	})
}
