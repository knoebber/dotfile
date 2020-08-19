package db

import (
	"fmt"
	"log"
	"time"

	"database/sql"
	"github.com/knoebber/dotfile/usererror"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// UserTheme is a users theme preference.
type UserTheme string

// Valid values for UserTheme.
const (
	UserThemeLight UserTheme = "Light"
	UserThemeDark  UserTheme = "Dark"
)

const cliTokenLength = 24

// User is the model for a dotfilehub user.
type User struct {
	ID             int64
	Username       string `validate:"alphanum"`        // TODO make regex, usernames should be allowed to have underscores etc.
	Email          string `validate:"omitempty,email"` // Not required; users may opt in to enable account recovery.
	EmailConfirmed bool
	PasswordHash   []byte
	CLIToken       string `validate:"required"` // Allows CLI to write to server.
	Theme          string
	Timezone       string
	CreatedAt      string
}

func (*User) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS users(
id              INTEGER PRIMARY KEY,
username        TEXT NOT NULL UNIQUE COLLATE NOCASE,
email           TEXT UNIQUE,
email_confirmed INTEGER NOT NULL DEFAULT 0,
password_hash   BLOB NOT NULL,
cli_token       TEXT NOT NULL,
timezone        TEXT,
theme           TEXT NOT NULL DEFAULT "Light",
created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS users_username_index ON users(username);`
}

func (u *User) insertStmt(e executor) (sql.Result, error) {
	var email *string
	if u.Email != "" {
		email = &u.Email
	}
	return e.Exec("INSERT INTO users(username, email, password_hash, cli_token) VALUES(?, ?, ?, ?)",
		u.Username,
		email,
		u.PasswordHash,
		u.CLIToken,
	)
}

func (u *User) check() error {
	var count int

	if err := checkUsernameAllowed(u.Username); err != nil {
		return err
	}

	err := connection.
		QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", u.Username).
		Scan(&count)
	if err != nil {
		return errors.Wrapf(err, "checking if username %#v is unique", u.Username)
	}

	if count > 0 {
		return usererror.Duplicate("Username", u.Username)
	}

	if u.Email != "" {
		return checkUniqueEmail(u.Email)
	}

	return nil
}

func checkUniqueEmail(email string) error {
	var count int

	err := connection.
		QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)

	if err != nil {
		return errors.Wrapf(err, "checking if email %#v is unique", email)
	}

	if count > 0 {
		return usererror.Duplicate("Email", email)
	}

	return nil
}

// Looks up a user by their username and compares the password to the stored hash.
func compareUserPassword(username string, password string) error {
	var hash []byte

	err := connection.
		QueryRow("SELECT password_hash FROM users WHERE username = ?", username).
		Scan(&hash)
	if err != nil {
		return errors.Wrapf(err, "querying for user %q password hash", username)
	}

	if err = bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		return err
	}

	return nil

}

// GetUser gets a user.
// Only one argument is required - userID will be used if both are present.
// This does not scan password_hash.
func GetUser(username string) (*User, error) {
	var (
		email, timezone *string
		createdAt       time.Time
	)

	user := new(User)

	err := connection.QueryRow(`
SELECT id,
       username,
       email,
       email_confirmed,
       cli_token,
       theme,
       timezone,
       created_at
FROM users
WHERE username = ?
`, username).Scan(
		&user.ID,
		&user.Username,
		&email,
		&user.EmailConfirmed,
		&user.CLIToken,
		&user.Theme,
		&timezone,
		&createdAt,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "querying for user %#v", username)
	}
	if email != nil {
		user.Email = *email
	}
	if timezone != nil {
		user.Timezone = *timezone
	}
	user.CreatedAt = formatTime(createdAt, timezone)

	return user, nil
}

// CreateUser inserts a new user into the users table.
func CreateUser(username, password string) (*User, error) {
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
		PasswordHash: hashed,
		CLIToken:     cliToken,
	}

	id, err := insert(u, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "creating record for new user %#v", username)
	}

	u.ID = id

	return u, nil
}

// UpdateEmail updates a users email and sets email_confirmed to false.
func UpdateEmail(userID int64, email string) error {
	if err := validate.Var(email, "email"); err != nil {
		return err
	}
	if err := checkUniqueEmail(email); err != nil {
		return err
	}

	_, err := connection.Exec("UPDATE users SET email = ?, email_confirmed = 0 WHERE id = ?", email, userID)
	return err
}

// UpdateTimezone checks if the timezone is able to be loaded and updates the user record.
func UpdateTimezone(userID int64, timezone string) error {
	if _, err := time.LoadLocation(timezone); err != nil {
		log.Printf("error updating timezone: %v", err)
		return usererror.Invalid(fmt.Sprintf("Timezone %q not found", timezone))
	}

	_, err := connection.Exec("UPDATE users SET timezone = ? WHERE id = ?", timezone, userID)
	return err
}

// RotateToken creates a new token for the user.
func RotateToken(userID int64, currentToken string) error {
	newToken, err := cliToken()
	if err != nil {
		return err
	}

	res, err := connection.Exec(`
UPDATE users
SET cli_token = ?
WHERE id = ? AND cli_token = ?`, newToken, userID, currentToken)
	if err != nil {
		return errors.Wrap(err, "rotating cli token")
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return usererror.Invalid("User token mismatch")
	}

	return err
}

// CheckPassword checks username and password combination.
// Tells the user when the password does not match.
func CheckPassword(username, password string) error {
	err := compareUserPassword(username, password)
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return usererror.Invalid("Password does not match.")
	} else if err != nil {
		return err
	}

	return nil
}

// UpdatePassword updates a users password.
// currentPass must match the current hash.
func UpdatePassword(username string, currentPass, newPass string) error {
	if err := CheckPassword(username, currentPass); err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.MinCost)
	if err != nil {
		return err
	}

	_, err = connection.Exec("UPDATE users SET password_hash = ? WHERE username = ?", hashed, username)
	return err
}

// UpdateTheme updates a users theme setting.
func UpdateTheme(username string, theme UserTheme) error {
	_, err := connection.Exec("UPDATE users SET theme = ? WHERE username = ?", theme, username)
	if err != nil {
		return errors.Wrapf(err, "updating %#v to theme %#v", username, theme)
	}

	return nil
}

// UserLogin checks a username / password.
// If the credentials are valid, returns a new session.
func UserLogin(username, password string) (*Session, error) {
	if err := compareUserPassword(username, password); err != nil {
		return nil, err
	}

	return createSession(username)
}

func cliToken() (string, error) {
	buff, err := randomBytes(cliTokenLength)
	if err != nil {
		return "", errors.Wrap(err, "generating cli token")
	}

	return fmt.Sprintf("%x", buff), nil
}
