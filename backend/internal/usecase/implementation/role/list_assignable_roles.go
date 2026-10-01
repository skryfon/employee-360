// Package role implements the role usecases.
package role

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/ctx"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	roleusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/role"
)

// ListAssignableRolesUseCaseImpl implements roleusecase.ListAssignableRolesUseCase.
type ListAssignableRolesUseCaseImpl struct {
	roleRepo repository.RoleRepository
}

var _ roleusecase.ListAssignableRolesUseCase = (*ListAssignableRolesUseCaseImpl)(nil)

// NewListAssignableRolesUseCase constructs a ListAssignableRolesUseCaseImpl.
func NewListAssignableRolesUseCase(roleRepo repository.RoleRepository) *ListAssignableRolesUseCaseImpl {
	return &ListAssignableRolesUseCaseImpl{roleRepo: roleRepo}
}

// Execute returns the caller's tenant roles excluding super_admin, which the
// invite flow rejects. Tenant comes from context only.
func (u *ListAssignableRolesUseCaseImpl) Execute(c context.Context) ([]*entity.Role, error) {
	roles, _ := ctx.RolesFromContext(c)
	isAdmin := false
	for _, r := range roles {
		if r == entity.RoleAdmin || r == entity.RoleSuperAdmin {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		return nil, domainerrors.ErrForbidden
	}
	raw, ok := ctx.TenantIDFromContext(c)
	tenantID, err := uuid.Parse(raw)
	if !ok || err != nil || tenantID == uuid.Nil {
		return nil, domainerrors.ErrUnauthorized
	}

	all, err := u.roleRepo.List(c)
	if err != nil {
		return nil, err
	}
	out := make([]*entity.Role, 0, len(all))
	for _, r := range all {
		// Defence in depth on top of the repository's tenant scoping.
		if r == nil || r.TenantID != tenantID || r.Name == entity.RoleSuperAdmin {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}
