package service

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	domainservice "github.com/skryfon/employee360/backend/internal/domain/service"
)

var (
	// ErrAuditInvalidIdentity is returned for a nil tenant or actor ID.
	ErrAuditInvalidIdentity = errors.New("audit: tenant and actor ids are required")
	// ErrAuditInvalidAction is returned when the action is not "entity.verb" shaped.
	ErrAuditInvalidAction = errors.New("audit: action must follow entity.verb naming")
	// ErrAuditInvalidEntityType is returned for an empty entity type.
	ErrAuditInvalidEntityType = errors.New("audit: entity type is required")
)

// actionPattern: two or more dot-separated lowercase snake_case segments.
var actionPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)

// sensitiveFragments are matched case-insensitively as substrings of keys.
var sensitiveFragments = []string{
	"password", "passwd", "token", "otp", "secret", "authorization",
	"api_key", "apikey", "credential", "private_key",
}

// auditRecorder implements domainservice.AuditRecorder over AuditRepository. The repository
// picks up the caller's transaction from ctx.
type auditRecorder struct {
	repo repository.AuditRepository
	now  func() time.Time
}

// NewAuditRecorder returns a repository-backed AuditRecorder.
func NewAuditRecorder(repo repository.AuditRepository) domainservice.AuditRecorder {
	return &auditRecorder{repo: repo, now: func() time.Time { return time.Now().UTC() }}
}

// Record validates, redacts and persists one audit entry.
func (r *auditRecorder) Record(ctx context.Context, tenantID, actorID uuid.UUID, action, entityType string, entityID uuid.UUID, meta map[string]any) error {
	if tenantID == uuid.Nil || actorID == uuid.Nil {
		return ErrAuditInvalidIdentity
	}
	if !actionPattern.MatchString(action) {
		return ErrAuditInvalidAction
	}
	if strings.TrimSpace(entityType) == "" {
		return ErrAuditInvalidEntityType
	}
	raw, err := json.Marshal(redact(meta))
	if err != nil {
		return err
	}
	now := r.now()
	return r.repo.Create(ctx, &entity.AuditLog{
		ID:          uuid.New(),
		TenantID:    tenantID,
		ActorUserID: &actorID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Metadata:    string(raw),
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func isSensitive(key string) bool {
	k := strings.ToLower(key)
	for _, f := range sensitiveFragments {
		if strings.Contains(k, f) {
			return true
		}
	}
	return false
}

// redact returns a copy of meta with sensitive keys removed, recursing into
// nested maps and slices. A nil meta stays nil (marshals to "null" as before).
func redact(meta map[string]any) map[string]any {
	if meta == nil {
		return nil
	}
	out := make(map[string]any, len(meta))
	for k, v := range meta {
		if isSensitive(k) {
			continue
		}
		out[k] = redactValue(v)
	}
	return out
}

func redactValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		return redact(t)
	case []any:
		cp := make([]any, len(t))
		for i := range t {
			cp[i] = redactValue(t[i])
		}
		return cp
	}
	return v
}
