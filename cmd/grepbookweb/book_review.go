package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

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
			BRJSON     string
			IsNew      bool
			*localPresenter
		}{
			BookReview:     br,
			BRHTML:         template.HTML(br.HTML),
			IsNew:          isNew,
			localPresenter: &localPresenter{PageTitle: "Summary template", PageURL: "/summary", globalPresenter: a.gp, User: user},
		}

		brjson, err := json.Marshal(br)
		if err == nil {
			pp.BRJSON = string(brjson)
		} else {
			a.logr.Log("problem marshalling book review: ", err)
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
		params := GetParamsObj(req)
		uid := params.ByName("id")
		br, err := db.GetBookReview(uid)
		if err != nil {
			if err == grepbook.ErrNoRows {
				return newError(http.StatusNotFound, "no book review with that uid found", err)
			}
			return newError(500, "error retrieving book review:", err)
		}

		var tbr *grepbook.BookReview
		jsonBody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return newError(500, "error reading request body", err)
		}
		err = json.Unmarshal(jsonBody, &tbr)
		if err != nil {
			return newError(500, "error unmarshalling jsonBody from update", err)
		}

		if tbr.UID != br.UID {
			return newError(403, "user does not own book review", err)
		}

		saveBR(br, tbr)
		br.DateTimeUpdated = time.Now()
		br.Save(db)
		return nil
	}
}

func (a *App) DeleteBookReviewHandler(db grepbook.BookReviewDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		return nil
	}
}

func saveBR(oldBR, newBR *grepbook.BookReview) {
	if newBR.Title != "" || newBR.Title != oldBR.Title {
		oldBR.Title = newBR.Title
	}
	if newBR.BookAuthor != "" || newBR.BookAuthor != oldBR.BookAuthor {
		oldBR.BookAuthor = newBR.BookAuthor
	}
	if newBR.BookURL != "" || newBR.BookURL != oldBR.BookURL {
		oldBR.BookURL = newBR.BookURL
	}
	if newBR.HTML != "" || newBR.HTML != oldBR.HTML {
		oldBR.HTML = newBR.HTML
	}
	if newBR.Delta != "" || newBR.Delta != oldBR.Delta {
		oldBR.Delta = newBR.Delta
	}
	if newBR.IsOngoing != oldBR.IsOngoing {
		oldBR.IsOngoing = newBR.IsOngoing
	}
}
