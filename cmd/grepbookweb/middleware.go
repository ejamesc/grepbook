package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ejamesc/grepbook"
	"github.com/gorilla/context"
)

const SessionName = "session-grepbook-7422573"
const UserKeyName = "user-grepbook-5320747"
const SessionKeyName = "session_key-9248129"

// appLogger is an interface for logging.
// Used to introduce a seam into the app, for testing.
type appLogger interface {
	Log(str string, v ...interface{})
}

// grepbookLogger is a wrapper for log.Logger.
type grepbookLogger struct {
	*log.Logger
}

// Log produces a log entry with the current time prepended.
func (ml *grepbookLogger) Log(str string, v ...interface{}) {
	// Prepend current time to the slice of arguments
	v = append(v, 0)
	copy(v[1:], v[0:])
	v[0] = grepbook.TimeNow().Format(time.RFC3339)
	ml.Printf("[%s] "+str, v...)
}

// newMiddlewareLogger returns a new middlewareLogger.
func newLogger() *grepbookLogger {
	return &grepbookLogger{log.New(os.Stdout, "[grepbook] ", 0)}
}

// loggingHandlerGenerator produces a loggingHandler middleware.
// loggingHandler middleware logs all requests.
func (a *App) loggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		t1 := grepbook.TimeNow()
		a.logr.Log("Started %s %s", req.Method, req.URL.Path)

		next.ServeHTTP(w, req)

		rw, ok := w.(ResponseWriter)
		if ok {
			a.logr.Log("Completed %v %s in %v", rw.Status(), http.StatusText(rw.Status()), time.Since(t1))
		} else {
			a.logr.Log("Unable to log due to invalid ResponseWriter conversion")
		}
	}
	return http.HandlerFunc(fn)
}

// responseWriterConverterHandler is a middleware to ensure that a http.Handler
// uses our own custom Response Writer, in order to capture response data.
func responseWriterWrapper(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(NewResponseWriter(w), req)
	}
	return http.HandlerFunc(fn)
}

// recoverHandlerGenerator produces a recoverHandler middleware.
// recoverHandler is a middleware that captures and recovers from panics.
func (a *App) recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				a.logr.Log("Panic: %+v", err)
				// TODO: use a nicer 500 page
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// userMiddleware is the middleware wrapper that detects and provides the user.
func (a *App) userMiddlewareGenerator(db *grepbook.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			session, err := a.store.Get(req, SessionName)
			if err != nil {
				a.logr.Log("error retrieving session from store", err)
				next.ServeHTTP(w, req)
				return
			}
			sessionKey, ok := session.Values[SessionKeyName]

			if ok {
				ssk := sessionKey.(string)
				u, err := db.GetUserBySessionKey(ssk)
				if err != nil {
					a.logr.Log("Error getting user with session key %s from DB: %s", ssk, err)
					delete(session.Values, sessionKey)
					session.Save(req, w)
				} else {
					context.Set(req, UserKeyName, u)
				}
			}
			next.ServeHTTP(w, req)
		}

		return http.HandlerFunc(fn)
	}
}

// Auth middleware is the middleware wrapper to protect authentication endpoints.
func (a *App) authMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		user := getUser(req)

		if user == nil {
			err := a.saveFlash(w, req, "You need to login to view that page!")
			if err != nil {
				a.logr.Log("Error saving flash: %s", err)
			}
			http.Redirect(w, req, "/login", 302)
			return
		} else {
			next.ServeHTTP(w, req)
		}
	}

	return http.HandlerFunc(fn)
}
