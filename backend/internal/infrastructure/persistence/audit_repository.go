package persistence

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/skryfon/employee360/backend/internal/domain/entity"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	"github.com/skryfon/employee360/backend/internal/infrastructure/database"
)

// gormAuditRepository implements repository.AuditRepository. Create persists
// the entity's TenantID (set by the usecase from context); reads are scoped by
// the tenant in context (a missing tenant matches nothing).
type gormAuditRepository struct {
	db *gorm.DB
}

// NewGormAuditRepository constructs a GORM-backed AuditRepository.
func NewGormAuditRepository(db *gorm.DB) repository.AuditRepository {
	return &gormAuditRepository{db: db}
}

var _ repository.AuditRepository = (*gormAuditRepository)(nil)

func (r *gormAuditRepository) Create(c context.Context, log *entity.AuditLog) error {
	return database.DBFromContext(c, r.db).Create(log).Error
}

func (r *gormAuditRepository) list(c context.Context, q *gorm.DB, limit, offset int) ([]*entity.AuditLog, int64, error) {
	q = q.Session(&gorm.Session{})
	var total int64
	if err := q.Model(&entity.AuditLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []*entity.AuditLog
	page := q.Order("created_at DESC, id DESC")
	if limit > 0 {
		page = page.Limit(limit)
	}
	if offset > 0 {
		page = page.Offset(offset)
	}
	if err := page.Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *gormAuditRepository) tenantQuery(c context.Context) *gorm.DB {
	tenantID, err := uuid.Parse(tenantString(c))
	if err != nil {
		tenantID = uuid.Nil
	}
	return database.DBFromContext(c, r.db).Where("tenant_id = ?", tenantID)
}

func (r *gormAuditRepository) ListByTenantID(c context.Context, limit, offset int) ([]*entity.AuditLog, int64, error) {
	return r.list(c, r.tenantQuery(c), limit, offset)
}

func (r *gormAuditRepository) ListByEntity(c context.Context, entityType string, entityID uuid.UUID, limit, offset int) ([]*entity.AuditLog, int64, error) {
	return r.list(c, r.tenantQuery(c).Where("entity_type = ? AND entity_id = ?", entityType, entityID), limit, offset)
}
