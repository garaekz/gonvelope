package entity

import "github.com/google/uuid"

// GenerateID generates a unique ID that can be used as an identifier for an entity.
func GenerateID() string {
	id, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}
	return id.String()
}
