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
	ErrTenantNotFound        = errors.New("tenant not found")
	ErrEmailDomainNotAllowed = errors.New("email domain is not registered for this organization")
	ErrDomainAlreadyExists   = errors.New("domain already registered")
	ErrDomainNotFound        = errors.New("tenant domain not found")
	ErrInvalidDomain         = errors.New("invalid domain name")
	ErrInvalidTenantName     = errors.New("invalid tenant name")
	// ErrLastDomain is returned when removing a tenant's only live domain.
	ErrLastDomain = errors.New("cannot remove the last domain of a tenant")

	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrEmailAlreadyExists = errors.New("email already registered in tenant")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters")

	// Role errors
	ErrRoleNotFound     = errors.New("role not found")
	ErrUserRoleNotFound = errors.New("user role not found")

	// Token errors
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenRevoked = errors.New("token has been revoked")

	// Invitation errors
	ErrInvitationNotFound   = errors.New("invitation not found")
	ErrInvitationNotPending = errors.New("invitation is no longer pending")
	// Distinct token-state errors returned by validate/accept when the token
	// hash matches a row but the invitation cannot be used.
	ErrInvitationExpired       = errors.New("invitation has expired")
	ErrInvitationRevoked       = errors.New("invitation has been revoked")
	ErrInvitationAccepted      = errors.New("invitation has already been accepted")
	ErrInvalidRole             = errors.New("role cannot be assigned by invitation")
	ErrInvalidEmail            = errors.New("invalid email address")
	ErrInvalidInvitationFilter = errors.New("invalid invitation list filter")
	ErrDepartmentNotFound      = errors.New("department not found")
	ErrPositionNotFound        = errors.New("position not found")
)
