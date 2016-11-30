package grepbook_test

import (
	"sort"
	"testing"
	"time"

	"github.com/ejamesc/grepbook"
)

func TestCreateBookReview(t *testing.T) {
	chapters := grepbook.CreateChapters("Introduction, Preface")
	br, err := testDB.CreateBookReview(
		"Superintelligence",
		"Nick Bostrom",
		"https://www.amazon.com/Superintelligence-Dangers-Strategies-Nick-Bostrom/dp/1501227742",
		"<p>Hello</p>",
		"{}",
		chapters)
	ok(t, err)
	defer testDB.DeleteBookReview(br.UID)

	assert(t, br.UID != "", "expect uid to be filled with string")
	assert(t, br.Title != "", "expect book review title to be filled with string")
	assert(t, !br.DateTimeCreated.Equal(time.Time{}), "expect book review date created to be non-zero")
	assert(t, !br.DateTimeUpdated.Equal(time.Time{}), "expect book review date created to be non-zero")
	assert(t, br.BookAuthor != "", "expect book author to be filled with string")
	assert(t, br.IsOngoing, "expect IsOngoing to be true")
	assert(t, br.OverviewHTML == "<p>Hello</p>", "expect HTML to be empty")
	equals(t, 2, len(br.Chapters))
}

func TestGetBookReview(t *testing.T) {
	br, err := testDB.GetBookReview(bookReview1.UID)
	ok(t, err)
	equals(t, bookReview1, br)
}

func TestDeleteBookReview(t *testing.T) {
	chapters := grepbook.CreateChapters("Introduction, Preface")
	br, err := testDB.CreateBookReview(
		"Superintelligence",
		"Nick Bostrom",
		"https://www.amazon.com/Superintelligence-Dangers-Strategies-Nick-Bostrom/dp/1501227742",
		"<p>Hello</p>",
		"{}", chapters)
	ok(t, err)

	err = testDB.DeleteBookReview(br.UID)
	ok(t, err)

	br2, err := testDB.GetBookReview(br.UID)
	assert(t, err == grepbook.ErrNoRows, "expect error to be ErrNoRows because already deleted, but isn't")
	assert(t, br2 == nil, "expect bookreview to be nil because already deleted, but wasn't")
}

func TestBookReviewSave(t *testing.T) {
	delta := "<p>Stupid siaaaa</p>"
	originalTime := bookReview1.DateTimeUpdated

	bookReview1.OverviewHTML = delta
	err := bookReview1.Save(testDB)
	ok(t, err)

	br, err := testDB.GetBookReview(bookReview1.UID)
	ok(t, err)
	equals(t, delta, br.OverviewHTML)
	assert(t, br.DateTimeUpdated.After(originalTime), "expect DateTimeUpdated of book review to have been updated")
}

func TestBookReviewGetAll(t *testing.T) {
	bra, err := testDB.GetAllBookReviews()
	ok(t, err)
	equals(t, 1, len(bra))
}

func TestBookReviewSort(t *testing.T) {
	bra := grepbook.BookReviewArray{
		&grepbook.BookReview{DateTimeCreated: time.Now()},
		&grepbook.BookReview{DateTimeCreated: time.Now().Add(time.Hour)},
		bookReview1}
	sort.Sort(bra)
	for i, _ := range bra {
		if i+1 < len(bra) {
			assert(t, bra[i].DateTimeCreated.Before(bra[i+1].DateTimeCreated), "expect DateTimeCreated to be ascending")
		}
	}
}
