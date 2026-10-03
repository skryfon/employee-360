// Package position implements tenant-scoped position usecases.
package position

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
)

const (
	auditEntityPosition   = "position"
	auditActionCreate     = "position.create"
	auditActionUpdate     = "position.update"
	auditActionDelete     = "position.delete"
	auditActionActivate   = "position.activate"
	auditActionDeactivate = "position.deactivate"
)

// writeAudit records an admin mutation for position operations.
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
		EntityType:  auditEntityPosition,
		EntityID:    entityID,
		Metadata:    string(raw),
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}
