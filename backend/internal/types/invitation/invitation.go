// Package invitation holds request DTOs for the onboarding-invitation usecases.
//
// None of these carry a tenant ID: the tenant is always resolved from the
// authenticated request context (or, for accept, from the token itself).
package invitation

import (
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// InviteUserRequest is what an admin submits to invite a user.
type InviteUserRequest struct {
	Email        string     `json:"email"`
	RoleID       uuid.UUID  `json:"role_id"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	PositionID   *uuid.UUID `json:"position_id,omitempty"`
	FirstName    string     `json:"first_name,omitempty"`
	LastName     string     `json:"last_name,omitempty"`
}

// AcceptInvitationRequest is what an invitee submits to activate their account.
type AcceptInvitationRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// InvitationResponse is the API representation of an invitation (never includes the token).
type InvitationResponse struct {
	ID           uuid.UUID               `json:"id"`
	Email        string                  `json:"email"`
	RoleID       uuid.UUID               `json:"role_id"`
	DepartmentID *uuid.UUID              `json:"department_id,omitempty"`
	PositionID   *uuid.UUID              `json:"position_id,omitempty"`
	InvitedBy    uuid.UUID               `json:"invited_by"`
	Status       entity.InvitationStatus `json:"status"`
	ExpiresAt    time.Time               `json:"expires_at"`
	AcceptedAt   *time.Time              `json:"accepted_at,omitempty"`
	RevokedAt    *time.Time              `json:"revoked_at,omitempty"`
	CreatedAt    time.Time               `json:"created_at"`
}

// ToInvitationResponse maps an invitation entity to its API representation.
func ToInvitationResponse(i *entity.UserInvitation) InvitationResponse {
	return InvitationResponse{
		ID: i.ID, Email: i.Email, RoleID: i.RoleID, DepartmentID: i.DepartmentID,
		PositionID: i.PositionID, InvitedBy: i.InvitedBy, Status: i.Status(time.Now()),
		ExpiresAt: i.ExpiresAt, AcceptedAt: i.AcceptedAt, RevokedAt: i.RevokedAt, CreatedAt: i.CreatedAt,
	}
}

// ValidateInvitationResponse is returned when checking if an invitation token is valid before password entry.
type ValidateInvitationResponse struct {
	Valid bool   `json:"valid"`
	Email string `json:"email"`
	// Role is the invitee's role name (e.g. "admin", "employee"); clients use it
	// to send the invitee to the correct app.
	Role string `json:"role"`
}
