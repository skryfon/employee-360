package invitation

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/event"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	"github.com/skryfon/employee360/backend/internal/domain/service"
	"github.com/skryfon/employee360/backend/internal/usecase/implementation/ucshared"
	invusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/invitation"
)

// ResendInvitationUseCaseImpl implements invusecase.ResendInvitationUseCase.
type ResendInvitationUseCaseImpl struct {
	roleRepo       repository.RoleRepository
	invitationRepo repository.UserInvitationRepository
	hashService    service.HashService
	eventPublisher service.EventPublisher
	transactor     ucshared.Transactor
	expiry         time.Duration
	frontendURL    string
}

var _ invusecase.ResendInvitationUseCase = (*ResendInvitationUseCaseImpl)(nil)

// NewResendInvitationUseCase constructs a ResendInvitationUseCaseImpl.
func NewResendInvitationUseCase(
	roleRepo repository.RoleRepository,
	invitationRepo repository.UserInvitationRepository,
	hashService service.HashService,
	eventPublisher service.EventPublisher,
	transactor ucshared.Transactor,
	frontendBaseURL string,
) *ResendInvitationUseCaseImpl {
	return &ResendInvitationUseCaseImpl{
		roleRepo: roleRepo, invitationRepo: invitationRepo, hashService: hashService,
		eventPublisher: eventPublisher, transactor: transactor,
		expiry: defaultInvitationExpiry, frontendURL: strings.TrimRight(frontendBaseURL, "/"),
	}
}

// Execute reissues the token and publishes InvitationResent, only while pending.
func (u *ResendInvitationUseCaseImpl) Execute(c context.Context, id uuid.UUID) (*entity.UserInvitation, error) {
	if err := requireAdmin(c); err != nil {
		return nil, err
	}
	tenantID, err := tenantFromContext(c)
	if err != nil {
		return nil, err
	}
	inv, err := u.invitationRepo.GetByID(c, tenantID, id)
	if err != nil || inv == nil {
		return nil, domainerrors.ErrInvitationNotFound
	}
	if !inv.IsPending() {
		return nil, domainerrors.ErrInvitationNotPending
	}
	role, err := u.roleRepo.GetByID(c, inv.RoleID)
	if err != nil || role == nil {
		return nil, domainerrors.ErrRoleNotFound
	}

	plainToken, err := newPlainToken()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	inv.TokenHash = u.hashService.HashToken(plainToken)
	inv.ExpiresAt = now.Add(u.expiry)
	inv.UpdatedAt = now

	evt := event.Event{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EventType:     event.EventTypeInvitationResent,
		AggregateType: "user_invitation",
		AggregateID:   inv.ID.String(),
		Payload: event.InvitationResentPayload{
			InvitationID: inv.ID, TenantID: tenantID, Email: inv.Email,
			RoleName: role.Name, PlainToken: plainToken,
			InviteURL: u.frontendURL + "/accept-invitation?token=" + plainToken,
			ExpiresAt: inv.ExpiresAt,
		},
		OccurredAt: now,
	}

	if err := u.transactor.WithinTransaction(c, func(txCtx context.Context) error {
		if err := u.invitationRepo.UpdateToken(txCtx, tenantID, inv.ID, inv.TokenHash, inv.ExpiresAt); err != nil {
			return err
		}
		return u.eventPublisher.Publish(txCtx, evt)
	}); err != nil {
		return nil, err
	}
	return inv, nil
}
