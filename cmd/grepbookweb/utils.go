package main

import (
	"net/http"

	"github.com/ejamesc/grepbook"
	"github.com/gorilla/context"
)

// saveFlash is a utility helper to save Flashes to the session cookie.
func (a *App) saveFlash(w http.ResponseWriter, req *http.Request, msg string) error {
	session, err := a.store.Get(req, SessionName)
	if err != nil {
		return err
	}
	session.AddFlash(msg)
	err = session.Save(req, w)
	if err != nil {
		return err
	}
	return nil
}

// getUser returns the user from the context object in the request.
func getUser(req *http.Request) *grepbook.User {
	if rv := context.Get(req, UserKeyName); rv != nil {
		res := rv.(*grepbook.User)
		return res
	}
	return nil
}
