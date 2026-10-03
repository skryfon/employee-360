package position

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// ListPositionsInput holds pagination parameters for listing positions.
type ListPositionsInput struct {
	Page     int
	PageSize int
	IsActive *bool // nil returns all positions
}

// ListPositionsOutput contains paginated positions and pagination metadata.
type ListPositionsOutput struct {
	Positions []*entity.Position
	Total     int64
	Page      int
	PageSize  int
}

// ListPositionsUseCase defines the contract for listing positions within the caller's tenant.
type ListPositionsUseCase interface {
	Execute(ctx context.Context, tenantID uuid.UUID, input ListPositionsInput) (*ListPositionsOutput, error)
}
