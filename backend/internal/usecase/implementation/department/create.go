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
	"github.com/skryfon/employee360/backend/internal/usecase/implementation/ucshared"
	deptuc "github.com/skryfon/employee360/backend/internal/usecase/interface/department"
)

var (
	ErrDepartmentNameRequired = errors.New("department name is required")
	ErrDepartmentNameTooLong  = errors.New("department name must not exceed 100 characters")
	ErrDescriptionTooLong     = errors.New("department description must not exceed 500 characters")
)

type createDepartmentUseCase struct {
	repo       repository.DepartmentRepository
	auditRepo  repository.AuditRepository
	transactor ucshared.Transactor
}

// NewCreateDepartmentUseCase creates a new CreateDepartmentUseCase.
func NewCreateDepartmentUseCase(
	repo repository.DepartmentRepository,
	auditRepo repository.AuditRepository,
	transactor ucshared.Transactor,
) deptuc.CreateDepartmentUseCase {
	if transactor == nil {
		transactor = ucshared.NewNopTransactor()
	}
	return &createDepartmentUseCase{
		repo:       repo,
		auditRepo:  auditRepo,
		transactor: transactor,
	}
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

	now := time.Now().UTC()
	dept := &entity.Department{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := uc.transactor.WithinTransaction(c, func(txCtx context.Context) error {
		exists, err := uc.repo.ExistsByName(txCtx, tenantID, name)
		if err != nil {
			return err
		}
		if exists {
			return domainerrors.ErrDepartmentNameTaken
		}

		if err := uc.repo.Create(txCtx, tenantID, actorID, dept); err != nil {
			return err
		}

		return writeAudit(txCtx, uc.auditRepo, tenantID, actorID, dept.ID, auditActionCreate, map[string]any{
			"name": dept.Name,
		})
	})
	if err != nil {
		return nil, err
	}

	return dept, nil
}
