// Package db stores and manipulates dotfiles via a sqlite3 database.
package db

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/knoebber/dotfile/dotfile"
	"github.com/knoebber/dotfile/usererror"
	"golang.org/x/crypto/bcrypt"
	// Driver for sql
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
)

const (
	timestampDisplayFormat = "January 02, 2006 3:04PM"
	maxFilesPerUser        = 100
	maxCommitsPerFile      = 100
	maxBlobSizeBytes       = 15000
	minPasswordLength      = 8
	maxStringSize          = 50
)

// Connection is a global database connection.
// Call Start() to initialize.
var Connection *sql.DB

// Validates data before its inserted.
var validate *validator.Validate

// Executor is an interface for executing SQL.
type Executor interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
}

type tableCreator interface {
	createStmt() string
}

type inserter interface {
	insertStmt(Executor) (sql.Result, error)
}

type checker interface {
	check(e Executor) error
}

func validateStringSizes(strings ...string) error {
	for _, s := range strings {
		if len(s) > maxStringSize {
			return usererror.Invalid(fmt.Sprintf(
				"The maximum string length is %d characters", maxStringSize))
		}
	}
	return nil
}

// Create the required tables when they don't exist
func createTables(e Executor) error {
	for _, model := range []tableCreator{
		new(UserRecord),
		new(ReservedUsernameRecord),
		new(SessionRecord),
		new(FileRecord),
		new(TempFileRecord),
		new(CommitRecord),
	} {
		_, err := e.Exec(model.createStmt())
		if err != nil {
			return errors.Wrap(err, "creating tables")
		}
	}
	return nil
}

func insert(e Executor, i inserter) (id int64, err error) {
	if err = validate.Struct(i); err != nil {
		log.Print(err)
		return 0, usererror.Invalid("Values are missing or improperly formatted.")
	}

	if c, ok := i.(checker); ok {
		if err := c.check(e); err != nil {
			return 0, err
		}
	}

	res, err := i.insertStmt(e)
	if err != nil {
		return 0, err
	}

	id, err = res.LastInsertId()

	if err != nil {
		return 0, err
	}

	return id, nil
}

// Rollback reverts a database transaction.
// Always returns an error.
func Rollback(tx *sql.Tx, err error) error {
	if rbError := tx.Rollback(); rbError != nil {
		return errors.Wrapf(err, "failed to rollback database transaction: %v", rbError)
	}

	return errors.Wrap(err, "rolled back from error")
}

// Used for generating random IDs.
func randomBytes(n int) ([]byte, error) {
	buff := make([]byte, n)

	if _, err := io.ReadFull(rand.Reader, buff); err != nil {
		return nil, err
	}

	return buff, nil
}

func formatTime(t time.Time, timezone *string) string {
	if timezone != nil {
		loc, err := time.LoadLocation(*timezone)
		if err != nil {
			log.Print(err)
			return "Unknown timezone"
		}

		t = t.In(loc)
	}

	return t.Format(timestampDisplayFormat)
}

func checkSize(content []byte, name string) error {
	if len(content) == 0 {
		return usererror.Invalid(fmt.Sprintf("%s is empty", name))
	}

	if len(content) > maxBlobSizeBytes {
		return usererror.Invalid(fmt.Sprintf("%s is too large (max=%dKB)", name, maxBlobSizeBytes/1000))
	}
	return nil

}

func checkFile(alias, path string) error {
	if err := validateStringSizes(alias, path); err != nil {
		return err
	}

	if err := dotfile.CheckAlias(alias); err != nil {
		return err
	}

	if err := dotfile.CheckPath(path); err != nil {
		return err
	}
	return nil
}

func hashPassword(password string) ([]byte, error) {

	if len(password) < minPasswordLength {
		return nil, usererror.Invalid(fmt.Sprintf("Password must be at least %d characters.", minPasswordLength))
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return nil, errors.Wrap(err, "hashing password")
	}

	return passwordHash, nil
}

// NotFound returns whether err is wrapping a no rows error.
func NotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

// Start opens a connection a sqlite3 database.
// Creates a new sqlite database with all required tables when not found.
func Start(dbPath string) (err error) {
	dsn := "?_foreign_keys=true"
	Connection, err = sql.Open("sqlite3", dbPath+dsn)
	if err != nil {
		return err
	}

	validate = validator.New()
	return createTables(Connection)
}

// Close closes the connection.
func Close() {
	if err := Connection.Close(); err != nil {
		log.Printf("failed to close database: %v", err)
	}
}
