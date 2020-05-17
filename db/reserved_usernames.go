package db

import (
	"fmt"
	"time"

	"database/sql"
	"github.com/knoebber/dotfile/usererr"
	"github.com/pkg/errors"
	"strings"
)

// ReservedUsername stores usernames that are not allowed to be registered.
type ReservedUsername struct {
	Username  string
	CreatedAt time.Time
}

func (*ReservedUsername) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS reserved_usernames(
id              INTEGER PRIMARY KEY,
username        TEXT NOT NULL UNIQUE,
created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
}

func (ru *ReservedUsername) insertStmt(e executor) (sql.Result, error) {
	return e.Exec("INSERT INTO reserved_usernames(username)", ru.Username)
}

func checkUsernameAllowed(username string) error {
	var count int

	err := connection.
		QueryRow("SELECT COUNT(*) FROM reserved_usernames WHERE username = ?", username).
		Scan(&count)
	if err != nil {
		return errors.Wrapf(err, "checking if %#v is reserved", username)
	}

	if count > 0 {
		return usererr.Invalid(fmt.Sprintf("Username %#v is reserved.", username))
	}

	return nil
}

// SeedReservedUsernames sets usernames which are not allowed to be used.
// This should be called when the service is started.
func SeedReservedUsernames(usernames []interface{}) error {
	var count int64

	placeholders := "(?)" + strings.Repeat(",(?)", len(usernames)-1)

	whereIn := fmt.Sprintf("WHERE username IN (%s)", placeholders)
	err := connection.
		QueryRow("SELECT COUNT(*) FROM users "+whereIn, usernames...).Scan(&count)

	if err != nil {
		return errors.Wrap(err, "checking if reserved usernames exist")
	}
	if count > 0 {
		return errors.New("reserved username exists in the user table")
	}

	sql := fmt.Sprintf(`
INSERT INTO reserved_usernames (username) 
VALUES %s 
ON CONFLICT DO NOTHING`, placeholders)

	if _, err = connection.Exec(sql, usernames...); err != nil {
		return errors.Wrap(err, "seeding reserved usernames")
	}

	return nil
}
