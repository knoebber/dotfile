package db

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/knoebber/dotfile/usererr"
	"github.com/pkg/errors"
)

const (
	timestampDisplayFormat = "January 02, 2006"

	// Some safe guards against abusing uploads.
	// Total ~150 megabytes of storage per user.
	// Making accounts is currently trivial however.
	// If it becomes a problem the next step is requiring email verification.
	maxFilesPerUser   = 100
	maxCommitsPerFile = 100
	maxBlobSizeBytes  = 15000
)

func checkSize(content []byte, name string) error {
	if len(content) > maxBlobSizeBytes {
		return usererr.Invalid(fmt.Sprintf("%s is too large (max=%dKB)", name, maxBlobSizeBytes/1000))
	}
	return nil

}

// NotFound returns whether err is wrapping a no rows error.
func NotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

type executor interface {
	Exec(string, ...interface{}) (sql.Result, error)
}

type inserter interface {
	insertStmt(executor) (sql.Result, error)
}

func insert(i inserter, tx *sql.Tx) (id int64, err error) {
	var res sql.Result

	if err = validate.Struct(i); err != nil {
		return 0, err
	}

	if tx != nil {
		res, err = i.insertStmt(tx)
	} else {
		res, err = i.insertStmt(connection)
	}

	if err != nil && tx != nil {
		return 0, rollback(tx, err)
	} else if err != nil {
		return 0, err
	}

	id, err = res.LastInsertId()

	if err != nil {
		return 0, err
	}

	return id, nil
}

// Rolls back a database transaction.
// It always returns an error.
func rollback(tx *sql.Tx, err error) error {
	if rbError := tx.Rollback(); rbError != nil {
		return errors.Wrap(rbError, "rolling back database transaction")
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

func formatTime(t time.Time) string {
	return t.Format(timestampDisplayFormat)

}
