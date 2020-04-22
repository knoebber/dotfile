package db

import (
	"fmt"
	"time"

	"database/sql"
	"github.com/knoebber/dotfile/usererr"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const (
	cliTokenLength = 24
	pwQuery        = "SELECT id, password_hash FROM users"
	userQuery      = "SELECT id, username, email, email_confirmed, cli_token, created_at FROM users"
)

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

// JoinDate returns a formatted date of a users join date.
func (u *User) JoinDate() string {
	return formatTime(u.CreatedAt)
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
);
CREATE INDEX IF NOT EXISTS users_username_index ON users(username);`
}

func (u *User) insertStmt() (sql.Result, error) {
	return connection.Exec("INSERT INTO users(username, email, password_hash, cli_token) VALUES(?, ?, ?, ?)",
		u.Username,
		u.Email,
		u.PasswordHash,
		u.CLIToken,
	)
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

// Looks up a user by their ID or username - only one is required.
// Returns their userID in for the case when the caller only has the username.
func compareUserPassword(userID int64, username string, password string) (int64, error) {
	var (
		row  *sql.Row
		hash []byte
		id   int64
	)

	if userID != 0 {
		row = connection.QueryRow(pwQuery+" WHERE id = ?", userID)
	} else if username != "" {
		row = connection.QueryRow(pwQuery+" WHERE username = ?", username)
	}
	if err := row.Scan(&id, &hash); err != nil {
		return 0, err
	}

	if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		return 0, err
	}

	return id, nil

}

// GetUser gets a user.
// Only one argument is required.
// This does not scan password_hash.
func GetUser(userID int64, username string) (*User, error) {
	var query *sql.Row

	user := new(User)
	if userID != 0 {
		query = connection.QueryRow(userQuery+" WHERE id = ?", userID)
	} else if username != "" {
		query = connection.QueryRow(userQuery+" WHERE username = ?", username)
	}

	err := query.
		Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.EmailConfirmed,
			&user.CLIToken,
			&user.CreatedAt,
		)
	if err != nil {
		return nil, errors.Wrapf(err, "querying for user (id:%d || username: %#v)", userID, username)
	}

	return user, nil
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

// UpdateEmail updates a users email and sets confirmed to false.
func UpdateEmail(userID int64, email string) error {
	if err := validate.Var(email, "email"); err != nil {
		return err
	}

	_, err := connection.Exec("UPDATE users SET email = ?, email_confirmed = 0 WHERE id = ?", email, userID)
	return err
}

// UpdatePassword updates a users password.
// currentPass must match the current hash.
func UpdatePassword(userID int64, currentPass, newPass string) error {
	_, err := compareUserPassword(userID, "", currentPass)
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return usererr.Invalid("Confirm does not match.")
	} else if err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.MinCost)
	if err != nil {
		return err
	}

	_, err = connection.Exec("UPDATE users SET password_hash = ? WHERE id = ?", hashed, userID)
	return err
}

// UserLogin checks a username / password.
// If the credentials are valid, returns a new session.
func UserLogin(username, password string) (*Session, error) {
	userID, err := compareUserPassword(0, username, password)
	if err != nil {
		return nil, err
	}

	return createSession(userID)
}

func cliToken() (string, error) {
	buff, err := randomBytes(cliTokenLength)
	if err != nil {
		return "", errors.Wrap(err, "generating cli token")
	}

	return fmt.Sprintf("%x", buff), nil
}
