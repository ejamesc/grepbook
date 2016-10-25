package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ejamesc/grepbook"
	"github.com/gorilla/context"
)

const sessionName = "session"
const userKeyName = "user"
const sessionKeyName = "session_key"

// middlewareLogger is a wrapper for log.Logger.
type middlewareLogger struct {
	*log.Logger
}

// newMiddlewareLogger returns a new middlewareLogger.
func newMiddlewareLogger() *middlewareLogger {
	return &middlewareLogger{log.New(os.Stdout, "[grepbook] ", 0)}
}

// loggingHandler is a middleware to log requests.
func (a *App) loggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		t1 := grepbook.TimeNow()
		a.Log("Started %s %s", req.Method, req.URL.Path)

		next.ServeHTTP(w, req)

		rw := w.(ResponseWriter)
		a.Log("Completed %v %s in %v", rw.Status(), http.StatusText(rw.Status()), time.Since(t1))
	}

	return http.HandlerFunc(fn)
}

// recoverHandler is a middleware that captures and recovers from panics.
func (a *App) recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				a.Log("Panic: %+v", err)
				// TODO: use a nicer 500 page
				http.Error(w, http.StatusText(500), 500)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// userMiddleware is the middleware wrapper that detects and provides the user.
func (a *App) userMiddlewareGenerator(db grepbook.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			session, _ := a.store.Get(req, sessionName)
			sessionKey, ok := session.Values[sessionKeyName]

			if ok {
				ssk := sessionKey.(string)
				u, err := db.GetUserBySessionKey(ssk)
				if err != nil {
					a.Log("Error getting user with session key %s from DB: %s", ssk, err)
					delete(session.Values, sessionKey)
					session.Save(req, w)
				} else {
					context.Set(req, userKeyName, u)
				}
			}
			next.ServeHTTP(w, req)
		}

		return http.HandlerFunc(fn)
	}
}
