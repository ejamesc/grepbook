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

		isNew := br.IsNew()
		pp := struct {
			BookReview *grepbook.BookReview
			BRHTML     template.HTML
			IsNew      bool
			*localPresenter
		}{
			BookReview:     br,
			BRHTML:         template.HTML(br.OverviewHTML),
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

func (a *App) WritePageDisplayHandler(db grepbook.BookReviewDB) HandlerWithError {
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

		isNew := strings.TrimSpace(br.OverviewHTML) == ""
		pp := struct {
			BookReview *grepbook.BookReview
			BRHTML     template.HTML
			BRJSON     string
			IsNew      bool
			*localPresenter
		}{
			BookReview:     br,
			BRHTML:         template.HTML(br.OverviewHTML),
			IsNew:          isNew,
			localPresenter: &localPresenter{PageTitle: "Summary template", PageURL: "/summary", globalPresenter: a.gp, User: user},
		}

		brjson, err := json.Marshal(br)
		if err == nil {
			pp.BRJSON = string(brjson)
		} else {
			a.logr.Log("problem marshalling book review: ", err)
		}

		err = a.rndr.HTML(w, http.StatusOK, "write", pp)
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
			a.saveFlash(w, req, "Book review title cannot be empty!")
			http.Redirect(w, req, "/", 302)
			return newError(400, "title cannot be empty", fmt.Errorf("title is empty"))
		}

		chapters := grepbook.CreateChapters(chapterList)
		br, err := db.CreateBookReview(title, author, url, "", "", chapters)
		if err != nil {
			return err
		}
		http.Redirect(w, req, "/summaries/"+br.UID+"/edit", 302)
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
			return newError(http.StatusInternalServerError, "error retrieving book review: ", err)
		}

		var tbr *grepbook.BookReview
		jsonBody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return newError(http.StatusInternalServerError, "error reading request body", err)
		}
		err = json.Unmarshal(jsonBody, &tbr)
		if err != nil {
			return newError(http.StatusInternalServerError, "error unmarshalling jsonBody from update", err)
		}

		if tbr.UID != br.UID {
			return newError(http.StatusForbidden, "user does not own book review", err)
		}

		mergeBookReviewDeltas(br, tbr)
		br.DateTimeUpdated = time.Now()
		err = br.Save(db)
		if err != nil {
			return newError(http.StatusInternalServerError, "error saving book review", err)
		}

		apiResp := &APIResponse{Message: "Book review updated successfully"}
		a.rndr.JSON(w, http.StatusOK, apiResp)
		return nil
	}
}

func (a *App) DeleteBookReviewHandler(db grepbook.BookReviewDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		params := GetParamsObj(req)
		uid := params.ByName("id")
		br, err := db.GetBookReview(uid)
		if err != nil {
			if err == grepbook.ErrNoRows {
				return newError(http.StatusNotFound, "no book review with that uid found", err)
			}
			return newError(http.StatusInternalServerError, "error retrieving book review: ", err)
		}

		err = db.DeleteBookReview(br.UID)
		if err != nil {
			return newError(http.StatusInternalServerError, "error deleting book review: ", err)
		}
		apiResp := &APIResponse{Message: "Book review deleted successfully"}
		a.rndr.JSON(w, http.StatusOK, apiResp)

		return nil
	}
}

func (a *App) CreateChapterHandler(db grepbook.BookReviewDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		return nil
	}
}

func mergeBookReviewDeltas(oldBR, newBR *grepbook.BookReview) {
	if newBR.Title != "" || newBR.Title != oldBR.Title {
		oldBR.Title = newBR.Title
	}
	if newBR.BookAuthor != "" || newBR.BookAuthor != oldBR.BookAuthor {
		oldBR.BookAuthor = newBR.BookAuthor
	}
	if newBR.BookURL != "" || newBR.BookURL != oldBR.BookURL {
		oldBR.BookURL = newBR.BookURL
	}
	if newBR.OverviewHTML != "" || newBR.OverviewHTML != oldBR.OverviewHTML {
		oldBR.OverviewHTML = newBR.OverviewHTML
	}
	if newBR.Delta != "" || newBR.Delta != oldBR.Delta {
		oldBR.Delta = newBR.Delta
	}
	if newBR.IsOngoing != oldBR.IsOngoing {
		oldBR.IsOngoing = newBR.IsOngoing
	}
}
