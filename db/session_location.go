package db

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// SessionLocation is the model for the session_location table.
// It tracks the different IP addresses that a session was used at.
type SessionLocation struct {
	ID        int64
	SessionID int64  `validate:"required"`
	IP        string `validate:"required"`
	CreatedAt time.Time
}

func (*SessionLocation) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS session_locations(
id         INTEGER PRIMARY KEY,
session_id INTEGER NOT NULL REFERENCES sessions,
ip         TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS session_locations_session_index ON session_locations(session_id);`
}

func (s *SessionLocation) insertStmt() (sql.Result, error) {
	return connection.Exec("INSERT INTO session_locations(session_id, ip) VALUES(?, ?)", s.SessionID, s.IP)
}

func addSessionLocation(sessionID int64, ip string) error {
	s := &SessionLocation{
		SessionID: sessionID,
		IP:        ip,
	}

	_, err := insert(s)
	if err != nil {
		return errors.Wrapf(err, "adding location %#v to session %d", ip, sessionID)
	}

	return nil
}
