package invitation

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	invtypes "github.com/skryfon/employee360/backend/internal/types/invitation"
)

// InviteUserUseCase invites a user into the caller's tenant (from context only).
type InviteUserUseCase interface {
	Execute(ctx context.Context, req invtypes.InviteUserRequest) (*entity.UserInvitation, error)
}

// AcceptInvitationUseCase consumes an invitation token; every invitee must set a password.
type AcceptInvitationUseCase interface {
	Execute(ctx context.Context, req invtypes.AcceptInvitationRequest) error
}

// ResendInvitationUseCase reissues the token of a pending invitation.
type ResendInvitationUseCase interface {
	Execute(ctx context.Context, id uuid.UUID) (*entity.UserInvitation, error)
}

// RevokeInvitationUseCase cancels a pending invitation.
type RevokeInvitationUseCase interface {
	Execute(ctx context.Context, id uuid.UUID) error
}

// ListInvitationsUseCase lists the caller's tenant invitations.
type ListInvitationsUseCase interface {
	Execute(ctx context.Context, limit, offset int) ([]*entity.UserInvitation, int64, error)
}

// ValidateInvitationUseCase verifies that an invitation token exists, is pending, not expired, and ready for acceptance.
type ValidateInvitationUseCase interface {
	Execute(ctx context.Context, token string) (*invtypes.ValidateInvitationResponse, error)
}
