package entity

import "time"

// User represents a user.
type User struct {
	ID        string     `db:"id"`
	Name      string     `db:"name"`
	Email     string     `db:"email"`
	Password  string     `db:"password"`
	Active    bool       `db:"active"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

// TableName returns the name of the database table for the User entity.
func (User) TableName() string {
	return "users"
}

// GetID returns the user ID.
func (u User) GetID() string {
	return u.ID
}

// GetName returns the user name.
func (u User) GetName() string {
	return u.Name
}

// GetEmail returns the user email.
func (u User) GetEmail() string {
	return u.Email
}
