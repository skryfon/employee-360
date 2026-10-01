package invitation

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/ctx"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/event"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	"github.com/skryfon/employee360/backend/internal/domain/service"
	invtypes "github.com/skryfon/employee360/backend/internal/types/invitation"
	"github.com/skryfon/employee360/backend/internal/usecase/implementation/ucshared"
	invusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/invitation"
)

// InviteUserUseCaseImpl implements invusecase.InviteUserUseCase.
type InviteUserUseCaseImpl struct {
	userRepo       repository.UserRepository
	userRoleRepo   repository.UserRoleRepository
	roleRepo       repository.RoleRepository
	invitationRepo repository.UserInvitationRepository
	orgRefRepo     repository.OrgReferenceRepository
	auditRepo      repository.AuditRepository
	hashService    service.HashService
	eventPublisher service.EventPublisher
	transactor     ucshared.Transactor
	expiry         time.Duration
	appURLs        AppURLs
}

var _ invusecase.InviteUserUseCase = (*InviteUserUseCaseImpl)(nil)

// NewInviteUserUseCase constructs an InviteUserUseCaseImpl.
func NewInviteUserUseCase(
	userRepo repository.UserRepository,
	userRoleRepo repository.UserRoleRepository,
	roleRepo repository.RoleRepository,
	invitationRepo repository.UserInvitationRepository,
	orgRefRepo repository.OrgReferenceRepository,
	auditRepo repository.AuditRepository,
	hashService service.HashService,
	eventPublisher service.EventPublisher,
	transactor ucshared.Transactor,
	appURLs AppURLs,
) *InviteUserUseCaseImpl {
	return &InviteUserUseCaseImpl{
		userRepo: userRepo, userRoleRepo: userRoleRepo, roleRepo: roleRepo,
		invitationRepo: invitationRepo, orgRefRepo: orgRefRepo, auditRepo: auditRepo, hashService: hashService,
		eventPublisher: eventPublisher, transactor: transactor,
		expiry:  defaultInvitationExpiry,
		appURLs: appURLs,
	}
}

// Execute creates a pending user + invitation and publishes UserInvited in one transaction.
func (u *InviteUserUseCaseImpl) Execute(c context.Context, req invtypes.InviteUserRequest) (*entity.UserInvitation, error) {
	if err := requireAdmin(c); err != nil {
		return nil, err
	}
	tenantID, err := tenantFromContext(c)
	if err != nil {
		return nil, err
	}
	inviterRaw, _ := ctx.UserIDFromContext(c)
	inviterID, err := uuid.Parse(inviterRaw)
	if err != nil {
		return nil, domainerrors.ErrUnauthorized
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if at := strings.Index(email, "@"); at <= 0 || at == len(email)-1 {
		return nil, domainerrors.ErrInvalidEmail
	}

	// The role must belong to the caller's tenant; super_admin is never invitable.
	role, err := u.roleRepo.GetByID(c, req.RoleID)
	if err != nil || role == nil || role.TenantID != tenantID {
		return nil, domainerrors.ErrRoleNotFound
	}
	if role.Name == roleSuperAdmin {
		return nil, domainerrors.ErrInvalidRole
	}

	// Department/position must belong to the caller's tenant (FKs only check existence).
	if req.DepartmentID != nil {
		ok, err := u.orgRefRepo.DepartmentExists(c, tenantID, *req.DepartmentID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, domainerrors.ErrDepartmentNotFound
		}
	}
	if req.PositionID != nil {
		ok, err := u.orgRefRepo.PositionExists(c, tenantID, *req.PositionID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, domainerrors.ErrPositionNotFound
		}
	}

	existing, err := u.userRepo.GetByTenantAndEmail(c, tenantID, email)
	if err != nil && !errors.Is(err, domainerrors.ErrNotFound) && !errors.Is(err, domainerrors.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domainerrors.ErrEmailAlreadyExists
	}

	plainToken, err := newPlainToken()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	inv := &entity.UserInvitation{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Email:        email,
		RoleID:       role.ID,
		DepartmentID: req.DepartmentID,
		PositionID:   req.PositionID,
		InvitedBy:    inviterID,
		TokenHash:    u.hashService.HashToken(plainToken),
		ExpiresAt:    now.Add(u.expiry),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	user := &entity.User{
		ID:           uuid.New(),
		TenantID:     tenantID,
		DepartmentID: req.DepartmentID,
		PositionID:   req.PositionID,
		FirstName:    strings.TrimSpace(req.FirstName),
		LastName:     strings.TrimSpace(req.LastName),
		Email:        email,
		IsActive:     false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	evt := event.Event{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EventType:     event.EventTypeUserInvited,
		AggregateType: "user_invitation",
		AggregateID:   inv.ID.String(),
		Payload: event.UserInvitedPayload{
			InvitationID: inv.ID, TenantID: tenantID, Email: email,
			RoleName: role.Name, PlainToken: plainToken,
			InviteURL: u.appURLs.acceptLink(role.Name, plainToken),
			InvitedBy: inviterID, ExpiresAt: inv.ExpiresAt,
		},
		OccurredAt: now,
	}

	if err := u.transactor.WithinTransaction(c, func(txCtx context.Context) error {
		if err := u.userRepo.Create(txCtx, user); err != nil {
			return err
		}
		if err := u.userRoleRepo.AssignRole(txCtx, &entity.UserRole{
			ID: uuid.New(), TenantID: tenantID, UserID: user.ID, RoleID: role.ID,
			CreatedAt: now, UpdatedAt: now,
		}); err != nil {
			return err
		}
		if err := u.invitationRepo.Create(txCtx, inv); err != nil {
			return err
		}
		if err := writeAudit(txCtx, u.auditRepo, tenantID, inviterID, inv.ID, auditActionInvite,
			map[string]any{"email": email, "role_id": role.ID, "user_id": user.ID}); err != nil {
			return err
		}
		return u.eventPublisher.Publish(txCtx, evt)
	}); err != nil {
		return nil, err
	}
	return inv, nil
}
