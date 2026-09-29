package auth

import (
	"context"
	"time"

	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// LoginInput contains parameters for authenticating a user.
type LoginInput struct {
	Email     string
	Password  string
	IPAddress string
	UserAgent string
}

// LoginOutput contains the result of a successful login.
type LoginOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	TokenType    string
	User         *entity.User
}

// TokenRefreshInput contains parameters for rotating a refresh token.
type TokenRefreshInput struct {
	RefreshToken string
	IPAddress    string
	UserAgent    string
}

// TokenRefreshOutput contains new tokens generated after refresh.
type TokenRefreshOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	TokenType    string
}

// LogoutInput contains parameters for revoking an active session.
type LogoutInput struct {
	RefreshToken string
}

// ForgotPasswordInput contains parameters for initiating a password reset.
type ForgotPasswordInput struct {
	Email string
}

// ResetPasswordInput contains parameters for completing a password reset.
type ResetPasswordInput struct {
	Token       string
	NewPassword string
}

// LoginUseCase defines the port for user login across all roles.
type LoginUseCase interface {
	Execute(ctx context.Context, input LoginInput) (*LoginOutput, error)
}

// TokenRefreshUseCase defines the port for rotating refresh tokens.
type TokenRefreshUseCase interface {
	Execute(ctx context.Context, input TokenRefreshInput) (*TokenRefreshOutput, error)
}

// LogoutUseCase defines the port for revoking sessions and refresh tokens.
type LogoutUseCase interface {
	Execute(ctx context.Context, input LogoutInput) error
}

// ForgotPasswordUseCase defines the port for requesting a password reset (enumeration-safe).
type ForgotPasswordUseCase interface {
	Execute(ctx context.Context, input ForgotPasswordInput) error
}

// ResetPasswordUseCase defines the port for executing a password reset with a valid token.
type ResetPasswordUseCase interface {
	Execute(ctx context.Context, input ResetPasswordInput) error
}
