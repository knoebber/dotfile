package db

import (
	"fmt"
	"time"

	"database/sql"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordLength = 8
	cliTokenLength    = 32
)

// User is the model for a dotfilehub user.
type User struct {
	ID             int64
	Username       string  `validate:"required"`
	Email          *string `validate:"email"` // Not required; users may opt in to enable account recovery.
	EmailConfirmed bool
	PasswordHash   []byte
	CLIToken       string `validate:"required"` // Allows CLI to write to server.
	CreatedAt      time.Time
}

func (*User) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS users(
id              INTEGER PRIMARY KEY,
username        TEXT NOT NULL UNIQUE,
email           TEXT UNIQUE,
email_confirmed INTEGER NOT NULL DEFAULT 0,
password_hash   BLOB NOT NULL,
cli_token       TEXT NOT NULL UNIQUE,
created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
}

func (u *User) insertStmt() (sql.Result, error) {
	return connection.Exec("INSERT INTO users(username, email, password_hash, cli_token) VALUES(?, ?, ?, ?)",
		u.Username,
		u.Email,
		u.PasswordHash,
		u.CLIToken,
	)
}

func CreateUser(username, password string, email *string) (*User, error) {
	if len(password) < minPasswordLength {
		return nil, fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return nil, err
	}

	cliToken, err := cliToken()
	if err != nil {
		return nil, err
	}

	u := &User{
		Username:     username,
		Email:        email,
		PasswordHash: hashed,
		CLIToken:     cliToken,
	}

	id, err := insert(u)
	if err != nil {
		return nil, errors.Wrapf(err, "creating record for new user %#v", username)
	}

	u.ID = id

	return u, nil
}

func cliToken() (string, error) {
	buff, err := randomBytes(cliTokenLength)
	if err != nil {
		return "", errors.Wrap(err, "generating cli token")
	}

	return fmt.Sprintf("%x", buff), nil
}
