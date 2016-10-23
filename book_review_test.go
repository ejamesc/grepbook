package grepbook_test

import (
	"testing"

	"github.com/ejamesc/grepbook"
)

func TestCreateBookReview(t *testing.T) {
	chapters := grepbook.CreateChapter("Introduction, Preface")
	br, err := testDB.CreateBookReview("Superintelligence", "Nick Bostrom", "<p>Hello</p>", "{}", chapters)
	ok(t, err)
	assert(t, br.UID != "", "expect uid to be filled with string")
	assert(t, br.Title != "", "expect book review title to be filled with string")
	assert(t, br.BookAuthor != "", "expect book author to be filled with string")
	assert(t, br.HTML == "<p>Hello</p>", "expect HTML to be empty")
	equals(t, 2, len(br.Chapters))
}

func TestGetBookReview(t *testing.T) {
	br, err := testDB.GetBookReview(bookReview1.UID)
	ok(t, err)
	equals(t, bookReview1, br)
}

func TestDeleteBookReview(t *testing.T) {

}

func TestBookReviewSave(t *testing.T) {

}
