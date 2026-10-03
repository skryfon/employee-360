package position

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// GetPositionInput specifies the position to look up.
type GetPositionInput struct {
	ID uuid.UUID
}

// GetPositionUseCase defines the contract for retrieving a position by ID within the caller's tenant.
type GetPositionUseCase interface {
	Execute(ctx context.Context, tenantID uuid.UUID, input GetPositionInput) (*entity.Position, error)
}
