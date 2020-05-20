package db

import (
	"database/sql"
	"encoding/base64"
	"time"

	"github.com/pkg/errors"
)

const (
	sessionLength = 24
	sessionQuery  = `
SELECT sessions.id,
       session,
       users.id,
       username,
       sessions.created_at,
       session_locations.ip
FROM sessions
JOIN users ON users.id = user_id 
LEFT JOIN session_locations ON session_id = sessions.id AND last = 1
`
)

// Session is the model for the sessions table.
// It tracks a user's active sessions.
type Session struct {
	ID        int64
	Session   string `validate:"required"`
	UserID    int64  `validate:"required"`
	Username  string
	LastIP    *string
	CreatedAt time.Time
	DeletedAt *time.Time
}

func (*Session) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS sessions(
id         INTEGER PRIMARY KEY,
session    TEXT NOT NULL UNIQUE,
user_id    INTEGER NOT NULL REFERENCES users,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
deleted_at DATETIME
);

CREATE INDEX IF NOT EXISTS sessions_user_index ON sessions(user_id);`
}

func (s *Session) insertStmt(e executor) (sql.Result, error) {
	return e.Exec("INSERT INTO sessions(session, user_id) VALUES(?, ?)", s.Session, s.UserID)
}

func createSession(username string) (*Session, error) {
	var userID int64

	err := connection.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		return nil, errors.Wrapf(err, "querying for userID from %#v", username)
	}

	session, err := session()
	if err != nil {
		return nil, err
	}

	s := &Session{
		Session: session,
		UserID:  userID,
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

// CheckSession checks if session exists, and adds a new row to session_locations if the IP is new.
func CheckSession(session, ip string) (*Session, error) {
	s := new(Session)

	err := connection.
		QueryRow(sessionQuery+" WHERE deleted_at IS NULL AND session = ?", session).
		Scan(
			&s.ID,
			&s.Session,
			&s.UserID,
			&s.Username,
			&s.CreatedAt,
			&s.LastIP,
		)

	if err != nil {
		return nil, errors.Wrapf(err, "querying for session %#v", session)
	}

	if s.LastIP != nil && *s.LastIP == ip {
		return s, nil
	}

	if err = addSessionLocation(s.ID, ip); err != nil {
		return nil, err
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
