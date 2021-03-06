package main_test

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestReadHandler(t *testing.T) {
	mockDB := &MockBookReviewDB{shouldFail: false}
	rh := app.Wrap(app.ReadHandler(mockDB))
	params := httprouter.Params{httprouter.Param{Key: "id", Value: "someUUID"}}
	test := GenerateHandleTesterWithURLParams(t, rh, false, params)
	w := test("GET", url.Values{})

	assert(t, w.Code == http.StatusOK, "expected read handler to return 200, instead got %d", w.Code)

	test = GenerateHandleTesterWithURLParams(t, rh, true, params)
	w = test("GET", url.Values{})
	assert(t, w.Code == http.StatusOK, "expected read handler to return 200, instead got %d", w.Code)
}

func TestWritePageDisplayHandler(t *testing.T) {
	mockDB := &MockBookReviewDB{shouldFail: false}
	rh := app.Wrap(app.WritePageDisplayHandler(mockDB))
	params := httprouter.Params{httprouter.Param{Key: "id", Value: "someUUID"}}
	test := GenerateHandleTesterWithURLParams(t, rh, true, params)
	w := test("GET", url.Values{})

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
	assert(t, w.Code == http.StatusFound, "expected create book review to return 400 bad request error on empty title field, instead got %d", w.Code)
}

// Not going to write this test until after the shift has been done.
func TestUpdateBookReviewHandler(t *testing.T) {
	mockDB := &MockBookReviewDB{shouldFail: false}
	updateBookReviewHandler := app.UpdateBookReviewHandler(mockDB)
	params := httprouter.Params{httprouter.Param{Key: "id", Value: "someUUID"}}
	test := GenerateHandleJSONTesterWithURLParams(t, app.Wrap(updateBookReviewHandler), true, params)
	br, _ := mockDB.GetBookReview("someUUID")

	// Successful update
	jsonString := fmt.Sprintf(`{"uid": "%s"}`, br.UID)
	w := test("PUT", strings.NewReader(jsonString))
	equals(t, http.StatusOK, w.Code)

	// User attempts to update a story that does not belong to him
	w = test("PUT", strings.NewReader(fmt.Sprintf(`{"uid": "someOtherUUID"}`)))
	equals(t, http.StatusForbidden, w.Code)

	// Malformed json supplied
	w = test("PUT", strings.NewReader("LOL"))
	equals(t, http.StatusInternalServerError, w.Code)

	// User gives nonexistent uuid
	params = httprouter.Params{httprouter.Param{Key: "id", Value: ""}}
	test = GenerateHandleJSONTesterWithURLParams(t, app.Wrap(updateBookReviewHandler), true, params)
	w = test("PUT", strings.NewReader(jsonString))
	equals(t, http.StatusNotFound, w.Code)
}

func TestDeleteBookReviewHandler(t *testing.T) {
	mockDB := &MockBookReviewDB{shouldFail: false}
	deleteBookHandler := app.Wrap(app.DeleteBookReviewHandler(mockDB))
	params := httprouter.Params{httprouter.Param{Key: "id", Value: "someUUID"}}
	test := GenerateHandleTesterWithURLParams(t, deleteBookHandler, true, params)
	w := test("DELETE", url.Values{})

	assert(t, w.Code == http.StatusOK, "expected delete book review to return 200 on success, instead got %d", w.Code)

	test = GenerateHandleTesterWithURLParams(t, deleteBookHandler, true, httprouter.Params{})
	w = test("DELETE", url.Values{})
	assert(t, w.Code == http.StatusNotFound, "expected delete book review to return 404 when no matching uid provided, instead got %d", w.Code)
}
