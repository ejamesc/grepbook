package main_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestCreateChapterAPIHandler(t *testing.T) {
	mockDB := &MockBookReviewDB{shouldFail: false}
	updateBookReviewHandler := app.CreateChapterAPIHandler(mockDB)
	params := httprouter.Params{httprouter.Param{Key: "id", Value: "someUUID"}}
	test := GenerateHandleBodyTesterWithURLParams(t, app.Wrap(updateBookReviewHandler), true, params)

	// Successful create
	jsonString := `{"heading": "Superintelligence"}`
	w := test("POST", strings.NewReader(jsonString))
	equals(t, http.StatusOK, w.Code)

	// Heading left empty
	w = test("POST", strings.NewReader(`{"heading": ""}`))
	equals(t, http.StatusBadRequest, w.Code)

	// Malformed json supplied
	w = test("POST", strings.NewReader("LOL"))
	equals(t, http.StatusInternalServerError, w.Code)

	// User gives nonexistent uuid
	params = httprouter.Params{httprouter.Param{Key: "id", Value: ""}}
	test = GenerateHandleBodyTesterWithURLParams(t, app.Wrap(updateBookReviewHandler), true, params)
	w = test("POST", strings.NewReader(jsonString))
	equals(t, http.StatusNotFound, w.Code)
}

func TestUpdateChapterAPIHandler(t *testing.T) {

}

func TestReorderChapterAPIHandler(t *testing.T) {

}

func TestDeleteChapterAPIHandler(t *testing.T) {

}
