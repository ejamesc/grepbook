package main_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/ejamesc/grepbook"
	"github.com/julienschmidt/httprouter"
)

func TestCreateChapterAPIHandler(t *testing.T) {
	mockDB := &MockBookReviewDB{shouldFail: false}
	createChapterHandler := app.CreateChapterAPIHandler(mockDB)
	params := httprouter.Params{httprouter.Param{Key: "id", Value: "someUUID"}}
	test := GenerateHandleBodyTesterWithURLParams(t, app.Wrap(createChapterHandler), true, params)

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
	test = GenerateHandleBodyTesterWithURLParams(t, app.Wrap(createChapterHandler), true, params)
	w = test("POST", strings.NewReader(jsonString))
	equals(t, http.StatusNotFound, w.Code)
}

func TestUpdateChapterAPIHandler(t *testing.T) {
	mockDB := &MockBookReviewDB{shouldFail: false}
	updateChapterHandler := app.UpdateChapterAPIHandler(mockDB)

	br, _ := mockDB.GetBookReview("someid")
	cp := grepbook.NewChapter("boo", "", "")
	br.AddChapter(mockDB, cp)

	params := httprouter.Params{
		httprouter.Param{Key: "id", Value: "someUUID"},
		httprouter.Param{Key: "cid", Value: cp.ID}}
	test := GenerateHandleBodyTesterWithURLParams(t, app.Wrap(updateChapterHandler), true, params)

	// Successful Update
	jsonString := `{"heading": "Superintelligence 2"}`
	w := test("PUT", strings.NewReader(jsonString))
	equals(t, http.StatusOK, w.Code)

	// Empty but valid json
	jsonString = `{}`
	w = test("PUT", strings.NewReader(jsonString))
	equals(t, http.StatusOK, w.Code)

	// User gives nonexistent chapter id
	params[1] = httprouter.Param{Key: "cid", Value: "blablabla"}
	test2 := GenerateHandleBodyTesterWithURLParams(t, app.Wrap(updateChapterHandler), true, params)
	w = test2("PUT", strings.NewReader(jsonString))
	equals(t, http.StatusNotFound, w.Code)

	// Malformed json supplied
	w = test("PUT", strings.NewReader("LOL"))
	equals(t, http.StatusInternalServerError, w.Code)

	// User gives nonexistent uuid
	params = httprouter.Params{httprouter.Param{Key: "id", Value: ""}}
	test = GenerateHandleBodyTesterWithURLParams(t, app.Wrap(updateChapterHandler), true, params)
	w = test("PUT", strings.NewReader(jsonString))
	equals(t, http.StatusNotFound, w.Code)
}

func TestReorderChapterAPIHandler(t *testing.T) {

}

func TestDeleteChapterAPIHandler(t *testing.T) {
	mockDB := &MockBookReviewDB{shouldFail: false}
	deleteChapterHandler := app.Wrap(app.DeleteChapterAPIHandler(mockDB))

	br, _ := mockDB.GetBookReview("someid")
	cp := grepbook.NewChapter("boo", "", "")
	br.AddChapter(mockDB, cp)

	params := httprouter.Params{
		httprouter.Param{Key: "id", Value: "someUUID"},
		httprouter.Param{Key: "cid", Value: cp.ID}}
	test := GenerateHandleTesterWithURLParams(t, deleteChapterHandler, true, params)
	w := test("DELETE", url.Values{})
	equals(t, http.StatusOK, w.Code)

	// User gives nonexistant chapter code
	params[1] = httprouter.Param{Key: "cid", Value: "blablabla"}
	test = GenerateHandleTesterWithURLParams(t, deleteChapterHandler, true, params)
	w = test("DELETE", url.Values{})
	equals(t, http.StatusNotFound, w.Code)
}
