// Package db stores and manipulates dotfiles via a sqlite3 database.
package db

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/knoebber/dotfile/file"
	"github.com/knoebber/dotfile/usererror"
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
)

var (
	connection *sql.DB
	validate   *validator.Validate
)

type tableCreator interface {
	createStmt() string
}

type executor interface {
	Exec(string, ...interface{}) (sql.Result, error)
}

type inserter interface {
	insertStmt(executor) (sql.Result, error)
}

type checker interface {
	check() error
}

// Creates the required tables if they doesn't exist
func createTables() error {
	for _, model := range []tableCreator{
		new(UserRecord),
		new(ReservedUsernameRecord),
		new(SessionRecord),
		new(FileRecord),
		new(TempFileRecord),
		new(CommitRecord),
	} {
		_, err := connection.Exec(model.createStmt())
		if err != nil {
			return errors.Wrap(err, "creating tables")
		}
	}
	return nil
}

func insert(i inserter, tx *sql.Tx) (id int64, err error) {
	handleErr := func(err error) error {
		if tx != nil {
			return rollback(tx, err)
		}
		return err
	}

	var res sql.Result

	if err = validate.Struct(i); err != nil {
		log.Print(err)
		invalidError := usererror.Invalid("Values are missing or improperly formatted.")
		return 0, handleErr(invalidError)
	}

	if c, ok := i.(checker); ok {
		if err := c.check(); err != nil {
			return 0, handleErr(err)
		}
	}

	if tx != nil {
		res, err = i.insertStmt(tx)
	} else {
		res, err = i.insertStmt(connection)
	}
	if err != nil {
		return 0, handleErr(err)
	}

	id, err = res.LastInsertId()

	if err != nil {
		return 0, handleErr(err)
	}

	return id, nil
}

// Rolls back a database transaction.
// It always returns an error.
func rollback(tx *sql.Tx, err error) error {
	if rbError := tx.Rollback(); rbError != nil {
		return errors.Wrapf(err, "rolling back database transaction: %s", rbError)
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
	if err := file.CheckAlias(alias); err != nil {
		return err
	}

	if err := file.CheckPath(path); err != nil {
		return err
	}
	return nil
}

// NotFound returns whether err is wrapping a no rows error.
func NotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

// Start opens a connection a sqlite3 database.
// Creates a new sqlite database with all required tables when not found.
func Start(dbPath string) (err error) {
	dsn := "?_foreign_keys=true"
	connection, err = sql.Open("sqlite3", dbPath+dsn)
	if err != nil {
		return err
	}

	validate = validator.New()
	return createTables()
}

// Close closes the connection.
func Close() {
	if err := connection.Close(); err != nil {
		log.Printf("failed to close database: %v", err)
	}
}
