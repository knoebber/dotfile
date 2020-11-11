package db

import (
	"fmt"
	"log"
	"strings"
	"time"

	"database/sql"
	"github.com/knoebber/dotfile/usererror"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// UserTheme is a users theme preference.
type UserTheme string

// Valid Values for UserTheme.
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

func (u *UserRecord) insertStmt(e Executor) (sql.Result, error) {
	var email *string
	if u.Email != "" {
		u.Email = strings.ToLower(u.Email)
		email = &u.Email
	}
	return e.Exec("INSERT INTO users(username, email, password_hash, cli_token) VALUES(?, ?, ?, ?)",
		u.Username,
		email,
		u.PasswordHash,
		u.CLIToken,
	)
}

func (u *UserRecord) check(e Executor) error {
	if err := validateStringSizes(u.Username, u.Email); err != nil {
		return err
	}

	if err := checkUsernameAllowed(e, u.Username); err != nil {
		return err
	}

	exists, err := userExists(e, u.Username)
	if err != nil {
		return err
	}

	if exists {
		return usererror.Duplicate("Username", u.Username)
	}

	if u.Email != "" {
		return checkUniqueEmail(e, u.Email)
	}

	return nil
}

func checkUniqueEmail(e Executor, email string) error {
	var count int

	err := e.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", strings.ToLower(email)).
		Scan(&count)

	if err != nil {
		return errors.Wrapf(err, "checking if email %q is unique", email)
	}

	if count > 0 {
		return usererror.Duplicate("Email", email)
	}

	return nil
}

// Looks up a user by their username and compares the password to the stored hash.
func compareUserPassword(e Executor, username string, password string) error {
	var hash []byte

	err := e.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).
		Scan(&hash)
	if err != nil {
		return errors.Wrapf(err, "querying for user %q password hash", username)
	}

	if err = bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		return err
	}

	return nil

}

// ValidateUserExists throws a sql.ErrNoRows when the username does not exist.
func ValidateUserExists(e Executor, username string) error {
	exists, err := userExists(e, username)
	if err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}
	return nil
}

