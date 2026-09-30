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
	"github.com/skryfon/employee360/backend/internal/infrastructure/database"
)

// gormUserInvitationRepository is a GORM-backed adapter implementing
// repository.UserInvitationRepository. Like the sibling repositories, every
// tenant-scoped method takes an explicit tenantID supplied by the usecase.
type gormUserInvitationRepository struct {
	db *gorm.DB
}

// NewGormUserInvitationRepository constructs a GORM-backed UserInvitationRepository.
func NewGormUserInvitationRepository(db *gorm.DB) repository.UserInvitationRepository {
	return &gormUserInvitationRepository{db: db}
}

var _ repository.UserInvitationRepository = (*gormUserInvitationRepository)(nil)

// pendingWhere restricts an update to a still-pending invitation in a tenant.
const pendingWhere = "id = ? AND tenant_id = ? AND accepted_at IS NULL AND revoked_at IS NULL"

func (r *gormUserInvitationRepository) Create(c context.Context, inv *entity.UserInvitation) error {
	return database.DBFromContext(c, r.db).Create(inv).Error
}

func (r *gormUserInvitationRepository) GetByID(c context.Context, tenantID, id uuid.UUID) (*entity.UserInvitation, error) {
	var inv entity.UserInvitation
	if err := database.DBFromContext(c, r.db).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&inv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrInvitationNotFound
		}
		return nil, err
	}
	return &inv, nil
}

// GetByTokenHash is the only unscoped lookup: the accept flow is
// unauthenticated and resolves the tenant from the token itself.
func (r *gormUserInvitationRepository) GetByTokenHash(c context.Context, tokenHash string) (*entity.UserInvitation, error) {
	var inv entity.UserInvitation
	if err := database.DBFromContext(c, r.db).
		Where("token_hash = ?", tokenHash).
		First(&inv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrInvitationNotFound
		}
		return nil, err
	}
	return &inv, nil
}

// conditionalUpdate applies updates only while the invitation is pending and
// returns ErrInvitationNotPending when no row matched (missing, foreign
// tenant, accepted or revoked).
func (r *gormUserInvitationRepository) conditionalUpdate(c context.Context, tenantID, id uuid.UUID, updates map[string]any) error {
	res := database.DBFromContext(c, r.db).
		Model(&entity.UserInvitation{}).
		Where(pendingWhere, id, tenantID).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainerrors.ErrInvitationNotPending
	}
	return nil
}

func (r *gormUserInvitationRepository) UpdateToken(c context.Context, tenantID, id uuid.UUID, tokenHash string, expiresAt time.Time) error {
	return r.conditionalUpdate(c, tenantID, id, map[string]any{
		"token_hash": tokenHash,
		"expires_at": expiresAt,
		"updated_at": time.Now().UTC(),
	})
}

func (r *gormUserInvitationRepository) MarkAccepted(c context.Context, tenantID, id uuid.UUID, at time.Time) error {
	return r.conditionalUpdate(c, tenantID, id, map[string]any{"accepted_at": at, "updated_at": at})
}

func (r *gormUserInvitationRepository) MarkRevoked(c context.Context, tenantID, id uuid.UUID, at time.Time) error {
	return r.conditionalUpdate(c, tenantID, id, map[string]any{"revoked_at": at, "updated_at": at})
}

func (r *gormUserInvitationRepository) List(c context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.UserInvitation, int64, error) {
	var total int64
	base := database.DBFromContext(c, r.db).Model(&entity.UserInvitation{}).Where("tenant_id = ?", tenantID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []*entity.UserInvitation
	q := database.DBFromContext(c, r.db).Where("tenant_id = ?", tenantID).Order("created_at DESC, id DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	if err := q.Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
