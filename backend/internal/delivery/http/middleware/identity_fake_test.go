package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	domainservice "github.com/skryfon/employee360/backend/internal/domain/service"
	authtypes "github.com/skryfon/employee360/backend/internal/types/auth"
)

// fakeIdentity is the fake database row state for one user (and its tenant).
type fakeIdentity struct {
	TenantID      uuid.UUID
	Roles         []string
	UserInactive  bool
	UserDeleted   bool
	TenantMissing bool
	TenantDeleted bool
	TenantOff     bool
}

// fakeVerifier is a VerifyIdentityUseCase backed by an in-memory table keyed by
// user id. Err, when set, simulates a repository/DB failure.
type fakeVerifier struct {
	mu    sync.Mutex
	users map[uuid.UUID]fakeIdentity
	Err   error
	calls int
}

func newFakeVerifier() *fakeVerifier { return &fakeVerifier{users: map[uuid.UUID]fakeIdentity{}} }

func (f *fakeVerifier) set(userID uuid.UUID, id fakeIdentity) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.users[userID] = id
}

func (f *fakeVerifier) Execute(_ context.Context, tenantID, userID uuid.UUID) (*authtypes.VerifiedIdentity, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.Err != nil {
		return nil, f.Err
	}
	id, ok := f.users[userID]
	if !ok || id.TenantMissing || id.TenantDeleted || id.TenantOff ||
		id.TenantID != tenantID || id.UserInactive || id.UserDeleted {
		return nil, domainerrors.ErrUnauthorized
	}
	return &authtypes.VerifiedIdentity{TenantID: tenantID, UserID: userID, Roles: id.Roles}, nil
}

// registeringTokenService wraps a TokenService so every access token minted in a
// test also registers a matching healthy identity in the fake verifier (roles
// taken from the claims), keeping pre-existing tests unchanged.
type registeringTokenService struct {
	domainservice.TokenService
	verifier *fakeVerifier
}

func (r *registeringTokenService) GenerateAccessToken(claims domainservice.AccessTokenClaims) (string, time.Time, error) {
	r.verifier.set(claims.UserID, fakeIdentity{TenantID: claims.TenantID, Roles: claims.Roles})
	return r.TokenService.GenerateAccessToken(claims)
}
