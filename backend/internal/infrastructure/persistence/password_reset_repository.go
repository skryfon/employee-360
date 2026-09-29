package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
)

// gormPasswordResetRepository is a GORM-backed adapter implementing
// repository.PasswordResetRepository.
type gormPasswordResetRepository struct {
	db *gorm.DB
}

// NewGormPasswordResetRepository constructs a GORM-backed PasswordResetRepository.
func NewGormPasswordResetRepository(db *gorm.DB) repository.PasswordResetRepository {
	return &gormPasswordResetRepository{db: db}
}

var _ repository.PasswordResetRepository = (*gormPasswordResetRepository)(nil)

// Create persists a new password reset token record.
func (r *gormPasswordResetRepository) Create(c context.Context, token *entity.PasswordResetToken) error {
	return r.db.WithContext(c).Create(token).Error
}

// GetByTokenHash looks up a password reset token by its hash alone. The hash
// is a unique, high-entropy value (uq_password_reset_tokens_token_hash), so
// no further scoping is required or possible before the token has been
// resolved to a user/tenant -- this is the existing pattern already used by
// ResetPasswordUseCase.
func (r *gormPasswordResetRepository) GetByTokenHash(c context.Context, tokenHash string) (*entity.PasswordResetToken, error) {
	var t entity.PasswordResetToken
	if err := r.db.WithContext(c).
		Where("token_hash = ?", tokenHash).
		First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

// MarkAsUsed marks a password reset token as consumed.
func (r *gormPasswordResetRepository) MarkAsUsed(c context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(c).
		Model(&entity.PasswordResetToken{}).
		Where("id = ?", id).
		Update("used_at", now).Error
}

// DeleteExpiredTokens hard-deletes password reset tokens that expired before
// the given time (housekeeping).
func (r *gormPasswordResetRepository) DeleteExpiredTokens(c context.Context, before time.Time) error {
	return r.db.WithContext(c).
		Where("expires_at < ?", before).
		Delete(&entity.PasswordResetToken{}).Error
}

// InvalidateAllForUser marks every outstanding password reset token for a
// user as used, so at most one reset token is ever valid at a time.
func (r *gormPasswordResetRepository) InvalidateAllForUser(c context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(c).
		Model(&entity.PasswordResetToken{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Update("used_at", now).Error
}
