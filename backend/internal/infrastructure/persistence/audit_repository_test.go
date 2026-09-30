//go:build integration

package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

func TestGormAuditRepository_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormAuditRepository(db)
	a := createTestTenant(t, db, "audit-a-"+uuid.NewString())
	b := createTestTenant(t, db, "audit-b-"+uuid.NewString())
	ca, cb, cn := tenantCtx(a.ID), tenantCtx(b.ID), context.Background()

	now := time.Now().UTC()
	entityID := uuid.New()
	mk := func(tenant uuid.UUID, eid uuid.UUID) *entity.AuditLog {
		return &entity.AuditLog{ID: uuid.New(), TenantID: tenant, Action: "update", EntityType: "role",
			EntityID: eid, Metadata: "{}", CreatedAt: now, UpdatedAt: now}
	}
	require.NoError(t, repo.Create(ca, mk(a.ID, entityID)))
	require.NoError(t, repo.Create(ca, mk(a.ID, uuid.New())))
	// Tenant B has a log on the SAME entity ID; it must not leak into A's results.
	require.NoError(t, repo.Create(cb, mk(b.ID, entityID)))

	// Happy path in own tenant.
	logs, total, err := repo.ListByTenantID(ca, 10, 0)
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, logs, 2)
	for _, l := range logs {
		require.Equal(t, a.ID, l.TenantID)
	}
	logs, total, err = repo.ListByEntity(ca, "role", entityID, 10, 0)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, logs, 1)
	require.Equal(t, a.ID, logs[0].TenantID)

	// Tenant B sees only its own row.
	logs, total, err = repo.ListByTenantID(cb, 10, 0)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, b.ID, logs[0].TenantID)

	// No tenant in context matches nothing.
	logs, total, err = repo.ListByTenantID(cn, 10, 0)
	require.NoError(t, err)
	require.Empty(t, logs)
	require.EqualValues(t, 0, total)
	logs, total, err = repo.ListByEntity(cn, "role", entityID, 10, 0)
	require.NoError(t, err)
	require.Empty(t, logs)
	require.EqualValues(t, 0, total)
}
