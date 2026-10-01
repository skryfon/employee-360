package department

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	deptuc "github.com/skryfon/employee360/backend/internal/usecase/interface/department"
)

type mockDepartmentRepo struct {
	createFn       func(ctx context.Context, tenantID, actorID uuid.UUID, department *entity.Department) error
	getByIDFn      func(ctx context.Context, tenantID, id uuid.UUID) (*entity.Department, error)
	listFn         func(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.Department, int64, error)
	updateFn       func(ctx context.Context, tenantID, actorID uuid.UUID, department *entity.Department) error
	deleteFn       func(ctx context.Context, tenantID, id uuid.UUID) error
	existsByNameFn func(ctx context.Context, tenantID uuid.UUID, name string) (bool, error)
	isReferencedFn func(ctx context.Context, tenantID, id uuid.UUID) (bool, error)
}

func (m *mockDepartmentRepo) Create(ctx context.Context, tenantID, actorID uuid.UUID, d *entity.Department) error {
	if m.createFn != nil {
		return m.createFn(ctx, tenantID, actorID, d)
	}
	return nil
}

func (m *mockDepartmentRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.Department, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, tenantID, id)
	}
	return nil, nil
}

func (m *mockDepartmentRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.Department, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, tenantID, limit, offset)
	}
	return nil, 0, nil
}

func (m *mockDepartmentRepo) Update(ctx context.Context, tenantID, actorID uuid.UUID, d *entity.Department) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, tenantID, actorID, d)
	}
	return nil
}

func (m *mockDepartmentRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, tenantID, id)
	}
	return nil
}

func (m *mockDepartmentRepo) ExistsByName(ctx context.Context, tenantID uuid.UUID, name string) (bool, error) {
	if m.existsByNameFn != nil {
		return m.existsByNameFn(ctx, tenantID, name)
	}
	return false, nil
}

func (m *mockDepartmentRepo) IsReferenced(ctx context.Context, tenantID, id uuid.UUID) (bool, error) {
	if m.isReferencedFn != nil {
		return m.isReferencedFn(ctx, tenantID, id)
	}
	return false, nil
}

func TestCreateDepartmentUseCase(t *testing.T) {
	tenantID := uuid.New()
	actorID := uuid.New()
	bg := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		repo := &mockDepartmentRepo{
			existsByNameFn: func(ctx context.Context, tID uuid.UUID, name string) (bool, error) {
				assert.Equal(t, tenantID, tID)
				return false, nil
			},
			createFn: func(ctx context.Context, tID, aID uuid.UUID, d *entity.Department) error {
				assert.Equal(t, tenantID, tID)
				assert.Equal(t, actorID, aID)
				assert.Equal(t, "Engineering", d.Name)
				assert.Equal(t, "Dev team", d.Description)
				assert.Equal(t, tenantID, d.TenantID)
				return nil
			},
		}

		uc := NewCreateDepartmentUseCase(repo)
		dept, err := uc.Execute(bg, tenantID, actorID, deptuc.CreateDepartmentInput{
			Name:        "  Engineering  ",
			Description: "  Dev team  ",
		})
		require.NoError(t, err)
		assert.Equal(t, "Engineering", dept.Name)
		assert.Equal(t, "Dev team", dept.Description)
		assert.Equal(t, tenantID, dept.TenantID)
	})

	t.Run("empty name returns validation error", func(t *testing.T) {
		uc := NewCreateDepartmentUseCase(&mockDepartmentRepo{})
		_, err := uc.Execute(bg, tenantID, actorID, deptuc.CreateDepartmentInput{Name: "   "})
		require.ErrorIs(t, err, ErrDepartmentNameRequired)
	})

	t.Run("name exceeds 100 characters", func(t *testing.T) {
		uc := NewCreateDepartmentUseCase(&mockDepartmentRepo{})
		_, err := uc.Execute(bg, tenantID, actorID, deptuc.CreateDepartmentInput{Name: strings.Repeat("a", 101)})
		require.ErrorIs(t, err, ErrDepartmentNameTooLong)
	})

	t.Run("description exceeds 500 characters", func(t *testing.T) {
		uc := NewCreateDepartmentUseCase(&mockDepartmentRepo{})
		_, err := uc.Execute(bg, tenantID, actorID, deptuc.CreateDepartmentInput{
			Name:        "Valid Name",
			Description: strings.Repeat("b", 501),
		})
		require.ErrorIs(t, err, ErrDescriptionTooLong)
	})

	t.Run("name taken returns conflict", func(t *testing.T) {
		repo := &mockDepartmentRepo{
			existsByNameFn: func(ctx context.Context, tID uuid.UUID, name string) (bool, error) {
				return true, nil
			},
		}
		uc := NewCreateDepartmentUseCase(repo)
		_, err := uc.Execute(bg, tenantID, actorID, deptuc.CreateDepartmentInput{Name: "Existing"})
		require.ErrorIs(t, err, domainerrors.ErrDepartmentNameTaken)
	})

	t.Run("nil tenant or actor returns unauthorized", func(t *testing.T) {
		uc := NewCreateDepartmentUseCase(&mockDepartmentRepo{})
		_, err := uc.Execute(bg, uuid.Nil, actorID, deptuc.CreateDepartmentInput{Name: "Engineering"})
		require.ErrorIs(t, err, domainerrors.ErrUnauthorized)

		_, err = uc.Execute(bg, tenantID, uuid.Nil, deptuc.CreateDepartmentInput{Name: "Engineering"})
		require.ErrorIs(t, err, domainerrors.ErrUnauthorized)
	})

	t.Run("repo error propagated", func(t *testing.T) {
		boom := errors.New("db error")
		repo := &mockDepartmentRepo{
			existsByNameFn: func(ctx context.Context, tID uuid.UUID, name string) (bool, error) {
				return false, nil
			},
			createFn: func(ctx context.Context, tID, aID uuid.UUID, d *entity.Department) error {
				return boom
			},
		}
		uc := NewCreateDepartmentUseCase(repo)
		_, err := uc.Execute(bg, tenantID, actorID, deptuc.CreateDepartmentInput{Name: "Engineering"})
		require.ErrorIs(t, err, boom)
	})
}