// SetPasswordResetToken creates and saves a reset token for the user.
// Returns the newly created token.
func SetPasswordResetToken(e Executor, email string) (string, error) {
	token, err := token()
	if err != nil {
		return "", err
	}

	res, err := e.Exec(`
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
func CheckPasswordResetToken(e Executor, token string) (string, error) {
	var (
		count    int
		username *string
	)

	err := e.QueryRow(`
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
func ResetPassword(e Executor, token, newPassword string) (string, error) {
	username, err := CheckPasswordResetToken(e, token)
	if err != nil {
		return "", err
	}

	passwordHash, err := hashPassword(newPassword)
	if err != nil {
		return "", err
	}

	_, err = e.Exec(`
UPDATE users
SET password_hash = ?, password_reset_token = NULL
WHERE username = ?`, passwordHash, username)
	if err != nil {
		return "", errors.Wrapf(err, "resetting %q password", username)
	}

	return username, nil
}

// CreateUser inserts a new user into the users table.
func CreateUser(e Executor, username, password string) (*UserRecord, error) {
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

	id, err := insert(e, u)
	if err != nil {
		return nil, errors.Wrapf(err, "creating record for new user %q", username)
	}

	u.ID = id

	return u, nil
}

// UpdateEmail updates a users email and sets email_confirmed to false.
func UpdateEmail(e Executor, userID int64, email string) error {
	email = strings.ToLower(email)

	if err := validateStringSizes(email); err != nil {
		return err
	}

	if err := validate.Var(email, "email"); err != nil {
		return err
	}
	if err := checkUniqueEmail(e, email); err != nil {
		return err
	}

	_, err := e.Exec("UPDATE users SET email = ?, email_confirmed = 0 WHERE id = ?", email, userID)
	return err
}

// UpdateTimezone checks if the timezone is able to be loaded and updates the user record.
func UpdateTimezone(e Executor, userID int64, timezone string) error {
	if _, err := time.LoadLocation(timezone); err != nil {
		log.Printf("error updating timezone: %v", err)
		return usererror.Invalid(fmt.Sprintf("Timezone %q not found", timezone))
	}

	_, err := e.Exec("UPDATE users SET timezone = ? WHERE id = ?", timezone, userID)
	return err
}

// RotateCLIToken creates a new CLI token for the user.
func RotateCLIToken(e Executor, userID int64, currentToken string) error {
	newToken, err := token()
	if err != nil {
		return err
	}

	res, err := e.Exec(`
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
func CheckPassword(e Executor, username, password string) error {
	err := compareUserPassword(e, username, password)
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return usererror.Invalid("Password does not match.")
	} else if err != nil {
		return err
	}

	return nil
}

// UpdatePassword updates a users password.
// currentPass must match the current hash.
func UpdatePassword(e Executor, username string, currentPass, newPassword string) error {
	if err := CheckPassword(e, username, currentPass); err != nil {
		return err
	}

	passwordHash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	_, err = e.Exec(`
UPDATE users
SET password_hash = ?
WHERE username = ?`, passwordHash, username)
	return err
}

// UpdateTheme updates a users theme setting.
func UpdateTheme(e Executor, userID int64, theme UserTheme) error {
	_, err := e.Exec("UPDATE users SET theme = ? WHERE id = ?", theme, userID)
	if err != nil {
		return errors.Wrapf(err, "updating user %d to theme %q", userID, theme)
	}

	return nil
}

// UserLogin checks a username / password.
// If the credentials are valid, returns a new session.
func UserLogin(e Executor, username, password, ip string) (*SessionRecord, error) {
	if err := compareUserPassword(e, username, password); err != nil {
		return nil, err
	}

	return createSession(e, username, ip)
}

// UserLoginAPI checks a username and CLI token combination.
func UserLoginAPI(e Executor, username, cliToken string) (int64, error) {
	var userID int64

	err := e.QueryRow(`
SELECT id FROM users WHERE username = ? AND cli_token = ?`, username, cliToken).
		Scan(&userID)
	if err != nil {
		return 0, errors.Wrapf(err, "attempting to authenticate user %q for API access", username)
	}

	return userID, nil
}

// DeleteUser deletes a user and their data.
func DeleteUser(username, password string) error {
	tx, err := Connection.Begin()
	if err != nil {
		return errors.Wrap(err, "starting transaction for delete user")
	}
	if err := deleteUser(tx, username, password); err != nil {
		return Rollback(tx, err)
	}
	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction for delete user")
	}

	return nil
}

func userExists(e Executor, username string) (bool, error) {
	var count int
	err := e.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).
		Scan(&count)
	if err != nil {
		return false, errors.Wrapf(err, "checking if username %q exists", username)
	}

	return count == 1, nil
}

func deleteUser(tx *sql.Tx, username, password string) error {
	if err := compareUserPassword(tx, username, password); err != nil {
		return err
	}

	if err := DeleteTempFile(tx, username); err != nil {
		return err
	}

	fileList, err := FilesByUsername(tx, username, nil)
	if err != nil {
		return err
	}

	for _, f := range fileList {
		if err := DeleteFile(tx, username, f.Alias); err != nil {
			return errors.Wrap(err, "deleting files by username")
		}
	}

	_, err = tx.Exec(`
DELETE FROM sessions 
WHERE user_id = (SELECT id FROM users WHERE username = ?)`, username)
	if err != nil {
		return errors.Wrapf(err, "deleting sessions for user %q", username)
	}

	_, err = tx.Exec("DELETE FROM users WHERE username = ?", username)
	if err != nil {
		return errors.Wrapf(err, "deleting user %q", username)
	}

	return nil
}

func token() (string, error) {
	buff, err := randomBytes(tokenLength)
	if err != nil {
		return "", errors.Wrap(err, "generating token")
	}

	return fmt.Sprintf("%x", buff), nil
}
