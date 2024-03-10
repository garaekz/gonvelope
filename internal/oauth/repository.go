package oauth

import (
	"context"

	"github.com/garaekz/gonvelope/internal/entity"
	"github.com/garaekz/gonvelope/pkg/dbcontext"
	"github.com/garaekz/gonvelope/pkg/log"
	dbx "github.com/go-ozzo/ozzo-dbx"
)

// Repository encapsulates the logic to access the data source.
type Repository interface {
	// GetProviderByName returns the oauth provider by name
	GetProviderByName(ctx context.Context, name string) (entity.Provider, error)
	// StoreUserProviderAccount stores the user provider account
	StoreUserProviderAccount(ctx context.Context, userProviderAccount entity.UserProviderAccount) error
}

// repository persists data in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new oauth repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// GetProviderByName returns the oauth provider by name
func (r repository) GetProviderByName(ctx context.Context, name string) (entity.Provider, error) {
	var provider entity.Provider
	err := r.db.With(ctx).Select().From("users").Where(dbx.HashExp{"name": name}).One(&provider)
	return provider, err
}

// GetUserProviderAccountByUserIDAndProviderID returns the user provider account by user id and provider id
func (r repository) GetUserProviderAccountByUserIDAndProviderID(ctx context.Context, userID string, providerID string) (entity.UserProviderAccount, error) {
	var userProviderAccount entity.UserProviderAccount
	err := r.db.With(ctx).Select().From("user_provider_accounts").Where(dbx.HashExp{"user_id": userID, "provider_id": providerID}).One(&userProviderAccount)
	return userProviderAccount, err
}

// StoreUserProviderAccount stores the user provider account
func (r repository) StoreUserProviderAccount(ctx context.Context, userProviderAccount entity.UserProviderAccount) error {
	err := r.db.With(ctx).Model(&userProviderAccount).Exclude("CreatedAt", "UpdatedAt").Insert()
	return err
}
