package position

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	posuc "github.com/skryfon/employee360/backend/internal/usecase/interface/position"
)

func TestUpdatePositionUseCase_Execute(t *testing.T) {
	tenantID := uuid.New()
	actorID := uuid.New()
	posID := uuid.New()
	ctx := context.Background()

	basePos := func() *entity.Position {
		return &entity.Position{
			ID:          posID,
			TenantID:    tenantID,
			Name:        "Software Engineer",
			Description: "Old desc",
			IsActive:    true,
			CreatedAt:   time.Now().UTC().Add(-time.Hour),
			UpdatedAt:   time.Now().UTC().Add(-time.Hour),
		}
	}

	t.Run("success updates position and writes audit logs", func(t *testing.T) {
		existing := basePos()
		var updatedPos *entity.Position
		repo := &mockPositionRepo{
			getByIDFn: func(ctx context.Context, tID, id uuid.UUID) (*entity.Position, error) {
				assert.Equal(t, tenantID, tID)
				assert.Equal(t, posID, id)
				return existing, nil
			},
			existsByNameFn: func(ctx context.Context, tID uuid.UUID, name string) (bool, error) {
				assert.Equal(t, "Senior Engineer", name)
				return false, nil
			},
			updateFn: func(ctx context.Context, tID, aID uuid.UUID, p *entity.Position) error {
				assert.Equal(t, tenantID, tID)
				assert.Equal(t, actorID, aID)
				updatedPos = p
				return nil
			},
		}
		auditRepo := &mockAuditRepo{}
		uc := NewUpdatePositionUseCase(repo, auditRepo, nil)

		res, err := uc.Execute(ctx, tenantID, actorID, posuc.UpdatePositionInput{
			ID:          posID,
			Name:        "Senior Engineer",
			Description: "New desc",
		})
		require.NoError(t, err)
		assert.Equal(t, "Senior Engineer", res.Name)
		assert.Equal(t, "New desc", res.Description)
		assert.True(t, res.IsActive)
		assert.Equal(t, updatedPos, res)

		require.Len(t, auditRepo.logs, 1)
		assert.Equal(t, auditActionUpdate, auditRepo.logs[0].Action)
	})

	t.Run("deactivate triggers both update and deactivate audit logs", func(t *testing.T) {
		existing := basePos()
		repo := &mockPositionRepo{
			getByIDFn: func(ctx context.Context, tID, id uuid.UUID) (*entity.Position, error) {
				return existing, nil
			},
			updateFn: func(ctx context.Context, tID, aID uuid.UUID, p *entity.Position) error {
				return nil
			},
		}
		auditRepo := &mockAuditRepo{}
		uc := NewUpdatePositionUseCase(repo, auditRepo, nil)

		fl := false
		res, err := uc.Execute(ctx, tenantID, actorID, posuc.UpdatePositionInput{
			ID:       posID,
			Name:     "Software Engineer", // unchanged name
			IsActive: &fl,
		})
		require.NoError(t, err)
		assert.False(t, res.IsActive)

		require.Len(t, auditRepo.logs, 2)
		assert.Equal(t, auditActionUpdate, auditRepo.logs[0].Action)
		assert.Equal(t, auditActionDeactivate, auditRepo.logs[1].Action)
	})

	t.Run("activate triggers both update and activate audit logs", func(t *testing.T) {
		existing := basePos()
		existing.IsActive = false
		repo := &mockPositionRepo{
			getByIDFn: func(ctx context.Context, tID, id uuid.UUID) (*entity.Position, error) {
				return existing, nil
			},
			updateFn: func(ctx context.Context, tID, aID uuid.UUID, p *entity.Position) error {
				return nil
			},
		}
		auditRepo := &mockAuditRepo{}
		uc := NewUpdatePositionUseCase(repo, auditRepo, nil)

		tr := true
		res, err := uc.Execute(ctx, tenantID, actorID, posuc.UpdatePositionInput{
			ID:       posID,
			Name:     "Software Engineer",
			IsActive: &tr,
		})
		require.NoError(t, err)
		assert.True(t, res.IsActive)

		require.Len(t, auditRepo.logs, 2)
		assert.Equal(t, auditActionUpdate, auditRepo.logs[0].Action)
		assert.Equal(t, auditActionActivate, auditRepo.logs[1].Action)
	})

	t.Run("changing name to duplicate returns ErrPositionNameTaken", func(t *testing.T) {
		existing := basePos()
		repo := &mockPositionRepo{
			getByIDFn: func(ctx context.Context, tID, id uuid.UUID) (*entity.Position, error) {
				return existing, nil
			},
			existsByNameFn: func(ctx context.Context, tID uuid.UUID, name string) (bool, error) {
				return true, nil
			},
		}
		uc := NewUpdatePositionUseCase(repo, &mockAuditRepo{}, nil)

		_, err := uc.Execute(ctx, tenantID, actorID, posuc.UpdatePositionInput{
			ID:   posID,
			Name: "Existing Different Position",
		})
		require.ErrorIs(t, err, domainerrors.ErrPositionNameTaken)
	})

	t.Run("unauthorized on nil tenantID or actorID", func(t *testing.T) {
		uc := NewUpdatePositionUseCase(&mockPositionRepo{}, &mockAuditRepo{}, nil)

		_, err := uc.Execute(ctx, uuid.Nil, actorID, posuc.UpdatePositionInput{ID: posID, Name: "Test"})
		require.ErrorIs(t, err, domainerrors.ErrUnauthorized)

		_, err = uc.Execute(ctx, tenantID, uuid.Nil, posuc.UpdatePositionInput{ID: posID, Name: "Test"})
		require.ErrorIs(t, err, domainerrors.ErrUnauthorized)
	})

	t.Run("nil uuid returns ErrPositionNotFound", func(t *testing.T) {
		uc := NewUpdatePositionUseCase(&mockPositionRepo{}, &mockAuditRepo{}, nil)

		_, err := uc.Execute(ctx, tenantID, actorID, posuc.UpdatePositionInput{ID: uuid.Nil, Name: "Test"})
		require.ErrorIs(t, err, domainerrors.ErrPositionNotFound)
	})

	t.Run("not found in repo returns ErrPositionNotFound", func(t *testing.T) {
		repo := &mockPositionRepo{
			getByIDFn: func(ctx context.Context, tID, id uuid.UUID) (*entity.Position, error) {
				return nil, domainerrors.ErrPositionNotFound
			},
		}
		uc := NewUpdatePositionUseCase(repo, &mockAuditRepo{}, nil)

		_, err := uc.Execute(ctx, tenantID, actorID, posuc.UpdatePositionInput{ID: posID, Name: "Test"})
		require.ErrorIs(t, err, domainerrors.ErrPositionNotFound)
	})

	t.Run("validation: name required", func(t *testing.T) {
		uc := NewUpdatePositionUseCase(&mockPositionRepo{}, &mockAuditRepo{}, nil)

		_, err := uc.Execute(ctx, tenantID, actorID, posuc.UpdatePositionInput{ID: posID, Name: ""})
		require.ErrorIs(t, err, ErrPositionNameRequired)
	})

	t.Run("validation: name too long", func(t *testing.T) {
		uc := NewUpdatePositionUseCase(&mockPositionRepo{}, &mockAuditRepo{}, nil)

		_, err := uc.Execute(ctx, tenantID, actorID, posuc.UpdatePositionInput{ID: posID, Name: strings.Repeat("x", 101)})
		require.ErrorIs(t, err, ErrPositionNameTooLong)
	})

	t.Run("validation: description too long", func(t *testing.T) {
		uc := NewUpdatePositionUseCase(&mockPositionRepo{}, &mockAuditRepo{}, nil)

		_, err := uc.Execute(ctx, tenantID, actorID, posuc.UpdatePositionInput{
			ID:          posID,
			Name:        "Valid",
			Description: strings.Repeat("x", 501),
		})
		require.ErrorIs(t, err, ErrDescriptionTooLong)
	})

	t.Run("repo error propagates", func(t *testing.T) {
		existing := basePos()
		expectedErr := errors.New("update failed")
		repo := &mockPositionRepo{
			getByIDFn: func(ctx context.Context, tID, id uuid.UUID) (*entity.Position, error) {
				return existing, nil
			},
			updateFn: func(ctx context.Context, tID, aID uuid.UUID, p *entity.Position) error {
				return expectedErr
			},
		}
		uc := NewUpdatePositionUseCase(repo, &mockAuditRepo{}, nil)

		_, err := uc.Execute(ctx, tenantID, actorID, posuc.UpdatePositionInput{ID: posID, Name: "Software Engineer"})
		require.ErrorIs(t, err, expectedErr)
	})
}
