package grepbook

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/boltdb/bolt"
	"golang.org/x/crypto/bcrypt"
)

const workFactor = 10

type User struct {
	ID       uint64 `json:"id"`
	Name     string `json:"string"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// save is a private function to save user details to the database
func (u *User) save(db *DB) error {
	isCreate := u.ID == 0
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(users_bucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(users_bucket))
		}

		// If this is user creation, we handle some special cases
		if isCreate {
			val := b.Get([]byte(u.Email))
			if val != nil {
				return ErrDuplicateRow
			}

			id, err := b.NextSequence()
			if err != nil {
				return err
			}
			u.ID = id
		}

		usrJSON, err := json.Marshal(u)
		if err != nil {
			return fmt.Errorf("error with marshalling user object: %s", err)
		}
		return b.Put([]byte(u.Email), usrJSON)
	})
}

// UserDelta is a struct to contain user details for updating
type UserDelta struct {
	Email    string
	Name     string
	Password string
}

// UpdateUser updates the user with the given email address.
// Note that we don't provide an update method on the user object,
// in order to handle all the scenarios that may arise from this.
func (db *DB) UpdateUser(userEmail string, ud UserDelta) (*User, error) {
	name, email := strings.TrimSpace(ud.Name), strings.TrimSpace(ud.Email)
	password := ud.Password

	user, err := db.getFullUser(userEmail)
	if err != nil {
		return nil, err
	}

	if password != "" {
		hashedP, err := bcrypt.GenerateFromPassword([]byte(password), workFactor)
		if err != nil {
			return nil, fmt.Errorf("error generating bcrypt hash: ", err)
		}
		user.Password = string(hashedP)
	}

	if name != "" {
		user.Name = name
	}

	if email != "" && email != user.Email {
		newU := &User{
			Email:    email,
			Name:     user.Name,
			Password: user.Password,
		}
		err = newU.save(db)
		if err != nil {
			return nil, err
		}
		err := db.DeleteUser(user.Email)
		if err != nil {
			return nil, err
		}
		newU.Password = ""
		return newU, nil
	} else {
		err = user.save(db)
		if err != nil {
			return nil, err
		}
		user.Password = ""
		return user, nil
	}
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

// CreateUser creates a user.
// It expects a valid email and password.
// It also returns an error if a duplicate user is found.
func (db *DB) CreateUser(email, password string) (*User, error) {
	user := &User{}
	password = strings.TrimSpace(password)
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}
	if !govalidator.IsEmail(email) {
		return nil, fmt.Errorf("email is not a valid email address")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), workFactor)
	if err != nil {
		return nil, err
	}

	user = &User{Email: email, Password: string(hashedPassword)}
	err = user.save(db)

	if err != nil {
		return nil, err
	}
	// we clear out the password
	user.Password = ""
	return user, nil
}

// GetUser returns a user. If no user exists, a grepbook.ErrNoRows error is returned.
func (db *DB) GetUser(email string) (*User, error) {
	u, err := db.getFullUser(email)
	if err != nil {
		return nil, err
	}
	u.Password = ""
	return u, nil
}

// GetName returns the username of the first (and usually only) user in the db
func (db *DB) GetName() (string, error) {
	username := ""
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(users_bucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(users_bucket))
		}

		b.ForEach(func(k, v []byte) error {
			var user User
			err := json.Unmarshal(v, &user)
			if err != nil {
				return err
			}
			username = user.Name
			return nil
		})
		return nil
	})
	if err != nil {
		return "", err
	}
	return username, nil
}

func (db *DB) getFullUser(email string) (*User, error) {
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
	CreateUser(email, password string) (*User, error)
	GetUser(email string) (*User, error)
	UpdateUser(userEmail string, ud UserDelta) (*User, error)
	IsUserPasswordCorrect(email, password string) bool
	DeleteUser(email string) error
	CreateSessionForUser(email string) (*Session, error)
}
