package grepbook_test

import (
	"testing"

	"github.com/ejamesc/grepbook"
)

func TestCreateSessionForUser(t *testing.T) {
	session, err := testDB.CreateSessionForUser(user1.Email)
	ok(t, err)
	defer testDB.DeleteSession(session.Key)
	equals(t, user1.Email, session.Email)
	assert(t, session.Key != "", "expect session key to not be empty")
}

func TestBadEmailCreateSessionForUser(t *testing.T) {
	session, err := testDB.CreateSessionForUser("blah")
	assert(t, err != nil, "expect there to be error for invalid email")
	assert(t, session == nil, "expect session to be nil on failed create")
}

func TestGetUserBySessionKey(t *testing.T) {
	session, err := testDB.CreateSessionForUser(user1.Email)
	ok(t, err)
	defer testDB.DeleteSession(session.Key)

	user, err := testDB.GetUserBySessionKey(session.Key)
	ok(t, err)
	equals(t, user, user1)
}

func TestGetEmptySession(t *testing.T) {
	user, err := testDB.GetUserBySessionKey("")
	assert(t, err == grepbook.ErrNoRows, "expect no such session to exist")
	assert(t, user == nil, "expect user to be nil because no such session exists")
}

func TestDeleteSession(t *testing.T) {
	session, err := testDB.CreateSessionForUser(user1.Email)
	ok(t, err)

	err = testDB.DeleteSession(session.Key)
	ok(t, err)

	user, err := testDB.GetUserBySessionKey(session.Key)
	assert(t, err == grepbook.ErrNoRows, "expected deleted session to return an ErrNoRows error")
	assert(t, user == nil, "expected deleted session to return nil")
}
