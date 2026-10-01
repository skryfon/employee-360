package container

import (
	"gorm.io/gorm"

	"github.com/skryfon/employee360/backend/internal/delivery/http/handlers"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	"github.com/skryfon/employee360/backend/internal/infrastructure/persistence"
	deptimpl "github.com/skryfon/employee360/backend/internal/usecase/implementation/department"
	deptuc "github.com/skryfon/employee360/backend/internal/usecase/interface/department"
)

// DepartmentContainer encapsulates dependencies, repositories, usecases, and handlers for Departments.
type DepartmentContainer struct {
	Repo repository.DepartmentRepository

	CreateUseCase deptuc.CreateDepartmentUseCase
	GetUseCase    deptuc.GetDepartmentUseCase
	ListUseCase   deptuc.ListDepartmentsUseCase
	UpdateUseCase deptuc.UpdateDepartmentUseCase
	DeleteUseCase deptuc.DeleteDepartmentUseCase

	Handler *handlers.DepartmentHandler
}

// NewDepartmentContainer initializes and wires all department-related repository, usecases, and handler.
func NewDepartmentContainer(db *gorm.DB) (*DepartmentContainer, error) {
	repo := persistence.NewGormDepartmentRepository(db)

	createUC := deptimpl.NewCreateDepartmentUseCase(repo)
	getUC := deptimpl.NewGetDepartmentUseCase(repo)
	listUC := deptimpl.NewListDepartmentsUseCase(repo)
	updateUC := deptimpl.NewUpdateDepartmentUseCase(repo)
	deleteUC := deptimpl.NewDeleteDepartmentUseCase(repo)

	handler := handlers.NewDepartmentHandler(createUC, getUC, listUC, updateUC, deleteUC)

	return &DepartmentContainer{
		Repo:          repo,
		CreateUseCase: createUC,
		GetUseCase:    getUC,
		ListUseCase:   listUC,
		UpdateUseCase: updateUC,
		DeleteUseCase: deleteUC,
		Handler:       handler,
	}, nil
}
