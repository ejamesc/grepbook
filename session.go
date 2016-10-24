package grepbook

import (
	"encoding/json"
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/boltdb/bolt"
	"github.com/renstrom/shortuuid"
)

type Session struct {
	Key   string `json:"key"`
	Email string `json:"email"`
}

// CreateSessionForUser creates a new session for a user.
// One user can have many sessions.
// Only valid email addresses are accepted
func (db *DB) CreateSessionForUser(email string) (*Session, error) {
	if !govalidator.IsEmail(email) {
		return nil, fmt.Errorf("email is not a valid email address")
	}
	session := &Session{Email: email, Key: shortuuid.New()}
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessions_bucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(sessions_bucket))
		}

		sessJSON, err := json.Marshal(session)
		if err != nil {
			return err
		}
		return b.Put([]byte(session.Key), sessJSON)
	})
	if err != nil {
		return nil, err
	}
	return session, nil
}

// GetUserBySessionKey returns a user based on the session.
func (db *DB) GetUserBySessionKey(ssk string) (*User, error) {
	var session Session
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessions_bucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(sessions_bucket))
		}
		sessJSON := b.Get([]byte(ssk))
		if sessJSON == nil {
			return ErrNoRows
		}
		return json.Unmarshal(sessJSON, &session)
	})
	if err != nil {
		return nil, err
	}
	return db.GetUser(session.Email)
}

// DeleteSession deletes a session. If no session is deleted, nothing happens
// and we return a nil error.
func (db *DB) DeleteSession(ssk string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessions_bucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(sessions_bucket))
		}
		return b.Delete([]byte(ssk))
	})
	return err
}

type SessionDB interface {
	GetUserBySessionKey(string) (*User, error)
	CreateSessionForUser(string) (*Session, error)
	DeleteSession(string) error
}
