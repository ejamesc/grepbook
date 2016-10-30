package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ejamesc/grepbook"
	"github.com/ejamesc/grepbook/cmd/grepbookweb"
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

func (db *MockUserDB) CreateSessionForUser(email string) (*grepbook.Session, error) {
	if db.hasError {
		return nil, fmt.Errorf("some error")
	}
	return &grepbook.Session{Key: "abcd1234", Email: email}, nil
}

func TestSignupPageHandler(t *testing.T) {
	// Test when no user exists
	mockDB := &MockUserDB{
		hasError:   false,
		userExists: false,
	}
	signupPageHandler := app.SignupPageHandler(mockDB)
	test := GenerateHandleTester(t, app.Wrap(signupPageHandler), false)
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
	test := GenerateHandleTester(t, app.Wrap(signupPostHandler), false)
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
	assert(t, len(w.HeaderMap["Set-Cookie"]) > 0, "expected session cookie to be set on successful signup.")

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

func TestLoginPageHandler(t *testing.T) {
	lp := app.LoginPageHandler()
	test := GenerateHandleTester(t, app.Wrap(lp), false)
	w := test("GET", url.Values{})
	assert(t, w.Code == http.StatusOK, "expected signup page to return 200 instead got %d", w.Code)

	// Test when user is already logged in
	// You need to set the session with the key
	test = GenerateHandleTester(t, app.Wrap(lp), true)
	w = test("GET", url.Values{})
	assert(t, w.Code == http.StatusFound, "expected logged in user to redirect away from login page, instead got %d", w.Code)
	assert(t,
		len(w.HeaderMap["Location"]) > 0 && w.HeaderMap["Location"][0] == "/",
		"expected redirect location on already logged in user to be / instead got %s", w.HeaderMap["Location"])
}

func TestLoginPostHandler(t *testing.T) {
	mockDB := &MockUserDB{
		hasError:              false,
		isUserPasswordCorrect: true,
		userExists:            true,
	}
	lp := app.LoginPostHandler(mockDB)
	test := GenerateHandleTester(t, app.Wrap(lp), false)
	w := test("POST", url.Values{"email": {"test@test.com"}, "password": {"temporary"}})
	assert(t, w.Code == http.StatusFound, "expected successful login with proper POST inputs to redirect 302 instead got %d", w.Code)
	assert(t,
		len(w.HeaderMap["Location"]) > 0 && w.HeaderMap["Location"][0] == "/",
		"expected redirect location on successful login to be / instead got %s", w.HeaderMap["Location"])
	assert(t, len(w.HeaderMap["Set-Cookie"]) > 0, "expected session cookie to be set on successful login.")

	// Wrong password
	mockDB.isUserPasswordCorrect = false
	w = test("POST", url.Values{"email": {"test@test.com"}, "password": {"temporary"}})
	assert(t, w.Code == http.StatusFound, "expected successful login with proper POST inputs to redirect 302 instead got %d", w.Code)
	assert(t,
		len(w.HeaderMap["Location"]) > 0 && w.HeaderMap["Location"][0] == "/login",
		"expected redirect location on unsuccessful login to be /login instead got %s", w.HeaderMap["Location"])
}

func TestLogoutHandler(t *testing.T) {
	lp := app.Wrap(app.LogoutHandler())
	req, err := http.NewRequest("POST", "", nil)
	ok(t, err)
	w := httptest.NewRecorder()

	store := app.GetStore()
	ss, err := store.Get(req, main.SessionName)
	ok(t, err)
	ss.Values[main.SessionKeyName] = "abcd1234"
	ss.Save(req, w)

	lp.ServeHTTP(w, req)
	assert(t, w.Code == http.StatusFound, "expected successful logout to redirect 302 instead got %d", w.Code)
	assert(t,
		len(w.HeaderMap["Location"]) > 0 && w.HeaderMap["Location"][0] == "/",
		"expected redirect location on successful logout to be / instead got %s", w.HeaderMap["Location"])

	session, err := store.Get(req, main.SessionName)
	_, exists := session.Values[main.SessionKeyName]
	assert(t, !exists, "expected session to have been deleted")
}
