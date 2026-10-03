package position

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// CreatePositionInput holds validated data required to create a position.
type CreatePositionInput struct {
	Name        string
	Description string
	IsActive    *bool // nil defaults to true
}

// CreatePositionUseCase defines the contract for creating a position within the caller's tenant.
type CreatePositionUseCase interface {
	Execute(ctx context.Context, tenantID, actorID uuid.UUID, input CreatePositionInput) (*entity.Position, error)
}
