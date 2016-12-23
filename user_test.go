package grepbook_test

import (
	"testing"

	"github.com/ejamesc/grepbook"
)

func TestCreateUser(t *testing.T) {
	user, err := testDB.CreateUser("blah@blah.com", "somepassword")
	ok(t, err)
	assert(t, user.ID != uint64(0), "expect user ID to be filled with sequence")
	assert(t, user.Email != "", "expect user email to be filled but got empty string")
	assert(t, user.Password == "", "expect user password to be empty but got some string instead")
	defer testDB.DeleteUser(user.Email)

	res := testDB.DoesAnyUserExist()
	equals(t, true, res)

	user, err = testDB.CreateUser("blah@blah.com", "somepassword")
	assert(t, err == grepbook.ErrDuplicateRow, "expected duplicate email to cause ErrDuplicateRow but no error received")
	assert(t, user == nil, "expected user to be nil on error")
}

func TestCreateUserValidation(t *testing.T) {
	user, err := testDB.CreateUser("blah", "somepassword")
	assert(t, err != nil, "expected invalid email to return error")
	assert(t, user == nil, "expected user to be nil on error")

	user, err = testDB.CreateUser("blah@blah.com", "")
	assert(t, err != nil, "expected invalid email to return error")
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

func TestUpdateUser(t *testing.T) {
	testEmail := "kimil@sung.com"
	uCreate, err := testDB.CreateUser(testEmail, "allthepower2")
	ok(t, err)

	// Test name edit
	resUser, err := testDB.UpdateUser(testEmail, grepbook.UserDelta{Name: "Kim Il Sung"})
	ok(t, err)
	assert(t, uCreate.Name != resUser.Name, "expect edited user to have a different name")
	assert(t, resUser.Name == "Kim Il Sung", "expect edited user to be called Kim Il Sung after updating")

	// Ensure the password isn't returned with the user
	equals(t, "", uCreate.Password)
	equals(t, "", resUser.Password)

	// Test password edit
	resUser, err = testDB.UpdateUser(testEmail, grepbook.UserDelta{Password: "someotherpasswd"})
	ok(t, err)
	assert(t, testDB.IsUserPasswordCorrect(testEmail, "someotherpasswd"), "expect password to have been udpated")

	// Test email edit
	resUser, err = testDB.UpdateUser(testEmail, grepbook.UserDelta{Email: "blah@dprk.com"})
	ok(t, err)
	assert(t, resUser.Email != testEmail, "expect email to have been changed")
	u, err := testDB.GetUser("blah@dprk.com")
	ok(t, err)
	equals(t, "Kim Il Sung", u.Name)
	equals(t, "blah@dprk.com", u.Email)
}
