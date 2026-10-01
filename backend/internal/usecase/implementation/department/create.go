package department

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	deptuc "github.com/skryfon/employee360/backend/internal/usecase/interface/department"
)

var (
	ErrDepartmentNameRequired = errors.New("department name is required")
	ErrDepartmentNameTooLong  = errors.New("department name must not exceed 100 characters")
	ErrDescriptionTooLong     = errors.New("department description must not exceed 500 characters")
)

type createDepartmentUseCase struct {
	repo repository.DepartmentRepository
}

// NewCreateDepartmentUseCase creates a new CreateDepartmentUseCase.
func NewCreateDepartmentUseCase(repo repository.DepartmentRepository) deptuc.CreateDepartmentUseCase {
	return &createDepartmentUseCase{repo: repo}
}

func (uc *createDepartmentUseCase) Execute(c context.Context, tenantID, actorID uuid.UUID, input deptuc.CreateDepartmentInput) (*entity.Department, error) {
	if tenantID == uuid.Nil || actorID == uuid.Nil {
		return nil, domainerrors.ErrUnauthorized
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrDepartmentNameRequired
	}
	if len(name) > 100 {
		return nil, ErrDepartmentNameTooLong
	}

	description := strings.TrimSpace(input.Description)
	if len(description) > 500 {
		return nil, ErrDescriptionTooLong
	}

	exists, err := uc.repo.ExistsByName(c, tenantID, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domainerrors.ErrDepartmentNameTaken
	}

	now := time.Now().UTC()
	dept := &entity.Department{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.repo.Create(c, tenantID, actorID, dept); err != nil {
		return nil, err
	}

	return dept, nil
}
