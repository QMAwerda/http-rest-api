package model

// the model has the same database fields
type User struct {
	ID                int
	Email             string
	EncryptedPassword string
}
