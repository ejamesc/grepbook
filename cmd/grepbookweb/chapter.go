package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ejamesc/grepbook"
)

// TODO: return API errors instead
func (a *App) CreateChapterAPIHandler(db grepbook.BookReviewDB) HandlerWithError {
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

		jsonBody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return newError(http.StatusInternalServerError, "error reading request body", err)
		}

		var cpt *struct {
			Heading string `json:"heading"`
		}
		err = json.Unmarshal(jsonBody, &cpt)
		if err != nil {
			return newError(http.StatusInternalServerError, "error unmarshalling jsonBody from update", err)
		}

		if strings.TrimSpace(cpt.Heading) == "" {
			return newError(http.StatusBadRequest, "heading cannot be empty", fmt.Errorf("heading is empty"))
		}

		cp := grepbook.NewChapter(cpt.Heading, "", "")
		err = br.AddChapter(db, cp)
		if err != nil {
			return newError(http.StatusInternalServerError, "problem saving new chapter", err)
		}

		a.rndr.JSON(w, http.StatusOK, &APIResponse{Message: "Chapter created successfully"})
		return nil
	}
}

func (a *App) UpdateChapterAPIHandler(db grepbook.BookReviewDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		return nil
	}
}

func (a *App) ReorderChapterAPIHandler(db grepbook.BookReviewDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		return nil
	}
}

func (a *App) DeleteChapterAPIHandler(db grepbook.BookReviewDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		return nil
	}
}
