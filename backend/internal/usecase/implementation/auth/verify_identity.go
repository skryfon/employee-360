package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	authtypes "github.com/skryfon/employee360/backend/internal/types/auth"
	authusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/auth"
)

// VerifyIdentityUseCaseImpl implements authusecase.VerifyIdentityUseCase.
type VerifyIdentityUseCaseImpl struct {
	tenantReader repository.TenantReader
	userRepo     repository.UserRepository
	userRoleRepo repository.UserRoleRepository
}

var _ authusecase.VerifyIdentityUseCase = (*VerifyIdentityUseCaseImpl)(nil)

// NewVerifyIdentityUseCase constructs a new VerifyIdentityUseCaseImpl.
func NewVerifyIdentityUseCase(
	tenantReader repository.TenantReader,
	userRepo repository.UserRepository,
	userRoleRepo repository.UserRoleRepository,
) *VerifyIdentityUseCaseImpl {
	return &VerifyIdentityUseCaseImpl{tenantReader: tenantReader, userRepo: userRepo, userRoleRepo: userRoleRepo}
}

// Execute verifies the tenant and user and loads the user's current roles.
func (u *VerifyIdentityUseCaseImpl) Execute(ctx context.Context, tenantID, userID uuid.UUID) (*authtypes.VerifiedIdentity, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return nil, domainerrors.ErrUnauthorized
	}

	tenant, err := u.tenantReader.GetByID(ctx, tenantID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrTenantNotFound) {
			return nil, domainerrors.ErrUnauthorized
		}
		return nil, err
	}
	if tenant == nil || !tenant.IsActive || tenant.DeletedAt != nil {
		return nil, domainerrors.ErrUnauthorized
	}

	// Scoped by tenantID: a user id from another tenant is simply not found.
	user, err := u.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return nil, domainerrors.ErrUnauthorized
		}
		return nil, err
	}
	if user == nil || !user.IsActive || user.DeletedAt != nil || user.TenantID != tenantID {
		return nil, domainerrors.ErrUnauthorized
	}

	roles, err := u.userRoleRepo.GetRolesByUserID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(roles))
	for _, r := range roles {
		if r != nil {
			names = append(names, r.Name)
		}
	}

	return &authtypes.VerifiedIdentity{TenantID: tenantID, UserID: userID, Roles: names}, nil
}
