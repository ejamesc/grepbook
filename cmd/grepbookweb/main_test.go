package main_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ejamesc/grepbook"
	"github.com/ejamesc/grepbook/cmd/grepbookweb"
	"github.com/gorilla/context"
	"github.com/spf13/viper"
)

var bookReview1 = &grepbook.BookReview{
	Title:           "War and Peace",
	BookAuthor:      "Leo Tolstoy",
	HTML:            "<p>Great book!</p>",
	BookURL:         "https://www.amazon.com/Superintelligence-Dangers-Strategies-Nick-Bostrom/dp/1501227742",
	Delta:           "{}",
	DateTimeCreated: time.Now().UTC(),
	DateTimeUpdated: time.Now().UTC(),
	Chapters:        []*grepbook.Chapter{},
}

var app *main.App

func TestMain(m *testing.M) {
	r := main.NewRouter()
	ml := &MockLogger{}
	err := main.LoadConfiguration("")
	if err != nil {
		log.Printf("error loading configuration file: %s", err)
	}
	templatePath := path.Join(viper.GetString("path"), "templates")
	app = main.SetupApp(r, ml, []byte("some-secret"), templatePath)

	retCode := m.Run()
	os.Exit(retCode)
}

type MockLogger struct{}

func (ml *MockLogger) Log(str string, v ...interface{}) {
	fmt.Printf("mockLogger: "+str+"\n", v...)
}

type HandleTester func(method string, params url.Values) *httptest.ResponseRecorder

// Given the current test runner and an http.Handler, generate a
// HandleTester which will test its given input against the
// handler.
func GenerateHandleTester(
	t *testing.T,
	handleFunc http.Handler,
	loggedIn bool,
) HandleTester {
	// Given a method type ("GET", "POST", etc) and
	// parameters, serve the response against the handler and
	// return the ResponseRecorder.
	return func(method string, params url.Values) *httptest.ResponseRecorder {
		req, err := http.NewRequest(
			method,
			"",
			strings.NewReader(params.Encode()),
		)
		ok(t, err)
		req.Header.Set(
			"Content-Type",
			"application/x-www-form-urlencoded; param=value",
		)
		w := httptest.NewRecorder()
		if loggedIn {
			context.Set(req, main.UserKeyName, user1)
		}
		handleFunc.ServeHTTP(w, req)
		return w
	}
}

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
