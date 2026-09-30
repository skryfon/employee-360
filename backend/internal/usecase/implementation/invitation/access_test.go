package invitation

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/ctx"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	invtypes "github.com/skryfon/employee360/backend/internal/types/invitation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Non-admins and callers without a tenant are rejected by every management usecase.
func TestManagementUseCases_RequireAdminAndTenant(t *testing.T) {
	f := newFixture()
	inv, _ := f.doInvite(t, "x@acme.com")
	employee := ctx.WithRoles(f.adminCtx, []string{"employee"})
	noTenant := ctx.WithRoles(context.Background(), []string{"admin"})
	list := NewListInvitationsUseCase(fakeInvRepo{f.s})

	for name, c := range map[string]struct {
		ctx context.Context
		err error
	}{
		"employee":  {employee, domainerrors.ErrForbidden},
		"no tenant": {noTenant, domainerrors.ErrUnauthorized},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := f.resend().Execute(c.ctx, inv.ID)
			assert.ErrorIs(t, err, c.err)
			assert.ErrorIs(t, f.revoke().Execute(c.ctx, inv.ID), c.err)
			_, _, err = list.Execute(c.ctx, 10, 0)
			assert.ErrorIs(t, err, c.err)
		})
	}
	assert.True(t, f.s.invitations[inv.ID].IsPending(), "rejected calls must not change the invitation")
}

func TestAccept_EmptyTokenIsInvalid(t *testing.T) {
	f := newFixture()
	err := f.accept().Execute(context.Background(), invtypes.AcceptInvitationRequest{Token: "  ", Password: "password123"})
	assert.ErrorIs(t, err, domainerrors.ErrInvalidToken)
}

// limitRecorder wraps the fake repo to record the paging args the usecase passes down.
type limitRecorder struct {
	fakeInvRepo
	limit, offset int
}

func (r *limitRecorder) List(c context.Context, t uuid.UUID, limit, offset int) ([]*entity.UserInvitation, int64, error) {
	r.limit, r.offset = limit, offset
	return r.fakeInvRepo.List(c, t, limit, offset)
}

func TestList_PaginationBounds(t *testing.T) {
	tests := []struct {
		name                  string
		limit, offset         int
		wantLimit, wantOffset int
	}{
		{"passes through", 50, 10, 50, 10},
		{"max allowed", 100, 0, 100, 0},
		{"over max uses default", 101, 0, 20, 0},
		{"zero uses default", 0, 0, 20, 0},
		{"negative offset clamps to 0", 10, -5, 10, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			rec := &limitRecorder{fakeInvRepo: fakeInvRepo{f.s}}
			_, _, err := NewListInvitationsUseCase(rec).Execute(f.adminCtx, tt.limit, tt.offset)
			require.NoError(t, err)
			assert.Equal(t, tt.wantLimit, rec.limit)
			assert.Equal(t, tt.wantOffset, rec.offset)
		})
	}
}
