package db

import (
	"database/sql"
	"encoding/base64"
	"time"

	"github.com/pkg/errors"
)

const sessionLength = 24

// SessionRecord models the sessions table.
// It tracks a user's active sessions.
type SessionRecord struct {
	ID        int64
	Session   string `validate:"required"`
	UserID    int64  `validate:"required"`
	Username  string
	Theme     UserTheme
	IP        string
	CreatedAt time.Time
	DeletedAt *time.Time
}

func (*SessionRecord) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS sessions(
id         INTEGER PRIMARY KEY,
session    TEXT NOT NULL UNIQUE,
user_id    INTEGER NOT NULL REFERENCES users,
ip         TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
deleted_at DATETIME
);

CREATE INDEX IF NOT EXISTS sessions_user_index ON sessions(user_id);`
}

func (s *SessionRecord) insertStmt(e executor) (sql.Result, error) {
	return e.Exec("INSERT INTO sessions(session, user_id, ip) VALUES(?, ?, ?)", s.Session, s.UserID, s.IP)
}

func createSession(username, ip string) (*SessionRecord, error) {
	var userID int64

	err := connection.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		return nil, errors.Wrapf(err, "querying for userID from %#v", username)
	}

	session, err := session()
	if err != nil {
		return nil, err
	}

	s := &SessionRecord{
		Session: session,
		UserID:  userID,
		IP:      ip,
	}

	id, err := insert(s, nil)
	if err != nil {
		return nil, err
	}

	s.ID = id

	return s, nil
}

func session() (string, error) {
	buff, err := randomBytes(sessionLength)
	if err != nil {
		return "", errors.Wrap(err, "generating session")
	}

	return base64.URLEncoding.EncodeToString(buff), nil
}

// Session returns a valid session.
func Session(session string) (*SessionRecord, error) {
	s := new(SessionRecord)

	err := connection.
		QueryRow(`
SELECT sessions.id,
       session,
       users.id,
       username,
       theme,
       sessions.created_at
FROM sessions
JOIN users ON users.id = user_id
WHERE deleted_at IS NULL AND session = ?`, session).
		Scan(
			&s.ID,
			&s.Session,
			&s.UserID,
			&s.Username,
			&s.Theme,
			&s.CreatedAt,
		)

	if err != nil {
		return nil, errors.Wrapf(err, "querying for session %#v", session)
	}

	return s, nil
}

// Logout sets the session to deleted.
func Logout(sessionID int64) error {
	_, err := connection.Exec("UPDATE sessions SET deleted_at = ? WHERE id = ?", time.Now(), sessionID)
	if err != nil {
		return errors.Wrapf(err, "setting session %d to deleted", sessionID)
	}

	return nil
}
