package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// PasswordResetRepository defines the data access methods for password reset tokens.
type PasswordResetRepository interface {
	Create(ctx context.Context, token *entity.PasswordResetToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.PasswordResetToken, error)
	MarkAsUsed(ctx context.Context, id uuid.UUID) error
	DeleteExpiredTokens(ctx context.Context, before time.Time) error
	InvalidateAllForUser(ctx context.Context, userID uuid.UUID) error
}
