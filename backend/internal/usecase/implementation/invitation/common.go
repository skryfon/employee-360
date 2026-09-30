// Package invitation implements the onboarding-invitation usecases.
package invitation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/ctx"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
)

const (
	defaultInvitationExpiry = 7 * 24 * time.Hour
	roleAdmin               = "admin"
	roleSuperAdmin          = "super_admin"
	minPasswordLength       = 8
)

// tenantFromContext returns the tenant resolved by auth middleware. It is the
// only source of tenant identity for admin-facing invitation usecases.
func tenantFromContext(c context.Context) (uuid.UUID, error) {
	raw, ok := ctx.TenantIDFromContext(c)
	if !ok {
		return uuid.Nil, domainerrors.ErrUnauthorized
	}
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, domainerrors.ErrUnauthorized
	}
	return id, nil
}

// requireAdmin defends in depth; route middleware is the primary check.
func requireAdmin(c context.Context) error {
	roles, _ := ctx.RolesFromContext(c)
	for _, r := range roles {
		if r == roleAdmin || r == roleSuperAdmin {
			return nil
		}
	}
	return domainerrors.ErrForbidden
}

func newPlainToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
