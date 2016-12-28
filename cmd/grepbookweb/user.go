package main

import (
	"net/http"

	"github.com/ejamesc/grepbook"
)

func (a *App) UserProfileHandler() HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		user := getUser(req)

		fs := a.getFlashes(w, req)
		pp := &struct {
			Flashes []interface{}
			localPresenter
		}{
			Flashes: fs,
			localPresenter: localPresenter{
				PageTitle:       user.Email + " Profile",
				PageURL:         "/user",
				globalPresenter: a.gp,
				User:            user,
			},
		}
		err := a.rndr.HTML(w, http.StatusOK, "user", pp)
		if err != nil {
			a.logr.Log(newRenderErrMsg(err))
		}
		return nil
	}
}

func (a *App) UserEditHandler(db grepbook.BookReviewDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		return nil
	}
}
