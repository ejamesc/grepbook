package main

import "net/http"

// handlerWithError is a handler function that returns an error.
// This is the primary function type we'll use for all http handlers in grepbook.
// The problem this solves is that we may forget to return in normal http.Handlers.
// Forcing us to return an error (and handling that error in this Wrap function) prevents this from happening.
type HandlerWithError func(http.ResponseWriter, *http.Request) error

// Wrap allows us to turn HandlerWithError into http.Handler
func (a *App) Wrap(hn HandlerWithError) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		err := hn(w, req)
		if err != nil {
			a.handleError(w, req, err)
		}
	}
	return http.HandlerFunc(fn)
}

// handleError is the catch-all error function.
// It handles generic errors that may be returned by any http handler.
func (a *App) handleError(w http.ResponseWriter, req *http.Request, err error) {
	u := getUser(req)
	lp := &localPresenter{PageTitle: "404 Page Not Found", PageURL: req.URL.String(), globalPresenter: a.gp, User: u}
	switch e := err.(type) {
	case Error:
		// We can retrieve the status here and write out a specific
		// HTTP status code.
		if e.Status() == 404 {
			a.rndr.HTML(w, e.Status(), "404", lp)
		} else {
			http.Error(w, http.StatusText(e.Status()), e.Status())
		}
		a.logr.Log("HTTP %d - %s\n", e.Status(), e)
	default:
		// Any error types we don't specifically look out for default to serving a terrible HTTP 500
		a.logr.Log("HTTP %d - %s\n", 500, e)
		a.rndr.HTML(w, http.StatusInternalServerError, "500", lp)
	}
}
