package main_test

import (
	"net/http"
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
