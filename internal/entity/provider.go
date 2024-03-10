package entity

import "time"

// Provider represents a oauth provider.
type Provider struct {
	ID        string     `db:"id"`
	Name      string     `db:"name"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

// TableName returns the name of the database table for the Provider entity.
func (Provider) TableName() string {
	return "providers"
}

// GetID returns the provider ID.
func (p Provider) GetID() string {
	return p.ID
}

// GetName returns the provider name.
func (p Provider) GetName() string {
	return p.Name
}
