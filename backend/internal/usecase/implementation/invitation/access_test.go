package invitation

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	invtypes "github.com/skryfon/employee360/backend/internal/types/invitation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A nil tenant or actor is rejected by every management usecase before any work happens.
func TestManagementUseCases_RequireIdentity(t *testing.T) {
	f := newFixture()
	inv, _ := f.doInvite(t, "x@acme.com")
	bg := context.Background()
	list := NewListInvitationsUseCase(fakeInvRepo{f.s})

	for name, c := range map[string]struct{ tenant, actor uuid.UUID }{
		"no tenant": {uuid.Nil, f.adminID},
		"no actor":  {f.tenantA, uuid.Nil},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := f.resend().Execute(bg, c.tenant, c.actor, inv.ID)
			assert.ErrorIs(t, err, domainerrors.ErrUnauthorized)
			assert.ErrorIs(t, f.revoke().Execute(bg, c.tenant, c.actor, inv.ID), domainerrors.ErrUnauthorized)
			if c.tenant == uuid.Nil {
				_, _, err = list.Execute(bg, c.tenant, 10, 0)
				assert.ErrorIs(t, err, domainerrors.ErrUnauthorized)
			}
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
			_, _, err := NewListInvitationsUseCase(rec).Execute(context.Background(), f.tenantA, tt.limit, tt.offset)
			require.NoError(t, err)
			assert.Equal(t, tt.wantLimit, rec.limit)
			assert.Equal(t, tt.wantOffset, rec.offset)
		})
	}
}
