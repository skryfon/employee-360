package entity

import (
	"time"

	"github.com/google/uuid"
)

// BaseEntity contains standard identity and timestamp fields.
type BaseEntity struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantBaseEntity contains standard tenant-scoped identity and timestamp fields.
type TenantBaseEntity struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
