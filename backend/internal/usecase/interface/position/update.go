package position

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// UpdatePositionInput specifies the position to modify and its new attributes.
type UpdatePositionInput struct {
	ID          uuid.UUID
	Name        string
	Description string
	IsActive    *bool // nil keeps the current value
}

// UpdatePositionUseCase defines the contract for updating a position within the caller's tenant.
type UpdatePositionUseCase interface {
	Execute(ctx context.Context, tenantID, actorID uuid.UUID, input UpdatePositionInput) (*entity.Position, error)
}
