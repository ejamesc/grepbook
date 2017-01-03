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

func (br BookReview) IsNew() bool {
	isNew := strings.TrimSpace(br.OverviewHTML) == ""
	if !isNew {
		return false
	}
	for _, chap := range br.Chapters {
		if strings.TrimSpace(chap.HTML) != "" {
			return false
		}
	}
	return true
}

// NewChapter creates a new chapter given a heading
func NewChapter(heading, delta, html string) *Chapter {
	delta, html = strings.TrimSpace(delta), strings.TrimSpace(html)
	return &Chapter{ID: shortuuid.New(), Heading: heading, Delta: delta, HTML: html}
}

// AddChapter prepends a chapter to the list of chapters in the BookReview.
func (br *BookReview) AddChapter(db BookReviewDB, chapter *Chapter) error {
	br.Chapters = append(br.Chapters, &Chapter{})
	copy(br.Chapters[1:], br.Chapters[0:])
	br.Chapters[0] = chapter

	return br.Save(db)
}

// GetChapter returns the chapter and index in the BookReview with the id given.
// If no such chapter id exists, -1 index and nil is returned.
func (br *BookReview) GetChapter(id string) (index int, chapter *Chapter) {
	for i, c := range br.Chapters {
		if c.ID == id {
			return i, c
		}
	}
	return -1, nil
}

// ChapterDelta is a struct for storing changes to a chapter
type ChapterDelta struct {
	Heading *string
	HTML    *string
	Delta   *string
}

// UpdateChapter updates the chapter given.
func (br *BookReview) UpdateChapter(db BookReviewDB, chapID string, cd ChapterDelta) error {
	i, cp := br.GetChapter(chapID)
	if cd.Heading != nil {
		cp.Heading = *cd.Heading
	}
	if cd.HTML != nil {
		cp.HTML = *cd.HTML
	}
	if cd.Delta != nil {
		cp.Delta = *cd.Delta
	}
	br.Chapters[i] = cp
	return br.Save(db)
}

func (br *BookReview) ReorderChapter(db BookReviewDB, oldIndex, newIndex int) error {
	if oldIndex >= len(br.Chapters) || newIndex >= len(br.Chapters) {
		return fmt.Errorf("either oldIndex (%d) or newIndex (%d) is out of bounds for Chapter list of len %d", oldIndex, newIndex, len(br.Chapters))
	}
	cp := br.Chapters[oldIndex]
	// Delete
	copy(br.Chapters[oldIndex:], br.Chapters[oldIndex+1:])
	br.Chapters[len(br.Chapters)-1] = &Chapter{}
	br.Chapters = br.Chapters[:len(br.Chapters)-1]

	// Insert
	br.Chapters = append(br.Chapters, &Chapter{})
	copy(br.Chapters[newIndex+1:], br.Chapters[newIndex:])
	br.Chapters[newIndex] = cp

	return br.Save(db)
}

// DeleteChapter deletes a chapter with the given chapter ID.
// If no such chapter exists, an error is returned
func (br *BookReview) DeleteChapter(db BookReviewDB, chapID string) error {
	i, _ := br.GetChapter(chapID)
	if i == -1 {
		return fmt.Errorf("no such chapter id found")
	}

	copy(br.Chapters[i:], br.Chapters[i+1:])
	br.Chapters[len(br.Chapters)-1] = &Chapter{}
	br.Chapters = br.Chapters[:len(br.Chapters)-1]

	return br.Save(db)
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
	input = strings.TrimSpace(input)
	if input == "" {
		return []*Chapter{}
	}
	headings := strings.Split(input, ",")
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
