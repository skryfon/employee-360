// Package invitation implements the onboarding-invitation usecases.
package invitation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/ctx"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
)

const (
	auditEntityInvitation = "user_invitation"
	auditActionInvite     = "invitation.invite"
	auditActionResend     = "invitation.resend"
	auditActionRevoke     = "invitation.revoke"
)

// writeAudit records an admin mutation. Call it inside the mutation's
// transaction so the audit entry commits or rolls back with it.
func writeAudit(c context.Context, repo repository.AuditRepository, tenantID, actorID, entityID uuid.UUID, action string, meta map[string]any) error {
	raw, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	return repo.Create(c, &entity.AuditLog{
		ID: uuid.New(), TenantID: tenantID, ActorUserID: &actorID,
		Action: action, EntityType: auditEntityInvitation, EntityID: entityID,
		Metadata: string(raw), CreatedAt: now, UpdatedAt: now,
	})
}

// actorFromContext returns the authenticated admin's user ID.
func actorFromContext(c context.Context) (uuid.UUID, error) {
	raw, _ := ctx.UserIDFromContext(c)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, domainerrors.ErrUnauthorized
	}
	return id, nil
}

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
