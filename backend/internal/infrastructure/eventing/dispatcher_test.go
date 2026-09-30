package eventing

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/skryfon/employee360/backend/internal/domain/event"
	"github.com/skryfon/employee360/backend/internal/domain/service"
	"github.com/skryfon/employee360/backend/internal/infrastructure/job"
)

func TestDispatcher_MapsAllEmailEvents(t *testing.T) {
	tenantID := uuid.New()
	exp := time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
	d := NewDispatcher()

	cases := []struct {
		name     string
		evt      event.Event
		template service.EmailTemplateName
		urlKey   string
		url      string
	}{
		{"invited", event.Event{EventType: event.EventTypeUserInvited, Payload: event.UserInvitedPayload{TenantID: tenantID, Email: "a@x.io", RoleName: "employee", InviteURL: "http://i", ExpiresAt: exp}}, service.EmailTemplateUserInvitation, "invite_url", "http://i"},
		{"resent (pointer payload)", event.Event{EventType: event.EventTypeInvitationResent, Payload: &event.InvitationResentPayload{TenantID: tenantID, Email: "a@x.io", InviteURL: "http://r", ExpiresAt: exp}}, service.EmailTemplateUserInvitation, "invite_url", "http://r"},
		{"reset", event.Event{EventType: event.EventTypePasswordResetRequested, Payload: event.PasswordResetRequestedPayload{TenantID: tenantID, UserID: uuid.New(), Email: "a@x.io", UserName: "Ann Lee", ResetURL: "http://p", ExpiresAt: exp}}, service.EmailTemplatePasswordReset, "reset_url", "http://p"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args, err := d.Dispatch(tc.evt)
			require.NoError(t, err)
			got, ok := args.(job.SendEmailArgs)
			require.True(t, ok)
			assert.Equal(t, tc.template, got.TemplateName)
			assert.Equal(t, "a@x.io", got.To)
			assert.Equal(t, tenantID.String(), got.TenantID)
			assert.Equal(t, tc.url, got.TemplateData[tc.urlKey])
			assert.Equal(t, "2030-01-02T03:04:05Z", got.TemplateData["expires_at"])
			if tc.template == service.EmailTemplatePasswordReset {
				assert.Equal(t, "Ann Lee", got.TemplateData["user_name"])
			}
		})
	}
}

func TestDispatcher_RejectsUnknownEventAndBadPayload(t *testing.T) {
	d := NewDispatcher()
	_, err := d.Dispatch(event.Event{EventType: "Nope"})
	assert.ErrorIs(t, err, ErrUnsupportedEvent)
	_, err = d.Dispatch(event.Event{EventType: event.EventTypeUserInvited, Payload: "wrong"})
	assert.ErrorIs(t, err, ErrUnsupportedEvent)
}
