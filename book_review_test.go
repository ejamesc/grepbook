package grepbook_test

import (
	"testing"

	"github.com/ejamesc/grepbook"
)

func TestCreateBookReview(t *testing.T) {
	chapters := grepbook.CreateChapter("Introduction, Preface")
	br, err := testDB.CreateBookReview("Superintelligence", "Nick Bostrom", "<p>Hello</p>", "{}", chapters)
	ok(t, err)
	defer testDB.DeleteBookReview(br.UID)

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
	chapters := grepbook.CreateChapter("Introduction, Preface")
	br, err := testDB.CreateBookReview("Superintelligence", "Nick Bostrom", "<p>Hello</p>", "{}", chapters)
	ok(t, err)

	err = testDB.DeleteBookReview(br.UID)
	ok(t, err)

	br2, err := testDB.GetBookReview(br.UID)
	assert(t, err == grepbook.ErrNoRows, "expect error to be ErrNoRows because already deleted, but isn't")
	assert(t, br2 == nil, "expect bookreview to be nil because already deleted, but wasn't")
}

func TestBookReviewSave(t *testing.T) {

}
