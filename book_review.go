package grepbook

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/renstrom/shortuuid"
)

type BookReview struct {
	UID             string     `json:"uid"`
	Title           string     `json:"title"`
	BookAuthor      string     `json:"book_author"`
	HTML            string     `json:"html"`
	Delta           string     `json:"delta"`
	DateTimeCreated time.Time  `json:"date_created"`
	DateTimeUpdated time.Time  `json:"date_updated"`
	Chapters        []*Chapter `json:"chapters"`
}

type Chapter struct {
	Heading string `json:"heading"`
}

// CreateChapter takes an input string of chapter headings separated by commas,
// and returns a list of Chapters with headings
func CreateChapter(input string) []*Chapter {
	headings := strings.Split(input, ",")
	if len(headings) < 1 {
		return []*Chapter{}
	}
	res := make([]*Chapter, len(headings))
	for i, h := range headings {
		res[i] = &Chapter{h}
	}
	return res
}

func (db *DB) CreateBookReview(title, author, html, delta string, chapters []*Chapter) (*BookReview, error) {
	now := TimeNow()
	bookReview := &BookReview{
		Title:           title,
		BookAuthor:      author,
		HTML:            html,
		Delta:           delta,
		DateTimeCreated: now,
		DateTimeUpdated: now,
		Chapters:        chapters,
	}

	err := bookReview.Save(db)
	if err != nil {
		return nil, err
	}

	return bookReview, nil
}

func (db *DB) GetBookReview(uid string) (*BookReview, error) {
	var bookReview *BookReview
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(reviews_bucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(reviews_bucket))
		}

		brJSON := b.Get([]byte(uid))
		if brJSON == nil {
			return ErrNoRows
		}
		return json.Unmarshal(brJSON, &bookReview)
	})
	if err != nil {
		return nil, err
	}
	return bookReview, nil
}

func (db *DB) DeleteBookReview(uid string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(reviews_bucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(reviews_bucket))
		}
		return b.Delete([]byte(uid))
	})
	return err
}

type BookReviewDB interface {
	CreateBookReview(string, string, string, string, []Chapter) (*BookReview, error)
	GetBookReview(string) (*BookReview, error)
	DeleteBookReview(string) error
}

func (br *BookReview) Save(db *DB) error {
	if br.UID == "" {
		br.UID = shortuuid.New()
	} else {
		br.DateTimeUpdated = TimeNow()
	}

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(reviews_bucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(reviews_bucket))
		}

		rJSON, err := json.Marshal(br)
		if err != nil {
			return fmt.Errorf("error with marshalling book review struct: %s", err)
		}
		return b.Put([]byte(br.UID), rJSON)
	})
	if err != nil {
		return err
	}
	return nil
}
