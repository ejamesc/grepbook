package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestUserProfileGet(t *testing.T) {
	userHandler := app.UserProfileHandler()
	test := GenerateHandleTester(t, app.Wrap(userHandler), true)
	w := test("GET", url.Values{})
	assert(t, http.StatusOK == w.Code, "expected user profile edit page to return 200, instead got %d", w.Code)

	// This shouldn't be possible, but say the user is logged out
	test = GenerateHandleTester(t, app.Wrap(userHandler), false)
	w = test("GET", url.Values{})
	assert(t, http.StatusInternalServerError == w.Code, "expected user profile edit page to return 500 when no user supplied, instead got %d", w.Code)
}

func TestUserProfilePost(t *testing.T) {
	mockDB := &MockUserDB{
		hasError:              false,
		isUserPasswordCorrect: true,
		userExists:            true,
	}
	userPostHandler := app.UserEditHandler(mockDB)
	test := GenerateHandleTester(t, app.Wrap(userPostHandler), true)
	tests := []struct {
		vals url.Values
		res  string
	}{
		{vals: url.Values{"email": {"stupid@stupid.com"}}, res: ""},
		{vals: url.Values{"email": {"test@test.com"}, "name": {"Kim Jong Un"}}, res: ""},
		{vals: url.Values{"old-password": {"blah"}, "new-password": {"stupid"}, "new-password2": {"stupid"}}, res: ""},
		{vals: url.Values{"name": {""}}, res: ""},
		{vals: url.Values{"email": {"blah"}}, res: "Bad Request\n"},
		{vals: url.Values{"old-password": {"blah"}, "new-password": {"stupid"}, "new-password2": {"crazy"}}, res: "Bad Request\n"},
		{vals: url.Values{"old-password": {""}, "new-password": {"stupid"}, "new-password2": {"stupid"}}, res: "Bad Request\n"},
		{vals: url.Values{"old-password": {"blah"}, "new-password": {"whoosh"}, "new-password2": {"whoosh"}}, res: ""},
	}
	var w *httptest.ResponseRecorder
	for _, ts := range tests {
		w = test("POST", ts.vals)
		fmt.Printf("%+v\n", w.Body.String())
		assert(t, w.Body.String() == ts.res, "expected user profile post handler to return %s instead got %s", ts.res, w.Body.String())
	}
}
