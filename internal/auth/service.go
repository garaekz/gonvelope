package auth

import (
	"context"
	"time"

	"github.com/garaekz/gonvelope/internal/entity"
	"github.com/garaekz/gonvelope/internal/errors"
	"github.com/garaekz/gonvelope/pkg/log"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Service encapsulates the authentication logic.
type Service interface {
	// authenticate authenticates a user using email and password.
	// It returns a JWT token if authentication succeeds. Otherwise, an error is returned.
	Login(ctx context.Context, email, password string) (string, error)
	// Register registers a new user.
	Register(ctx context.Context, name, email, password string) error
}

// Identity represents an authenticated user identity.
type Identity interface {
	// GetID returns the user ID.
	GetID() string
	// GetName returns the user name.
	GetName() string
	// GetEmail returns the user email.
	GetEmail() string
}

type service struct {
	repo            Repository
	signingKey      string
	tokenExpiration int
	logger          log.Logger
}

// NewService creates a new authentication service.
func NewService(repo Repository, signingKey string, tokenExpiration int, logger log.Logger) Service {
	return service{repo, signingKey, tokenExpiration, logger}
}

// Login authenticates a user and generates a JWT token if authentication succeeds.
// Otherwise, an error is returned.
func (s service) Login(ctx context.Context, email, password string) (string, error) {
	if identity := s.authenticate(ctx, email, password); identity != nil {
		return s.generateJWT(identity)
	}
	return "", errors.Unauthorized("Authentication failed, check your email and password and try again")
}

// Register registers a new user.
func (s service) Register(ctx context.Context, name, email, password string) error {
	if _, err := s.repo.GetUserByEmail(ctx, email); err == nil {
		return errors.BadRequest("email already exists")
	}
	pass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.InternalServerError("Failed to hash password")
	}

	user := entity.User{
		ID:       entity.GenerateID(),
		Name:     name,
		Email:    email,
		Password: string(pass),
	}

	return s.repo.CreateUser(ctx, user)
}

// authenticate authenticates a user using email and password.
// If username and password are correct, an identity is returned. Otherwise, nil is returned.
func (s service) authenticate(ctx context.Context, email, password string) Identity {
	logger := s.logger.With(ctx, "email", email)

	user, err := s.repo.GetActiveUserByEmail(ctx, email)
	if err != nil {
		logger.Infof("User not found: Authentication failed")
		return nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		logger.Infof("Authentication failed")
		return nil
	}

	return entity.User{ID: user.GetID(), Name: user.GetName(), Email: user.GetEmail()}
}

// generateJWT generates a JWT that encodes an identity.
func (s service) generateJWT(identity Identity) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    identity.GetID(),
		"name":  identity.GetName(),
		"email": identity.GetEmail(),
		"exp":   time.Now().Add(time.Duration(s.tokenExpiration) * time.Hour).Unix(),
	}).SignedString([]byte(s.signingKey))
}
