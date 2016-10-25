package main

import (
	"time"

	"github.com/ejamesc/grepbook"
	"github.com/gorilla/sessions"
	"github.com/microcosm-cc/bluemonday"
	"github.com/unrolled/render"
)

// App is the main app.
type App struct {
	rndr   *render.Render
	router *Router
	store  *sessions.CookieStore
	gp     globalPresenter
	log    *middlewareLogger
	bm     *bluemonday.Policy
}

// globalPresenter contains the fields necessary for presenting in all templates
type globalPresenter struct {
	SiteName    string
	Description string
	SiteURL     string
}

// localPresenter contains the fields necessary for specific pages.
type localPresenter struct {
	PageTitle string
	PageURL   string
	User      *grepbook.User
	globalPresenter
}

// Log produces a log entry with the current time prepended.
func (a *App) Log(str string, v ...interface{}) {
	// Prepend current time to the slice of arguments
	v = append(v, 0)
	copy(v[1:], v[0:])
	v[0] = grepbook.TimeNow().Format(time.RFC3339)
	a.log.Printf("[%s] "+str, v...)
}
