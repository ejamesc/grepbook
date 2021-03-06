package main_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/ejamesc/grepbook"
)

type MockBookReviewDB struct {
	shouldFail bool
}

func (db *MockBookReviewDB) CreateBookReview(title, author, bookURL, html, delta string, chapters []*grepbook.Chapter) (*grepbook.BookReview, error) {

	if db.shouldFail {
		return nil, fmt.Errorf("some error")
	}
	now := grepbook.TimeNow()
	return &grepbook.BookReview{
		UID:             "giEa2JTKrWEbTy2nb3U5wc",
		Title:           title,
		BookAuthor:      author,
		OverviewHTML:    html,
		BookURL:         bookURL,
		Delta:           delta,
		DateTimeCreated: now,
		DateTimeUpdated: now,
		Chapters:        chapters,
	}, nil
}

func (db *MockBookReviewDB) GetBookReview(uid string) (*grepbook.BookReview, error) {
	if db.shouldFail {
		return nil, fmt.Errorf("some error")
	}
	if uid == "" {
		return nil, grepbook.ErrNoRows
	}
	return bookReview1, nil
}

func (db *MockBookReviewDB) DeleteBookReview(uid string) error {
	if db.shouldFail {
		return fmt.Errorf("some error")
	}
	if uid == "" {
		return grepbook.ErrNoRows
	}
	return nil
}

func (db *MockBookReviewDB) GetAllBookReviews() (grepbook.BookReviewArray, error) {
	if db.shouldFail {
		return nil, fmt.Errorf("some error")
	}
	return grepbook.BookReviewArray{bookReview1}, nil
}

func (db *MockBookReviewDB) Update(func(tx *bolt.Tx) error) error {
	if db.shouldFail {
		return fmt.Errorf("some error")
	}
	return nil
}

func TestIndexHandler(t *testing.T) {
	mockDB := &MockBookReviewDB{shouldFail: false}
	indexHandler := app.IndexHandler(mockDB)
	test := GenerateHandleTester(t, app.Wrap(indexHandler), false)
	w := test("GET", url.Values{})
	assert(t, http.StatusOK == w.Code, "expected index page to return 200 instead got %d", w.Code)
}
