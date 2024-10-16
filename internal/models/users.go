package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
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
	hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created)
    VALUES(?, ?, ?, UTC_TIMESTAMP())`

	_, err = m.DB.Exec(stmt, name, email, hashed_password)
	if err != nil {
		var mySQLError *mysql.MySQLError

		// NOTE: errors.As() writes the matching error to the target if it exists
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}

		return err
	}

	return nil
}

// Authenticate() checks if the user exists with the povided email and password and returns there userID
func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashed_password []byte

	stmt := `SELECT id, hashed_password FROM users WHERE email = ?;`

	// NOTE: we are using queryrow beacuse this stmt returns a single row
	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashed_password)
	if err != nil {
		return 0, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword(hashed_password, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		}

		return 0, err
	}

	return id, nil
}

// Exists() checks if the user exists with the provided ID
func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool
	stmt := `SELECT EXISTS(SELECT true FROM users WHERE id = ?)`

	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}
