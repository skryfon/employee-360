package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// UserInvitationRepository defines the data access methods for user invitations.
//
// Like UserRepository, tenant-scoped methods take an explicit tenantID that
// the calling usecase obtained from the verified request context; it is never
// client-supplied. GetByTokenHash is the sole unscoped lookup: the accept flow
// is unauthenticated and resolves the tenant from the token itself.
type UserInvitationRepository interface {
	Create(ctx context.Context, invitation *entity.UserInvitation) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.UserInvitation, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.UserInvitation, error)
	// UpdateToken replaces the token hash and expiry (resend). It is a
	// conditional update (accepted_at IS NULL AND revoked_at IS NULL) and MUST
	// return domainerrors.ErrInvitationNotPending when 0 rows were updated.
	UpdateToken(ctx context.Context, tenantID, id, actorID uuid.UUID, tokenHash string, expiresAt time.Time) error
	// MarkAccepted is a conditional update (accepted_at IS NULL AND
	// revoked_at IS NULL) and MUST return domainerrors.ErrInvitationNotPending
	// when 0 rows were updated (the accept usecase maps it to ErrInvalidToken).
	MarkAccepted(ctx context.Context, tenantID, id uuid.UUID, at time.Time) error
	// MarkRevoked is a conditional update (accepted_at IS NULL AND
	// revoked_at IS NULL) and MUST return domainerrors.ErrInvitationNotPending
	// when 0 rows were updated.
	MarkRevoked(ctx context.Context, tenantID, id, actorID uuid.UUID, at time.Time) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.UserInvitation, int64, error)
}
