package main_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

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
		Title:           title,
		BookAuthor:      author,
		HTML:            html,
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
	return bookReview1, nil
}

func (db *MockBookReviewDB) DeleteBookReview(uid string) error {
	if db.shouldFail {
		return fmt.Errorf("some error")
	}
	return nil
}

func TestIndexHandler(t *testing.T) {
	mockDB := &MockBookReviewDB{shouldFail: true}
	indexHandler := app.IndexHandler(mockDB)
	test := GenerateHandleTester(t, app.Wrap(indexHandler))
	w := test("GET", url.Values{})
	assert(t, http.StatusOK == w.Code, "expected index page to return 200 instead got %d", w.Code)
}
