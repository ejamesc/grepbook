package main_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/ejamesc/grepbook"
)

type MockUserDB struct {
	userExists            bool
	hasError              bool
	isUserPasswordCorrect bool
}

var user1 = &grepbook.User{ID: uint64(1), Email: "test@test.com", Password: ""}

func (db *MockUserDB) DoesAnyUserExist() bool { return db.userExists }

func (db *MockUserDB) CreateUser(email, password string) (*grepbook.User, error) {
	if db.hasError {
		return nil, fmt.Errorf("some error")
	}
	return user1, nil
}

func (db *MockUserDB) GetUser(email string) (*grepbook.User, error) {
	if db.hasError {
		return nil, fmt.Errorf("some error")
	}
	user1.Email = email
	return user1, nil
}

func (db *MockUserDB) DeleteUser(email string) error {
	if db.hasError {
		return fmt.Errorf("some error")
	}
	return nil
}

func (db *MockUserDB) IsUserPasswordCorrect(op, np string) bool {
	return db.isUserPasswordCorrect
}

func TestSignupPageHandler(t *testing.T) {
	// Test when no user exists
	mockDB := &MockUserDB{
		hasError:   false,
		userExists: false,
	}
	signupPageHandler := app.SignupPageHandler(mockDB)
	test := GenerateHandleTester(t, app.Wrap(signupPageHandler))
	w := test("GET", url.Values{})
	assert(t, w.Code == http.StatusOK, "expected signup page to return 200 instead got %d", w.Code)

	// Test when user exists
	mockDB.userExists = true
	w = test("GET", url.Values{})
	assert(t, w.Code == http.StatusFound, "expected signup with user to redirect 302 instead got %d", w.Code)
}

func TestSignupPostHandler(t *testing.T) {
	// Test when no user exists
	mockDB := &MockUserDB{
		hasError:   false,
		userExists: true,
	}
	signupPostHandler := app.SignupPostHandler(mockDB)
	test := GenerateHandleTester(t, app.Wrap(signupPostHandler))
	w := test("POST", url.Values{})
	assert(t, w.Code == http.StatusFound, "expected signup with user to redirect 302 instead got %d", w.Code)
	assert(
		t,
		len(w.HeaderMap["Location"]) > 0 && w.HeaderMap["Location"][0] == "/login",
		"expected redirect location to be /login instead got %s",
		w.HeaderMap["Location"])

	// Test when user exists and form inputs are right
	mockDB.userExists = false
	w = test("POST", url.Values{"email": {"test@test.com"}, "password": {"temporary"}})
	assert(t, w.Code == http.StatusFound, "expected signup with proper POST inputs to redirect 302 instead got %d", w.Code)
	assert(t,
		len(w.HeaderMap["Location"]) > 0 && w.HeaderMap["Location"][0] == "/",
		"expected redirect location on successful signup to be / instead got %s", w.HeaderMap["Location"])

	// Test has bad form inputs
	w = test("POST", url.Values{"email": {"blah"}, "password": {""}})
	assert(t, w.Code == http.StatusFound, "expected signup with bad POST inputs to redirect 302 instead got %d", w.Code)
	assert(t,
		len(w.HeaderMap["Location"]) > 0 && w.HeaderMap["Location"][0] == "/signup",
		"expected redirect location on unsuccessful signup to be /signup instead got %s", w.HeaderMap["Location"])

	mockDB.hasError = true
	w = test("POST", url.Values{"email": {"test@test.com"}, "password": {"temporary"}})
	assert(t, w.Code == http.StatusInternalServerError, "expected error to trigger 500 instead got %d", w.Code)
}
