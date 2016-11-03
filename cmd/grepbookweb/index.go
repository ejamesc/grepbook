package main

import (
	"net/http"

	"github.com/ejamesc/grepbook"
)

func (a *App) IndexHandler(db grepbook.BookReviewDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		user := getUser(req)
		p := &localPresenter{PageTitle: "", PageURL: "", globalPresenter: a.gp, User: user}

		err := a.rndr.HTML(w, http.StatusOK, "index", p)
		if err != nil {
			a.logr.Log(newRenderErrMsg(err))
		}
		return nil
	}
}

func (a *App) AboutHandler() HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		user := getUser(req)
		p := &localPresenter{PageTitle: "About grepbook", PageURL: "/about", globalPresenter: a.gp, User: user}
		err := a.rndr.HTML(w, http.StatusOK, "about", p)
		if err != nil {
			a.logr.Log(newRenderErrMsg(err))
		}
		return nil
	}
}
		return nil
	}
}
