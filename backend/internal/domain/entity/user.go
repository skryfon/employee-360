package entity

import (
	"time"

	"github.com/google/uuid"
)

// User represents an individual user within a tenant organization.
type User struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	DepartmentID    *uuid.UUID `json:"department_id,omitempty"`
	PositionID      *uuid.UUID `json:"position_id,omitempty"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	Email           string     `json:"email"`
	PasswordHash    *string    `json:"-"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	IsActive        bool       `json:"is_active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Associations (populated when loaded)
	Roles []Role `json:"roles,omitempty"`
}

// FullName returns the concatenated first and last name.
func (u *User) FullName() string {
	if u.FirstName == "" {
		return u.LastName
	}
	if u.LastName == "" {
		return u.FirstName
	}
	return u.FirstName + " " + u.LastName
}
