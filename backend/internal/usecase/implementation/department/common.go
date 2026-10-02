// Package department implements tenant-scoped department usecases.
package department

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
)

const (
	auditEntityDepartment = "department"
	auditActionCreate     = "department.create"
	auditActionUpdate     = "department.update"
	auditActionDelete     = "department.delete"
	auditActionActivate   = "department.activate"
	auditActionDeactivate = "department.deactivate"
)

// writeAudit records an admin mutation for department operations.
func writeAudit(c context.Context, repo repository.AuditRepository, tenantID, actorID, entityID uuid.UUID, action string, meta map[string]any) error {
	if repo == nil {
		return nil
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	return repo.Create(c, &entity.AuditLog{
		ID:          uuid.New(),
		TenantID:    tenantID,
		ActorUserID: &actorID,
		Action:      action,
		EntityType:  auditEntityDepartment,
		EntityID:    entityID,
		Metadata:    string(raw),
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}
