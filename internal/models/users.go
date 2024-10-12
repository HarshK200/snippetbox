package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID              int
	name            string
	email           string
	hashed_password []byte
	create          time.Time
}

type UserModel struct {
	DB *sql.DB
}

// inserts a new user in the database with the provided values. if failed returns an error
func (m *UserModel) Insert(name, email, password string) error {
	return nil
}

// Authenticate() checks if the user exists with the povided email and password and returns there userID
func (m *UserModel) Authenticate(email, password string) (int, error) {
	return 0, nil
}

// Exists() checks if the user exists with the provided ID
func (m *UserModel) Exists(id int) (bool, error) {
	return false, nil
}
