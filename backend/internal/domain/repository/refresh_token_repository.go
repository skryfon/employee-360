package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// RefreshTokenRepository defines the data access methods for refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeFamily(ctx context.Context, family uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	DeleteExpiredTokens(ctx context.Context, before time.Time) error
}
