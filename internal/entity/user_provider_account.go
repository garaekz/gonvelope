package entity

import "time"

// UserProviderAccount represents a oauth provider.
type UserProviderAccount struct {
	ID           string     `db:"id"`
	UserID       string     `db:"user_id"`
	ProviderID   string     `db:"provider_id"`
	AccessToken  string     `db:"access_token"`
	RefreshToken string     `db:"refresh_token"`
	TokenExpiry  time.Time  `db:"token_expiry"`
	IsDefault    bool       `db:"is_default"`
	CreatedAt    *time.Time `db:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at"`
}

// TableName returns the name of the database table for the UserProviderAccount entity.
func (UserProviderAccount) TableName() string {
	return "user_provider_accounts"
}

// GetID returns the UserProviderAccount ID.
func (u UserProviderAccount) GetID() string {
	return u.ID
}
