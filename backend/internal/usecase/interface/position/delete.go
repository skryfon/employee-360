package position

import (
	"context"

	"github.com/google/uuid"
)

// DeletePositionInput specifies the position to remove.
type DeletePositionInput struct {
	ID uuid.UUID
}

// DeletePositionUseCase defines the contract for deleting a position within the caller's tenant.
type DeletePositionUseCase interface {
	Execute(ctx context.Context, tenantID, actorID uuid.UUID, input DeletePositionInput) error
}
