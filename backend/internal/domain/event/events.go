package event

import (
	"time"

	"github.com/google/uuid"
)

// EventType identifies the kind of domain event.
type EventType string

const (
	EventTypeUserInvited            EventType = "UserInvited"
	EventTypeInvitationResent       EventType = "InvitationResent"
	EventTypePasswordResetRequested EventType = "PasswordResetRequested"
)

// Event represents an immutable domain event occurring within a tenant.
type Event struct {
	ID            uuid.UUID   `json:"id"`
	TenantID      uuid.UUID   `json:"tenant_id"`
	EventType     EventType   `json:"event_type"`
	AggregateType string      `json:"aggregate_type"`
	AggregateID   string      `json:"aggregate_id"`
	Payload       interface{} `json:"payload"`
	OccurredAt    time.Time   `json:"occurred_at"`
}

// UserInvitedPayload holds the data carried by a UserInvited domain event.
type UserInvitedPayload struct {
	InvitationID uuid.UUID `json:"invitation_id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	Email        string    `json:"email"`
	RoleName     string    `json:"role_name"`
	PlainToken   string    `json:"plain_token"`
	InviteURL    string    `json:"invite_url"`
	InvitedBy    uuid.UUID `json:"invited_by"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// InvitationResentPayload holds the data carried by an InvitationResent domain event.
type InvitationResentPayload struct {
	InvitationID uuid.UUID `json:"invitation_id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	Email        string    `json:"email"`
	RoleName     string    `json:"role_name"`
	PlainToken   string    `json:"plain_token"`
	InviteURL    string    `json:"invite_url"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// PasswordResetRequestedPayload holds the data carried by a PasswordResetRequested domain event.
type PasswordResetRequestedPayload struct {
	UserID     uuid.UUID `json:"user_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	Email      string    `json:"email"`
	UserName   string    `json:"user_name,omitempty"`
	PlainToken string    `json:"plain_token"`
	ResetURL   string    `json:"reset_url"`
	ExpiresAt  time.Time `json:"expires_at"`
}
