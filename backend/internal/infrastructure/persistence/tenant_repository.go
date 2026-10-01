package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	"github.com/skryfon/employee360/backend/internal/infrastructure/database"
)

// gormTenantReader is a GORM-backed adapter implementing repository.TenantReader.
// The tenants table is the tenant boundary itself, so it carries no tenant_id
// and is looked up by primary key.
type gormTenantReader struct {
	db *gorm.DB
}

// NewGormTenantReader constructs a GORM-backed TenantReader.
func NewGormTenantReader(db *gorm.DB) repository.TenantReader {
	return &gormTenantReader{db: db}
}

var _ repository.TenantReader = (*gormTenantReader)(nil)

// GetByID returns the non-soft-deleted tenant with the given id.
func (r *gormTenantReader) GetByID(c context.Context, id uuid.UUID) (*entity.Tenant, error) {
	var tenant entity.Tenant
	if err := database.DBFromContext(c, r.db).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrTenantNotFound
		}
		return nil, err
	}
	return &tenant, nil
}
