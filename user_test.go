package grepbook_test

import (
	"testing"

	"github.com/ejamesc/grepbook"
)

func TestCreateUser(t *testing.T) {
	user, err := testDB.CreateUser("blah@blah.com", "somepassword")
	ok(t, err)
	assert(t, user.Email != "", "expected user email to be filled but got empty string")
	assert(t, user.Password == "", "expected user password to be empty but got some string instead")
	defer testDB.DeleteUser(user.Email)

	res := testDB.DoesAnyUserExist()
	equals(t, true, res)

	user, err = testDB.CreateUser("blah@blah.com", "somepassword")
	assert(t, err == grepbook.ErrDuplicateRow, "expected duplicate email to cause ErrDuplicateRow but no error received")
	assert(t, user == nil, "expected user to be nil on error")
}

func TestGetUser(t *testing.T) {
	user, err := testDB.GetUser(user1.Email)
	ok(t, err)
	assert(t, user.Email != "", "expected user email to be filled but got empty string")
	assert(t, user.Password == "", "expected user password to be empty but got some string instead")

	user, err = testDB.GetUser("")
	assert(t, err == grepbook.ErrNoRows, "expected no email to return an ErrNoRows error")
}

func TestDeleteUser(t *testing.T) {
	uCreate, err := testDB.CreateUser("kimjong@il.com", "allthepower")
	ok(t, err)
	err = testDB.DeleteUser(uCreate.Email)
	ok(t, err)
	uGet, err := testDB.GetUser(uCreate.Email)
	assert(t, err == grepbook.ErrNoRows, "expect error to be ErrNoRows because user has already been deleted")
	assert(t, uGet == nil, "expect user to no longer exist, return nil")
}

func TestIsUserPasswordCorrect(t *testing.T) {
	res := testDB.IsUserPasswordCorrect(user1.Email, "test")
	equals(t, true, res)
	res = testDB.IsUserPasswordCorrect(user1.Email, "")
	equals(t, false, res)
	res = testDB.IsUserPasswordCorrect(user1.Email, "someshit")
	equals(t, false, res)
}
