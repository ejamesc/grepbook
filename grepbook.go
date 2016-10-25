package grepbook

import (
	"errors"
	"time"

	"github.com/boltdb/bolt"
)

// This list of variables are a indication of what the list of buckets are in our db.
var users_bucket = []byte("users")
var reviews_bucket = []byte("book_reviews")
var sessions_bucket = []byte("sessions")
var buckets_list = [][]byte{users_bucket, reviews_bucket, sessions_bucket}

// Errors
var ErrNoRows = errors.New("db: no rows in result set")
var ErrDuplicateRow = errors.New("db: duplicate row found for unique constraint")

// Wrapper for bolt db. This allows us to attach methods
// to the db object.
type DB struct {
	*bolt.DB
}

func (db *DB) CreateAllBuckets() error {
	err := db.Update(func(tx *bolt.Tx) error {
		for _, b := range buckets_list {
			_, err := tx.CreateBucketIfNotExists(b)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func TimeNow() time.Time {
	return time.Now().UTC()
}
