package grepbook_test

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/ejamesc/grepbook"
	"github.com/kardianos/osext"
)

var testDB *grepbook.DB
var user1 *grepbook.User
var bookReview1 *grepbook.BookReview

// TestMain sets up the entire suite
func TestMain(m *testing.M) {
	// Global db setup
	pwd, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatalf("cannot retrieve present working directory: %s", err)
	}
	db, err := bolt.Open(path.Join(pwd, "test.db"), 0600, nil)
	if err != nil {
		log.Fatalf("unable to open bolt db: %s", err)
	}

	testDB = &grepbook.DB{DB: db}
	err = testDB.CreateAllBuckets()
	if err != nil {
		log.Fatalf("unable to create all buckets: %s", err)
	}

	// Consider this a test for DoesAnyUserExist
	exists := testDB.DoesAnyUserExist()
	if exists {
		log.Fatal("some users currently exist before the test begins")
	}

	user1, _ = testDB.CreateUser("test@test.com", "test")
	chapters := grepbook.CreateChapters("Introduction, Prelude, Conclusion")
	bookReview1, _ = testDB.CreateBookReview(
		"The Inner Game of Tennis",
		"W. Timothy Gallwey",
		"https://www.amazon.com/Inner-Game-Tennis-Classic-Performance/dp/0679778314/ref=sr_1_1?s=books&ie=UTF8&qid=1477563421&sr=1-1&keywords=inner+game+of+tennis",
		"<p>Great book!</p>",
		"{}", chapters)

	retCode := m.Run()

	db.Close()
	os.Exit(retCode)
}

// TEST HELPERS

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
