package grepbook

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/renstrom/shortuuid"
)

type BookReview struct {
	UID             string     `json:"uid"`
	Title           string     `json:"title"`
	BookAuthor      string     `json:"book_author"`
	BookURL         string     `json:"book_url"`
	OverviewHTML    string     `json:"html"`
	Delta           string     `json:"delta"`
	DateTimeCreated time.Time  `json:"date_created"`
	DateTimeUpdated time.Time  `json:"date_updated"`
	IsOngoing       bool       `json:"is_ongoing"`
	Chapters        []*Chapter `json:"chapters"`
}

// Sorting BookReviewArray
type BookReviewArray []*BookReview

func (bra BookReviewArray) Len() int { return len(bra) }
func (bra BookReviewArray) Swap(i, j int) {
	bra[i], bra[j] = bra[j], bra[i]
}
func (bra BookReviewArray) Less(i, j int) bool {
	return bra[i].DateTimeCreated.Before(bra[j].DateTimeCreated)
}

type Chapter struct {
	ID      string `json:"id"`
	Heading string `json:"heading"`
	HTML    string `json:"html"`
	Delta   string `json:"delta"`
}

// CreateChapter takes an input string of chapter headings separated by commas,
// and returns a list of Chapters with headings
func CreateChapters(input string) []*Chapter {
	headings := strings.Split(input, ",")
	if len(headings) < 1 {
		return []*Chapter{}
	}
	res := make([]*Chapter, len(headings))
	for i, h := range headings {
		res[i] = &Chapter{ID: shortuuid.New(), Heading: strings.TrimSpace(h)}
	}
	return res
}

func (db *DB) CreateBookReview(title, author, bookURL, html, delta string, chapters []*Chapter) (*BookReview, error) {
	now := TimeNow()
	bookReview := &BookReview{
		Title:           title,
		BookAuthor:      author,
		BookURL:         bookURL,
		OverviewHTML:    html,
		Delta:           delta,
		DateTimeCreated: now,
		DateTimeUpdated: now,
		IsOngoing:       true,
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

// GetAllBookReview returns an array of all book reviews sorted by DateTimeCreated
func (db *DB) GetAllBookReviews() (BookReviewArray, error) {
	bra := BookReviewArray{}
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(reviews_bucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(reviews_bucket))
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			br, err := loadBookReviewFromJSON(v)
			if err != nil {
				return err
			}
			bra = append(bra, br)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Sort(bra)
	return bra, nil
}

type BookReviewDB interface {
	CreateBookReview(title, author, bookURL, html, delta string, chapters []*Chapter) (*BookReview, error)
	GetBookReview(uid string) (*BookReview, error)
	DeleteBookReview(uid string) error
	GetAllBookReviews() (BookReviewArray, error)
	Update(func(tx *bolt.Tx) error) error
}

func (br *BookReview) Save(db BookReviewDB) error {
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

func loadBookReviewFromJSON(jsonStr []byte) (*BookReview, error) {
	var br *BookReview
	err := json.Unmarshal(jsonStr, &br)
	if err != nil {
		return nil, err
	}
	return br, nil
}
