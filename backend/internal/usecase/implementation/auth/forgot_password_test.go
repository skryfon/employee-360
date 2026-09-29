package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/event"
	"github.com/skryfon/employee360/backend/internal/usecase/implementation/ucshared"
	authusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/auth"
)

// MockPasswordResetRepository implements repository.PasswordResetRepository.
type mockPasswordResetRepository struct {
	tokensByHash       map[string]*entity.PasswordResetToken
	tokensByID         map[uuid.UUID]*entity.PasswordResetToken
	invalidatedUserIDs []uuid.UUID
	usedTokenIDs       []uuid.UUID
}

func newMockPasswordResetRepository() *mockPasswordResetRepository {
	return &mockPasswordResetRepository{
		tokensByHash: make(map[string]*entity.PasswordResetToken),
		tokensByID:   make(map[uuid.UUID]*entity.PasswordResetToken),
	}
}

func (m *mockPasswordResetRepository) Create(ctx context.Context, token *entity.PasswordResetToken) error {
	m.tokensByHash[token.TokenHash] = token
	m.tokensByID[token.ID] = token
	return nil
}

func (m *mockPasswordResetRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.PasswordResetToken, error) {
	t, ok := m.tokensByHash[tokenHash]
	if !ok {
		return nil, domainerrors.ErrNotFound
	}
	return t, nil
}

func (m *mockPasswordResetRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	m.usedTokenIDs = append(m.usedTokenIDs, id)
	if t, ok := m.tokensByID[id]; ok {
		now := time.Now().UTC()
		t.UsedAt = &now
	}
	return nil
}

func (m *mockPasswordResetRepository) InvalidateAllForUser(ctx context.Context, userID uuid.UUID) error {
	m.invalidatedUserIDs = append(m.invalidatedUserIDs, userID)
	now := time.Now().UTC()
	for _, t := range m.tokensByID {
		if t.UserID == userID {
			t.UsedAt = &now
		}
	}
	return nil
}

func (m *mockPasswordResetRepository) DeleteExpiredTokens(ctx context.Context, before time.Time) error {
	return nil
}

// MockEventPublisher implements service.EventPublisher.
type mockEventPublisher struct {
	publishedEvents []event.Event
}

func (m *mockEventPublisher) Publish(ctx context.Context, events ...event.Event) error {
	m.publishedEvents = append(m.publishedEvents, events...)
	return nil
}

func TestForgotPasswordUseCase_ExistingUser(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()

	userRepo := newMockUserRepository()
	resetRepo := newMockPasswordResetRepository()
	hashSvc := &mockHashService{}
	eventPub := &mockEventPublisher{}
	transactor := ucshared.NewNopTransactor()

	user := &entity.User{
		ID:       userID,
		TenantID: tenantID,
		Email:    "alice@example.com",
		IsActive: true,
	}
	_ = userRepo.Create(context.Background(), user)

	uc := NewForgotPasswordUseCase(userRepo, resetRepo, hashSvc, eventPub, transactor)

	err := uc.Execute(context.Background(), authusecase.ForgotPasswordInput{
		Email: "alice@example.com",
	})

	if err != nil {
		t.Fatalf("expected nil error (enumeration-safe), got: %v", err)
	}

	// Verify token was stored in repository with a hash
	if len(resetRepo.tokensByID) != 1 {
		t.Fatalf("expected 1 password reset token in repository, got %d", len(resetRepo.tokensByID))
	}

	var storedToken *entity.PasswordResetToken
	for _, tk := range resetRepo.tokensByID {
		storedToken = tk
	}

	if storedToken.UserID != userID {
		t.Errorf("expected user ID %s, got %s", userID, storedToken.UserID)
	}
	if storedToken.TenantID != tenantID {
		t.Errorf("expected tenant ID %s, got %s", tenantID, storedToken.TenantID)
	}
	// Verify raw token is never stored directly
	if storedToken.TokenHash == "" || storedToken.TokenHash[:7] != "sha256_" {
		t.Errorf("expected token to be hashed before persistence, got: %s", storedToken.TokenHash)
	}

	// Verify event was published
	if len(eventPub.publishedEvents) != 1 {
		t.Fatalf("expected 1 domain event published, got %d", len(eventPub.publishedEvents))
	}

	evt := eventPub.publishedEvents[0]
	if evt.EventType != event.EventTypePasswordResetRequested {
		t.Errorf("expected EventTypePasswordResetRequested, got %s", evt.EventType)
	}
	if evt.TenantID != tenantID {
		t.Errorf("expected event tenant ID %s, got %s", tenantID, evt.TenantID)
	}

	payload, ok := evt.Payload.(event.PasswordResetRequestedPayload)
	if !ok {
		t.Fatalf("expected PasswordResetRequestedPayload type in event payload")
	}
	if payload.Email != "alice@example.com" {
		t.Errorf("expected payload email alice@example.com, got %s", payload.Email)
	}
	if payload.PlainToken == "" {
		t.Errorf("expected non-empty plain token in event payload")
	}
	if hashSvc.HashToken(payload.PlainToken) != storedToken.TokenHash {
		t.Errorf("hash of event plain token does not match stored token hash")
	}
}

func TestForgotPasswordUseCase_EnumerationSafe(t *testing.T) {
	userRepo := newMockUserRepository()
	resetRepo := newMockPasswordResetRepository()
	hashSvc := &mockHashService{}
	eventPub := &mockEventPublisher{}
	transactor := ucshared.NewNopTransactor()

	uc := NewForgotPasswordUseCase(userRepo, resetRepo, hashSvc, eventPub, transactor)

	tests := []struct {
		name  string
		email string
	}{
		{"nonexistent user", "unknown@example.com"},
		{"invalid format", "not-an-email"},
		{"empty email", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			eventPub.publishedEvents = nil
			err := uc.Execute(context.Background(), authusecase.ForgotPasswordInput{
				Email: tc.email,
			})

			// Acceptance criterion: identical response (nil error)
			if err != nil {
				t.Errorf("expected nil error for enumeration safety, got %v", err)
			}
			if len(eventPub.publishedEvents) != 0 {
				t.Errorf("expected no events published for invalid/missing user, got %d", len(eventPub.publishedEvents))
			}
		})
	}
}

func TestForgotPasswordUseCase_InactiveUser(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()

	userRepo := newMockUserRepository()
	resetRepo := newMockPasswordResetRepository()
	hashSvc := &mockHashService{}
	eventPub := &mockEventPublisher{}
	transactor := ucshared.NewNopTransactor()

	user := &entity.User{
		ID:       userID,
		TenantID: tenantID,
		Email:    "inactive@example.com",
		IsActive: false,
	}
	_ = userRepo.Create(context.Background(), user)

	uc := NewForgotPasswordUseCase(userRepo, resetRepo, hashSvc, eventPub, transactor)

	err := uc.Execute(context.Background(), authusecase.ForgotPasswordInput{
		Email: "inactive@example.com",
	})

	if err != nil {
		t.Fatalf("expected nil error for inactive user (enumeration-safe), got %v", err)
	}
	if len(eventPub.publishedEvents) != 0 {
		t.Errorf("expected 0 published events for inactive user, got %d", len(eventPub.publishedEvents))
	}
}
