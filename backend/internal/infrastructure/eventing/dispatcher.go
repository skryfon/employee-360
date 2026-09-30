// Package eventing bridges domain events to the River-backed transactional outbox.
package eventing

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/riverqueue/river"
	"github.com/skryfon/employee360/backend/internal/domain/event"
	"github.com/skryfon/employee360/backend/internal/domain/service"
	"github.com/skryfon/employee360/backend/internal/infrastructure/job"
)

// ErrUnsupportedEvent is returned for events the dispatcher has no mapping for.
var ErrUnsupportedEvent = errors.New("eventing: unsupported event type or payload")

// Dispatcher maps domain events to River job args.
type Dispatcher struct{}

// NewDispatcher constructs a Dispatcher.
func NewDispatcher() *Dispatcher { return &Dispatcher{} }

// Dispatch converts an event into the River job args that deliver it.
func (d *Dispatcher) Dispatch(evt event.Event) (river.JobArgs, error) {
	switch evt.EventType {
	case event.EventTypeUserInvited:
		p, ok := payloadAs[event.UserInvitedPayload](evt.Payload)
		if !ok {
			return nil, fmt.Errorf("%w: %s payload %T", ErrUnsupportedEvent, evt.EventType, evt.Payload)
		}
		return invitationArgs(p.TenantID.String(), p.Email, p.RoleName, p.PlainToken, p.InviteURL, p.ExpiresAt), nil
	case event.EventTypeInvitationResent:
		p, ok := payloadAs[event.InvitationResentPayload](evt.Payload)
		if !ok {
			return nil, fmt.Errorf("%w: %s payload %T", ErrUnsupportedEvent, evt.EventType, evt.Payload)
		}
		return invitationArgs(p.TenantID.String(), p.Email, p.RoleName, p.PlainToken, p.InviteURL, p.ExpiresAt), nil
	case event.EventTypePasswordResetRequested:
		p, ok := payloadAs[event.PasswordResetRequestedPayload](evt.Payload)
		if !ok {
			return nil, fmt.Errorf("%w: %s payload %T", ErrUnsupportedEvent, evt.EventType, evt.Payload)
		}
		return job.SendEmailArgs{
			TenantID:     p.TenantID.String(),
			TemplateName: service.EmailTemplatePasswordReset,
			To:           p.Email,
			TemplateData: map[string]interface{}{
				"email":       p.Email,
				"user_name":   strings.TrimSpace(p.UserName),
				"user_id":     p.UserID.String(),
				"tenant_id":   p.TenantID.String(),
				"plain_token": p.PlainToken,
				"reset_url":   p.ResetURL,
				"expires_at":  p.ExpiresAt.UTC().Format(time.RFC3339),
			},
		}, nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedEvent, evt.EventType)
	}
}

func invitationArgs(tenantID, email, role, token, inviteURL string, expiresAt time.Time) job.SendEmailArgs {
	return job.SendEmailArgs{
		TenantID:     tenantID,
		TemplateName: service.EmailTemplateUserInvitation,
		To:           email,
		TemplateData: map[string]interface{}{
			"email":       email,
			"tenant_id":   tenantID,
			"role_name":   role,
			"plain_token": token,
			"invite_url":  inviteURL,
			"expires_at":  expiresAt.UTC().Format(time.RFC3339),
		},
	}
}

// payloadAs accepts a payload as either a value or a pointer of type T.
func payloadAs[T any](v interface{}) (T, bool) {
	switch p := v.(type) {
	case T:
		return p, true
	case *T:
		if p != nil {
			return *p, true
		}
	}
	var zero T
	return zero, false
}
