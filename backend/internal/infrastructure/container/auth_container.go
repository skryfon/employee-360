package container

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/skryfon/employee360/backend/config"
	"github.com/skryfon/employee360/backend/internal/delivery/http/handlers"
	"github.com/skryfon/employee360/backend/internal/domain/event"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	domainservice "github.com/skryfon/employee360/backend/internal/domain/service"
	"github.com/skryfon/employee360/backend/internal/infrastructure/persistence"
	infraservice "github.com/skryfon/employee360/backend/internal/infrastructure/service"
	authusecaseimpl "github.com/skryfon/employee360/backend/internal/usecase/implementation/auth"
	"github.com/skryfon/employee360/backend/internal/usecase/implementation/ucshared"
	authusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/auth"
	"gorm.io/gorm"
)

type zerologLoggerAdapter struct {
	logger zerolog.Logger
}

func (z *zerologLoggerAdapter) Error(ctx context.Context, msg string, err error) {
	z.logger.Error().Err(err).Msg(msg)
}

type nopEventPublisher struct{}

func (n *nopEventPublisher) Publish(ctx context.Context, events ...event.Event) error {
	return nil
}

// AuthContainer encapsulates dependencies, repositories, services, usecases, and handlers for Authentication.
type AuthContainer struct {
	TokenService domainservice.TokenService
	HashService  domainservice.HashService
	Handler      *handlers.AuthHandler

	// Usecases
	LoginUseCase          authusecase.LoginUseCase
	TokenRefreshUseCase   authusecase.TokenRefreshUseCase
	LogoutUseCase         authusecase.LogoutUseCase
	ForgotPasswordUseCase authusecase.ForgotPasswordUseCase
	ResetPasswordUseCase  authusecase.ResetPasswordUseCase

	// Repositories
	UserRepo          repository.UserRepository
	RefreshTokenRepo  repository.RefreshTokenRepository
	PasswordResetRepo repository.PasswordResetRepository
}

// NewAuthContainer initializes and wires all auth-related repositories, services, usecases, and handlers.
func NewAuthContainer(
	cfg *config.Config,
	db *gorm.DB,
	log zerolog.Logger,
	transactor ucshared.Transactor,
	eventPublisher domainservice.EventPublisher,
) (*AuthContainer, error) {
	jwtSecret := cfg.JWT.Secret
	if jwtSecret == "" {
		jwtSecret = config.DefaultJWTSecret
	}
	accessExpiry := cfg.JWT.AccessExpiry
	if accessExpiry == 0 {
		accessExpiry = 15 * time.Minute
	}
	refreshExpiry := cfg.JWT.RefreshExpiry
	if refreshExpiry == 0 {
		refreshExpiry = 7 * 24 * time.Hour
	}

	tokenService, err := infraservice.NewJWTService(jwtSecret, accessExpiry, refreshExpiry)
	if err != nil {
		return nil, err
	}

	hashService := infraservice.NewHashService()
	loggerAdapter := &zerologLoggerAdapter{logger: log}

	userRepo := persistence.NewGormUserRepository(db)
	refreshTokenRepo := persistence.NewGormRefreshTokenRepository(db)
	tenantDomainRepo := persistence.NewGormTenantDomainRepository(db)
	passwordResetRepo := persistence.NewGormPasswordResetRepository(db)

	if transactor == nil {
		transactor = ucshared.NewNopTransactor()
	}
	if eventPublisher == nil {
		eventPublisher = &nopEventPublisher{}
	}

	loginUC := authusecaseimpl.NewLoginUseCase(userRepo, tenantDomainRepo, tokenService, hashService, refreshTokenRepo, loggerAdapter)
	tokenRefreshUC := authusecaseimpl.NewTokenRefreshUseCase(userRepo, tokenService, hashService, refreshTokenRepo, loggerAdapter)
	logoutUC := authusecaseimpl.NewLogoutUseCase(tokenService, hashService, refreshTokenRepo)
	forgotPasswordUC := authusecaseimpl.NewForgotPasswordUseCase(userRepo, tenantDomainRepo, passwordResetRepo, hashService, eventPublisher, transactor, loggerAdapter)
	resetPasswordUC := authusecaseimpl.NewResetPasswordUseCase(userRepo, passwordResetRepo, refreshTokenRepo, hashService, transactor)

	handler := handlers.NewAuthHandler(
		loginUC,
		tokenRefreshUC,
		logoutUC,
		forgotPasswordUC,
		resetPasswordUC,
	)

	return &AuthContainer{
		TokenService:          tokenService,
		HashService:           hashService,
		Handler:               handler,
		LoginUseCase:          loginUC,
		TokenRefreshUseCase:   tokenRefreshUC,
		LogoutUseCase:         logoutUC,
		ForgotPasswordUseCase: forgotPasswordUC,
		ResetPasswordUseCase:  resetPasswordUC,
		UserRepo:              userRepo,
		RefreshTokenRepo:      refreshTokenRepo,
		PasswordResetRepo:     passwordResetRepo,
	}, nil
}
