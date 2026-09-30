// Package invitation holds request DTOs for the onboarding-invitation usecases.
//
// None of these carry a tenant ID: the tenant is always resolved from the
// authenticated request context (or, for accept, from the token itself).
package invitation

import "github.com/google/uuid"

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
