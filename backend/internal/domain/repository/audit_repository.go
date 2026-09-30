package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// AuditRepository defines the data access methods for audit logs.
type AuditRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
	ListByTenantID(ctx context.Context, limit, offset int) ([]*entity.AuditLog, int64, error)
	ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID, limit, offset int) ([]*entity.AuditLog, int64, error)
}
