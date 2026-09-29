package service

import "context"

// EmailTemplateName identifies transactional email templates.
type EmailTemplateName string

const (
	EmailTemplatePasswordReset  EmailTemplateName = "PasswordReset"
	EmailTemplateUserInvitation EmailTemplateName = "UserInvitation"
)

// EmailMessage defines the data structure for sending transactional emails.
type EmailMessage struct {
	To           string                 `json:"to"`
	Subject      string                 `json:"subject"`
	TemplateName EmailTemplateName      `json:"template_name"`
	TemplateData map[string]interface{} `json:"template_data"`
}

// EmailService is the interface for sending transactional emails.
// It is called exclusively by background workers, never directly by usecases.
type EmailService interface {
	Send(ctx context.Context, msg EmailMessage) error
}
