package entity

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog tracks actions and mutations performed within a tenant.
type AuditLog struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	ActorUserID *uuid.UUID `json:"actor_user_id,omitempty"`
	Action      string     `json:"action"`
	EntityName  string     `json:"entity_name"`
	EntityID    string     `json:"entity_id"`
	OldValues   string     `json:"old_values,omitempty"`
	NewValues   string     `json:"new_values,omitempty"`
	IPAddress   string     `json:"ip_address,omitempty"`
	UserAgent   string     `json:"user_agent,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
