package db

import (
	"fmt"
	"time"

	"database/sql"
	"github.com/knoebber/dotfile/usererr"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const cliTokenLength = 24

// User is the model for a dotfilehub user.
type User struct {
	ID             int64
	Username       string  `validate:"required"`
	Email          *string `validate:"omitempty,email"` // Not required; users may opt in to enable account recovery.
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

func getUser(username string) (*User, error) {
	user := new(User)

	err := connection.
		QueryRow("SELECT * FROM users WHERE username = ?", username).
		Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.EmailConfirmed,
			&user.PasswordHash,
			&user.CLIToken,
			&user.CreatedAt,
		)
	if err != nil {
		return nil, errors.Wrapf(err, "querying for user %#v", username)
	}

	return user, nil
}

func validateUserInfo(username, password string, email *string) error {
	n, err := count("users", "username", username)
	if err != nil {
		return err
	}

	if n > 0 {
		return usererr.Duplicate("Username", username)
	}

	if email == nil {
		return nil
	}

	n, err = count("users", "email", *email)
	if err != nil {
		return err
	}

	if n > 0 {
		return usererr.Duplicate("Email", *email)
	}

	return nil
}

// CreateUser inserts a new user into the users table.
// Email is optional.
func CreateUser(username, password string, email *string) (*User, error) {
	if err := validateUserInfo(username, password, email); err != nil {
		return nil, err
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

// UserLogin checks a username / password.
// If the credentials are valid, returns a new session.
func UserLogin(username, password string) (*Session, error) {
	user, err := getUser(username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return nil, err
	}

	return createSession(user.ID)
}

func cliToken() (string, error) {
	buff, err := randomBytes(cliTokenLength)
	if err != nil {
		return "", errors.Wrap(err, "generating cli token")
	}

	return fmt.Sprintf("%x", buff), nil
}
