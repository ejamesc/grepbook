package grepbook_test

import (
	"fmt"
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
	for _, chap := range br.Chapters {
		assert(t, chap.ID != "", "expect chapter ID to not be empty")
		assert(t, chap.Heading != "", "expect chapter headings to not be empty")
		assert(t, chap.HTML == "", "expect chapter html to be empty")
		assert(t, chap.Delta == "", "expect chapter delta to be empty")
	}
}

func TestGetBookReview(t *testing.T) {
	br, err := testDB.GetBookReview(bookReview1.UID)
	ok(t, err)
	equals(t, bookReview1, br)
}

func TestDeleteBookReview(t *testing.T) {
	br, err := createTestBookReview("Introduction, Preface")
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

func TestBookReviewIsNew(t *testing.T) {
	br, err := createTestBookReview("")
	ok(t, err)
	equals(t, true, br.IsNew())

	equals(t, false, bookReview1.IsNew())
}

func TestAddChapter(t *testing.T) {
	br, err := createTestBookReview("")
	ok(t, err)

	err = br.AddChapter(testDB, grepbook.NewChapter("New Chapter", "", ""))
	ok(t, err)
	equals(t, 1, len(br.Chapters))

	br2, err := testDB.GetBookReview(br.UID)
	ok(t, err)
	equals(t, 1, len(br2.Chapters))
}

func TestGetChapter(t *testing.T) {
	br, err := createTestBookReview("Terrible World")
	ok(t, err)
	ct := br.Chapters[0]

	index, chapUnderTest := br.GetChapter(ct.ID)
	equals(t, ct, chapUnderTest)
	equals(t, 0, index)
}

func TestUpdateChapter(t *testing.T) {
	br, err := createTestBookReview("Intro, Outtro")
	ok(t, err)
	cp := br.Chapters[0]
	newTitle := "Preface"
	delta := grepbook.ChapterDelta{Heading: &newTitle}

	err = br.UpdateChapter(testDB, cp.ID, delta)
	ok(t, err)
	equals(t, "Preface", cp.Heading)
	equals(t, "", cp.HTML)
	equals(t, "", cp.Delta)

	br2, err := testDB.GetBookReview(br.UID)
	ok(t, err)
	_, cp2 := br2.GetChapter(cp.ID)
	equals(t, "Preface", cp2.Heading)
	equals(t, cp.HTML, cp2.HTML)
	equals(t, cp.Delta, cp2.Delta)
}

func TestReorderChapter(t *testing.T) {
	br, err := createTestBookReview("First chap, Second chap, Third chap, Fourth chap")
	ok(t, err)
	equals(t, 4, len(br.Chapters))

	err = br.ReorderChapter(testDB, 4, 0)
	assert(t, err != nil, "expect ReorderChapter to return error when index >= chapter len")

	// Table tests for reordering
	tables := []struct {
		Heading1 string
		Heading2 string
		oldIndex int
		newIndex int
	}{
		{"Third chap", "Fourth chap", 3, 2},
		{"Fourth chap", "Second chap", 1, 2},
		{"Fourth chap", "First chap", 0, 1},
		{"Second chap", "Third chap", 3, 0},
	}

	for _, tb := range tables {
		err = br.ReorderChapter(testDB, tb.oldIndex, tb.newIndex)
		// Uncomment this to visualise the run of the table test:
		// visualiseOrdering(br.Chapters)
		ok(t, err)
		equals(t, tb.Heading1, br.Chapters[tb.oldIndex].Heading)
		equals(t, tb.Heading2, br.Chapters[tb.newIndex].Heading)

		br2, err := testDB.GetBookReview(br.UID)
		ok(t, err)
		equals(t, tb.Heading1, br2.Chapters[tb.oldIndex].Heading)
		equals(t, tb.Heading2, br2.Chapters[tb.newIndex].Heading)
	}
}

func TestDeleteChapter(t *testing.T) {
	br, err := createTestBookReview("First chap, Second chap")
	ok(t, err)
	cp := br.Chapters[0]
	equals(t, 2, len(br.Chapters))

	err = br.DeleteChapter(testDB, cp.ID)
	ok(t, err)
	equals(t, 1, len(br.Chapters))

	br2, err := testDB.GetBookReview(br.UID)
	ok(t, err)
	equals(t, 1, len(br2.Chapters))

	err = br2.DeleteChapter(testDB, cp.ID)
	assert(t, err != nil, "expect error to return when deleting non-existing chapter in book review, instead got nil")
}

// Helpers

func createTestBookReview(chapStr string) (*grepbook.BookReview, error) {
	chapters := grepbook.CreateChapters(chapStr)
	return testDB.CreateBookReview(
		"Superintelligence",
		"Nick Bostrom",
		"https://www.amazon.com/Superintelligence-Dangers-Strategies-Nick-Bostrom/dp/1501227742",
		"",
		"",
		chapters)
}

// Small helper function to visualise ordering.
// To use, plunk into TestReorderChapter, and run with go test -v
func visualiseOrdering(chps []*grepbook.Chapter) {
	fmt.Printf("%+v\n", chps[0])
	fmt.Printf("%+v\n", chps[1])
	fmt.Printf("%+v\n", chps[2])
	fmt.Printf("%+v\n", chps[3])
	fmt.Println("====")
}
