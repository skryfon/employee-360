package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	"github.com/skryfon/employee360/backend/internal/domain/event"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	"github.com/skryfon/employee360/backend/internal/domain/service"
	"github.com/skryfon/employee360/backend/internal/usecase/implementation/ucshared"
	authusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/auth"
)

const (
	defaultResetTokenExpiry = 15 * time.Minute
)

// ForgotPasswordUseCaseImpl implements authusecase.ForgotPasswordUseCase.
type ForgotPasswordUseCaseImpl struct {
	userRepo          repository.UserRepository
	passwordResetRepo repository.PasswordResetRepository
	hashService       service.HashService
	eventPublisher    service.EventPublisher
	transactor        ucshared.Transactor
	tokenExpiry       time.Duration
}

var _ authusecase.ForgotPasswordUseCase = (*ForgotPasswordUseCaseImpl)(nil)

// NewForgotPasswordUseCase constructs a new ForgotPasswordUseCaseImpl.
func NewForgotPasswordUseCase(
	userRepo repository.UserRepository,
	passwordResetRepo repository.PasswordResetRepository,
	hashService service.HashService,
	eventPublisher service.EventPublisher,
	transactor ucshared.Transactor,
) *ForgotPasswordUseCaseImpl {
	return &ForgotPasswordUseCaseImpl{
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		hashService:       hashService,
		eventPublisher:    eventPublisher,
		transactor:        transactor,
		tokenExpiry:       defaultResetTokenExpiry,
	}
}

// Execute initiates an enumeration-safe password reset flow.
// It always returns nil to prevent user enumeration attacks.
func (u *ForgotPasswordUseCaseImpl) Execute(ctx context.Context, input authusecase.ForgotPasswordInput) error {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" {
		_ = u.hashService.HashToken("dummy-timing-mitigation-token")
		return nil
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		_ = u.hashService.HashToken("dummy-timing-mitigation-token")
		return nil
	}

	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil || !user.IsActive {
		_ = u.hashService.HashToken("dummy-timing-mitigation-token")
		return nil
	}

	// Generate 32 cryptographically secure random bytes
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil
	}
	plainToken := hex.EncodeToString(tokenBytes)
	tokenHash := u.hashService.HashToken(plainToken)

	now := time.Now().UTC()
	tokenEntity := &entity.PasswordResetToken{
		ID:        uuid.New(),
		TenantID:  user.TenantID,
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(u.tokenExpiry),
		CreatedAt: now,
		UpdatedAt: now,
	}

	resetEvent := event.Event{
		ID:            uuid.New(),
		TenantID:      user.TenantID,
		EventType:     event.EventTypePasswordResetRequested,
		AggregateType: "user",
		AggregateID:   user.ID.String(),
		Payload: event.PasswordResetRequestedPayload{
			UserID:     user.ID,
			TenantID:   user.TenantID,
			Email:      user.Email,
			PlainToken: plainToken,
			ResetURL:   "",
			ExpiresAt:  tokenEntity.ExpiresAt,
		},
		OccurredAt: now,
	}

	_ = u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := u.passwordResetRepo.InvalidateAllForUser(txCtx, user.ID); err != nil {
			return err
		}

		if err := u.passwordResetRepo.Create(txCtx, tokenEntity); err != nil {
			return err
		}

		if err := u.eventPublisher.Publish(txCtx, resetEvent); err != nil {
			return err
		}

		return nil
	})

	return nil
}
