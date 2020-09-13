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

const tokenLength = 24

// UserRecord models the user table.
type UserRecord struct {
	ID                 int64
	Username           string `validate:"alphanum"`
	Email              string `validate:"omitempty,email"` // Not required; users may opt in to enable account recovery.
	EmailConfirmed     bool
	PasswordHash       []byte
	CLIToken           string `validate:"required"` // Allows CLI to write to server.
	PasswordResetToken *string
	Theme              string
	Timezone           string
	CreatedAt          string
}

func (*UserRecord) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS users(
id                   INTEGER PRIMARY KEY,
username             TEXT NOT NULL UNIQUE COLLATE NOCASE,
email                TEXT UNIQUE,
email_confirmed      INTEGER NOT NULL DEFAULT 0,
password_hash        BLOB NOT NULL,
cli_token            TEXT NOT NULL,
password_reset_token TEXT,
timezone             TEXT,
theme                TEXT NOT NULL DEFAULT "Light",
created_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS users_username_index ON users(username);`
}

func (u *UserRecord) insertStmt(e executor) (sql.Result, error) {
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

func (u *UserRecord) check() error {
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

// SetPasswordResetToken creates and saves a reset token for the user.
// Returns the newly created token.
func SetPasswordResetToken(email string) (string, error) {
	token, err := token()
	if err != nil {
		return "", err
	}

	res, err := connection.Exec(`
UPDATE users
SET password_reset_token = ?
WHERE email = ?`, token, email)
	if err != nil {
		return "", errors.Wrapf(err, "setting password reset token for %q", email)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return "", err
	}

	if affected == 0 {
		return "", fmt.Errorf("email %q does not exist", email)
	}

	return token, nil
}

// CheckPasswordResetToken checks if the password reset token exists.
// Returns username for the token on success.
func CheckPasswordResetToken(token string) (string, error) {
	var (
		count    int
		username *string
	)

	err := connection.
		QueryRow(`
SELECT COUNT(*),
       username
FROM users
WHERE password_reset_token = ?`, token).
		Scan(&count, &username)
	if err != nil {
		return "", errors.Wrap(err, "counting users for password reset")
	}
	if count == 0 {
		return "", usererror.Invalid("Token not found")
	}
	if count > 1 {
		return "", fmt.Errorf("token %q has %d matches", token, count)
	}

	return *username, nil

}

// ResetPassword hashes and sets a new password to the user with the password reset token.
func ResetPassword(token, newPassword string) (string, error) {
	username, err := CheckPasswordResetToken(token)
	if err != nil {
		return "", err
	}

	passwordHash, err := hashPassword(newPassword)
	if err != nil {
		return "", err
	}

	_, err = connection.Exec(`
UPDATE users
SET password_hash = ?, password_reset_token = NULL
WHERE username = ?`, passwordHash, username)
	if err != nil {
		return "", errors.Wrapf(err, "resetting %q password", username)
	}

	return username, nil
}

// User returns the user with username.
// This does not set password_hash or password_reset_token.
func User(username string) (*UserRecord, error) {
	var (
		email, timezone *string
		createdAt       time.Time
	)

	user := new(UserRecord)

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
		return nil, errors.Wrapf(err, "querying for user %q", username)
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
func CreateUser(username, password string) (*UserRecord, error) {
	passwordHash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	cliToken, err := token()
	if err != nil {
		return nil, err
	}

	u := &UserRecord{
		Username:     username,
		PasswordHash: passwordHash,
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

// RotateCLIToken creates a new CLI token for the user.
func RotateCLIToken(userID int64, currentToken string) error {
	newToken, err := token()
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
func UpdatePassword(username string, currentPass, newPassword string) error {
	if err := CheckPassword(username, currentPass); err != nil {
		return err
	}

	passwordHash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	_, err = connection.Exec(`
UPDATE users
SET password_hash = ?
WHERE username = ?`, passwordHash, username)
	return err
}

// UpdateTheme updates a users theme setting.
func UpdateTheme(username string, theme UserTheme) error {
	_, err := connection.Exec("UPDATE users SET theme = ? WHERE username = ?", theme, username)
	if err != nil {
		return errors.Wrapf(err, "updating %q to theme %q", username, theme)
	}

	return nil
}

// UserLogin checks a username / password.
// If the credentials are valid, returns a new session.
func UserLogin(username, password, ip string) (*SessionRecord, error) {
	if err := compareUserPassword(username, password); err != nil {
		return nil, err
	}

	return createSession(username, ip)
}

func token() (string, error) {
	buff, err := randomBytes(tokenLength)
	if err != nil {
		return "", errors.Wrap(err, "generating token")
	}

	return fmt.Sprintf("%x", buff), nil
}
