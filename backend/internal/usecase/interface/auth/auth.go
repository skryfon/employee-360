package auth

import (
	"context"

	"github.com/google/uuid"
	authtypes "github.com/skryfon/employee360/backend/internal/types/auth"
)

// LoginUseCase defines the port for user login across all roles.
//
// ipAddress and userAgent are passed as explicit parameters rather than fields
// on LoginRequest: they are server-derived HTTP request metadata used only for
// refresh-token audit fields -- neither is part of what the client actually
// submits to log in (see authtypes.LoginRequest).
type LoginUseCase interface {
	Execute(ctx context.Context, req authtypes.LoginRequest, ipAddress, userAgent string) (*authtypes.LoginResponse, error)
}

// TokenRefreshUseCase defines the port for rotating refresh tokens.
type TokenRefreshUseCase interface {
	Execute(ctx context.Context, req authtypes.TokenRefreshRequest) (*authtypes.TokenRefreshResponse, error)
}

// LogoutUseCase defines the port for revoking sessions and refresh tokens.
type LogoutUseCase interface {
	Execute(ctx context.Context, req authtypes.LogoutRequest) error
}

// ForgotPasswordUseCase defines the port for requesting a password reset (enumeration-safe).
//
// tenantID is passed as an explicit parameter rather than a field on
// ForgotPasswordRequest, for the same reason as LoginUseCase: it is resolved
// by the handler, not client-supplied.
type ForgotPasswordUseCase interface {
	Execute(ctx context.Context, tenantID uuid.UUID, req authtypes.ForgotPasswordRequest) error
}

// ResetPasswordUseCase defines the port for executing a password reset with a valid token.
type ResetPasswordUseCase interface {
	Execute(ctx context.Context, req authtypes.ResetPasswordRequest) error
}
