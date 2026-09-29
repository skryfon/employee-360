package errors

import "errors"

var (
	// Generic errors
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrInternal      = errors.New("internal error")

	// Tenant errors
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrDomainAlreadyExists = errors.New("domain already registered")

	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrEmailAlreadyExists = errors.New("email already registered in tenant")
	ErrInvalidCredentials = errors.New("invalid email or password")

	// Role errors
	ErrRoleNotFound     = errors.New("role not found")
	ErrUserRoleNotFound = errors.New("user role not found")

	// Token errors
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenRevoked = errors.New("token has been revoked")
)
