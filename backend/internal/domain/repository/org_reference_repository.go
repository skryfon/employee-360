package repository

import (
	"context"

	"github.com/google/uuid"
)

// OrgReferenceRepository answers tenant-scoped existence checks for
// organisation records referenced by other aggregates (e.g. an invitation's
// department and position). tenantID always comes from the verified request
// context, never from client input.
type OrgReferenceRepository interface {
	DepartmentExists(ctx context.Context, tenantID, id uuid.UUID) (bool, error)
	PositionExists(ctx context.Context, tenantID, id uuid.UUID) (bool, error)
}
