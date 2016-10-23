package grepbook

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       uint64 `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// DoesAnyUserExist returns true if any users exists in the db, and
// false otherwise.
func (db *DB) DoesAnyUserExist() bool {
	res := false
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(users_bucket)
		if b != nil {
			stats := b.Stats()
			res = stats.KeyN > 0
		}

		return nil
	})
	if err != nil {
		return false
	}

	return res
}

// CreateUser creates a user. No email and password validation is done here.
// It expects a valid email and password.
// It also returns an error if a duplicate user is found.
func (db *DB) CreateUser(email, password string) (*User, error) {
	user := &User{}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(users_bucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(users_bucket))
		}
		val := b.Get([]byte(email))
		if val != nil {
			return ErrDuplicateRow
		}

		id, err := b.NextSequence()
		if err != nil {
			return err
		}

		user = &User{ID: id, Email: email, Password: string(hashedPassword)}
		userJSON, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("error with marshalling new user object: %s", err)
		}
		return b.Put([]byte(email), userJSON)
	})
	if err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
}

// GetUser returns a user. If no user exists, a grepbook.ErrNoRows error is returned.
func (db *DB) GetUser(email string) (*User, error) {
	var user User
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(users_bucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(users_bucket))
		}
		userJSON := b.Get([]byte(email))
		if userJSON == nil {
			return ErrNoRows
		}
		return json.Unmarshal(userJSON, &user)
	})
	if err != nil {
		return nil, err
	}
	user.Password = ""
	return &user, nil
}

// DeleteUser deletes a user. If no user is deleted, nothing happens
// and we return a nil error.
func (db *DB) DeleteUser(email string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(users_bucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(users_bucket))
		}
		return b.Delete([]byte(email))
	})
	return err
}

// IsUserPasswordCorrect returns true if the password for the user with email
// is correct.
func (db *DB) IsUserPasswordCorrect(email, password string) bool {
	var user User
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(users_bucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(users_bucket))
		}
		userJSON := b.Get([]byte(email))
		if userJSON == nil {
			return ErrNoRows
		}
		return json.Unmarshal(userJSON, &user)
	})
	if err != nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return (err == nil)
}

type UserDB interface {
	DoesAnyUserExist() bool
	CreateUser(string, string) (*User, error)
	GetUser(string) (*User, error)
	DeleteUser(string) error
	IsUserPasswordCorrect(string, string) bool
}
