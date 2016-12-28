package main

import (
	"net/http"
	"sort"

	"github.com/ejamesc/grepbook"
)

func (a *App) IndexHandler(db grepbook.BookReviewDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		user := getUser(req)

		brs, err := db.GetAllBookReviews()
		if err != nil {
			return newError(500, "problem retrieving book reviews", err)
		}

		obr, dbr := sortBookReviews(brs)
		pp := struct {
			Ongoing grepbook.BookReviewArray
			Done    grepbook.BookReviewArray
			*localPresenter
		}{
			Ongoing:        obr,
			Done:           dbr,
			localPresenter: &localPresenter{PageTitle: "", PageURL: "", globalPresenter: a.gp, User: user},
		}
		err = a.rndr.HTML(w, http.StatusOK, "index", pp)
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

func (a *App) BookIndexHandler(db grepbook.BookReviewDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		return nil
	}
}

func (a *App) NotFoundHandler(w http.ResponseWriter, req *http.Request) {
	user := getUser(req)
	lp := &localPresenter{
		PageTitle:       "Page not found",
		PageURL:         "/404",
		globalPresenter: a.gp,
	}
	if user != nil {
		lp.User = user
	}
	a.rndr.HTML(w, http.StatusNotFound, "404", lp)
	return
}

// sortBookReviews returns ongoing and done book reviews, sorted in reverse chronological order
func sortBookReviews(brs grepbook.BookReviewArray) (ongoing, done grepbook.BookReviewArray) {
	og, dn := grepbook.BookReviewArray{}, grepbook.BookReviewArray{}
	for _, v := range brs {
		if v.IsOngoing {
			og = append(og, v)
		} else {
			dn = append(dn, v)
		}
	}
	sort.Sort(sort.Reverse(og))
	sort.Sort(sort.Reverse(dn))
	return og, dn
}
