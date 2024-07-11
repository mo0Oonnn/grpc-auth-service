package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/mo0Oonnn/sso/internal/lib/jwt"
	"github.com/mo0Oonnn/sso/internal/lib/logger/helpers"
	"github.com/mo0Oonnn/sso/internal/models"
	"github.com/mo0Oonnn/sso/internal/storage"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

const (
	hashingCost = 15
)

type Auth struct {
	logger       *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passwordHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

// New returns a new instance of the Auth service
func New(
	logger *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		userSaver:    userSaver,
		userProvider: userProvider,
		logger:       logger,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

// Login checks the user credentials and returns a JWT token.
//
// If user doesn't exist, returns an error.
// If user exists, but password is incorrect, returns an error.
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const operation = "auth.Login"

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.logger.Warn("user not found", helpers.Error(err))

			return "", fmt.Errorf("%s: %w", operation, ErrInvalidCredentials)
		}

		a.logger.Error("failed to get user", helpers.Error(err))

		return "", fmt.Errorf("%s: %w", operation, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		a.logger.Warn("invalid credentials", ErrInvalidCredentials)

		return "", fmt.Errorf("%s: %w", operation, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", operation, err)
	}

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.logger.Error("failed to create token", helpers.Error(err))

		return "", fmt.Errorf("%s: %w", operation, err)
	}

	return token, nil
}

// RegisterNewUser registers a new user in the system and returns user ID.
//
// If user with the same username already exists, returns an error.
// If user with the same email already exists, returns an error.
func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
) (int64, error) {
	const operation = "auth.RegisterNewUser"

	logger := a.logger.With(
		slog.String("operation", operation),
	)

	passwordHash, err := hashPassword(password)
	if err != nil {
		logger.Error("failed to hash password", helpers.Error(err))

		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	uid, err := a.userSaver.SaveUser(ctx, email, passwordHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			logger.Warn("user already exists", ErrUserExists)

			return 0, fmt.Errorf("%s: %w", operation, ErrUserExists)
		}
		logger.Error("failed to save user", helpers.Error(err))

		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	return uid, nil
}

func hashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), hashingCost)
}

// IsAdmin checks if the user is an admin.
func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const operation = "auth.IsAdmin"

	isAdmin, err := a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.logger.Warn("user not found", helpers.Error(err))

			return false, fmt.Errorf("%s: %w", operation, ErrInvalidCredentials)
		}
		return false, fmt.Errorf("%s: %w", operation, err)
	}

	return isAdmin, nil
}
