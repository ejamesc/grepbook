package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/ejamesc/grepbook"
)

func (a *App) ReadHandler(db grepbook.BookReviewDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		params := GetParamsObj(req)
		uid := params.ByName("id")

		user := getUser(req)

		br, err := db.GetBookReview(uid)
		if err != nil {
			if err == grepbook.ErrNoRows {
				return newError(http.StatusNotFound, "no book review with that uid found", err)
			}
			return newError(500, "error retrieving book review:", err)
		}

		isNew := strings.TrimSpace(br.HTML) == ""
		pp := struct {
			BookReview *grepbook.BookReview
			BRHTML     template.HTML
			IsNew      bool
			*localPresenter
		}{
			BookReview:     br,
			BRHTML:         template.HTML(br.HTML),
			IsNew:          isNew,
			localPresenter: &localPresenter{PageTitle: "Summary template", PageURL: "/summary", globalPresenter: a.gp, User: user},
		}

		err = a.rndr.HTML(w, http.StatusOK, "read", pp)
		if err != nil {
			a.logr.Log(newRenderErrMsg(err))
		}
		return nil
	}
}

func (a *App) CreateBookReviewHandler(db grepbook.BookReviewDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		title, author, url, chapterList := req.FormValue("title"), req.FormValue("author"), req.FormValue("url"), req.FormValue("chapters")

		if strings.TrimSpace(title) == "" {
			return newError(400, "title cannot be empty", fmt.Errorf("title is empty"))
		}

		chapters := grepbook.CreateChapters(chapterList)
		br, err := db.CreateBookReview(title, author, url, "", "", chapters)
		if err != nil {
			return err
		}
		http.Redirect(w, req, "/summaries/"+br.UID, 302)
		return nil
	}
}

func (a *App) UpdateBookReviewHandler(db grepbook.BookReviewDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		return nil
	}
}

func (a *App) DeleteBookReviewHandler(db grepbook.BookReviewDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		return nil
	}
}
