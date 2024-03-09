package auth

import (
	"context"

	"github.com/garaekz/gonvelope/internal/entity"
	"github.com/garaekz/gonvelope/pkg/dbcontext"
	"github.com/garaekz/gonvelope/pkg/log"
	dbx "github.com/go-ozzo/ozzo-dbx"
)

// Repository encapsulates the logic to access info from the data source.
type Repository interface {
	// GetUserByEmail passes the email to the database and returns the user
	GetUserByEmail(ctx context.Context, email string) (entity.User, error)
	// GetActiveUserByEmail passes the email to the database and returns the active user
	GetActiveUserByEmail(ctx context.Context, email string) (entity.User, error)
	// CreateUser stores a new user in the database
	CreateUser(ctx context.Context, user entity.User) error
}

// repository persists users in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new auth repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// GetUserByEmail passes the email to the database and returns the user even if it's not active
func (r repository) GetUserByEmail(ctx context.Context, email string) (entity.User, error) {
	var user entity.User
	err := r.db.With(ctx).Select().From("users").Where(dbx.HashExp{"email": email}).One(&user)
	if err != nil {
		return user, err
	}

	return user, nil
}

// GetActiveUserByEmail passes the email to the database and returns the active user
func (r repository) GetActiveUserByEmail(ctx context.Context, email string) (entity.User, error) {
	var user entity.User
	err := r.db.With(ctx).Select().From("users").Where(dbx.HashExp{"email": email, "active": true}).One(&user)
	if err != nil {
		return user, err
	}

	return user, nil
}

// CreateUser stores a new user in the database
func (r repository) CreateUser(ctx context.Context, user entity.User) error {
	return r.db.With(ctx).Model(&user).Exclude("CreatedAt", "UpdatedAt", "Active").Insert()
}
