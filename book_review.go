package grepbook

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/renstrom/shortuuid"
)

type BookReview struct {
	UID        string     `json:"uid"`
	Title      string     `json:"title"`
	BookAuthor string     `json:"book_author"`
	HTML       string     `json:"html"`
	Delta      string     `json:"delta"`
	Chapters   []*Chapter `json:"chapters"`
}

type Chapter struct {
	Heading string `json:"heading"`
}

// CreateChapter takes an input string of chapter headings separated by commas,
// and returns a list of Chapters with headings
func CreateChapter(input string) []*Chapter {
	headings := strings.Split(input, ",")
	res := make([]*Chapter, len(headings))
	for i, h := range headings {
		res[i] = &Chapter{h}
	}
	return res
}

func (db *DB) CreateBookReview(title, author, html, delta string, chapters []*Chapter) (*BookReview, error) {
	bookReview := &BookReview{
		Title:      title,
		BookAuthor: author,
		HTML:       html,
		Delta:      delta,
		Chapters:   chapters,
	}

	err := bookReview.Save(db)
	if err != nil {
		return nil, err
	}

	return bookReview, nil
}

func (db *DB) GetBookReview(uid string) (*BookReview, error) {
	return &BookReview{}, nil
}

func (db *DB) DeleteBookReview(uid string) error {
	return nil
}

type BookReviewDB interface {
	CreateBookReview(string, string, string, string, []Chapter) (*BookReview, error)
	GetBookReview(string) (*BookReview, error)
	DeleteBookReview(string) error
}

func (br *BookReview) Save(db *DB) error {
	if br.UID == "" {
		br.UID = shortuuid.New()
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
