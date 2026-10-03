package position

import (
	"context"

	"github.com/google/uuid"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	postypes "github.com/skryfon/employee360/backend/internal/types/position"
	posuc "github.com/skryfon/employee360/backend/internal/usecase/interface/position"
)

type listPositionsUseCase struct {
	repo repository.PositionRepository
}

// NewListPositionsUseCase creates a new ListPositionsUseCase.
func NewListPositionsUseCase(repo repository.PositionRepository) posuc.ListPositionsUseCase {
	return &listPositionsUseCase{repo: repo}
}

func (uc *listPositionsUseCase) Execute(c context.Context, tenantID uuid.UUID, input postypes.ListPositionsQuery) (*postypes.ListPositionsResult, error) {
	if tenantID == uuid.Nil {
		return nil, domainerrors.ErrUnauthorized
	}

	page := input.Page
	if page < 1 {
		page = 1
	}

	pageSize := input.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	items, total, err := uc.repo.List(c, tenantID, input.IsActive, limit, offset)
	if err != nil {
		return nil, err
	}

	return &postypes.ListPositionsResult{
		Positions: items,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}
