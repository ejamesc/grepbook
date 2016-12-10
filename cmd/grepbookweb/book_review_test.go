package main_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ejamesc/grepbook/cmd/grepbookweb"
	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
)

func TestReadHandler(t *testing.T) {
	mockDB := &MockBookReviewDB{shouldFail: false}
	rh := app.Wrap(app.ReadHandler(mockDB))
	req, err := http.NewRequest("GET", "", nil)
	ok(t, err)
	w := httptest.NewRecorder()

	context.Set(req, main.Params, httprouter.Params{})

	rh.ServeHTTP(w, req)
	assert(t, w.Code == http.StatusOK, "expected read handler to return 200, instead got %d", w.Code)
	assert(t, !strings.Contains(w.Body.String(), "editor"), "expect non logged in user of read handler to not be able to view editor, but editor was shown")

	context.Set(req, main.UserKeyName, user1)
	rh.ServeHTTP(w, req)
	assert(t, w.Code == http.StatusOK, "expected read handler to return 200, instead got %d", w.Code)
	assert(t, strings.Contains(w.Body.String(), "editor"), "expect non logged in user of read handler to not be able to view editor, but editor was shown")
}

func TestCreateBookReviewHandler(t *testing.T) {
	// Test success scenario
	mockDB := &MockBookReviewDB{shouldFail: false}
	createBookHandler := app.CreateBookReviewHandler(mockDB)
	test := GenerateHandleTester(t, app.Wrap(createBookHandler), true)
	w := test("POST", url.Values{"title": {"Superintelligence"}})
	assert(t, w.Code == http.StatusFound, "expected create book review to redirect on success, instead got %d", w.Code)
	assert(t,
		len(w.HeaderMap["Location"]) > 0 && strings.Contains(w.HeaderMap["Location"][0], "/summaries/"),
		"expected redirect location on successful signup to be /summaries/:uid instead got %s", w.HeaderMap["Location"])

	// Test empty title
	w = test("POST", url.Values{})
	assert(t, w.Code == http.StatusBadRequest, "expected create book review to return 400 bad request error on empty title field, instead got %d", w.Code)
}

func TestUpdateBookReviewHandler(t *testing.T) {
	//mockDB := &MockBookReviewDB{shouldFail: false}
	//updateBookHandler := app.UpdateBookReviewHandler(mockDB)
	//test := GenerateHandleTester(t, app.Wrap(updateBookHandler), true)
	//w := test("PUT", url.Values{""})
}
