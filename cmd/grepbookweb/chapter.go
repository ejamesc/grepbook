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
		jsonBody, br, sErr := processChapterReq(req, db)
		if sErr != nil {
			return sErr
		}

		var cpt *struct {
			Heading string `json:"heading"`
		}
		err := json.Unmarshal(jsonBody, &cpt)
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
		jsonBody, bookReview, sErr := processChapterReq(req, db)
		if sErr != nil {
			return sErr
		}
		params := GetParamsObj(req)
		chapterID := params.ByName("cid")

		var cpd grepbook.ChapterDelta
		err := json.Unmarshal(jsonBody, &cpd)
		if err != nil {
			return new500Error("error unmarshalling jsonBody from update", err)
		}

		err = bookReview.UpdateChapter(db, chapterID, cpd)
		if err != nil {
			if err == grepbook.ErrNoRows {
				return new404Error(fmt.Sprintf("chapter ID %s not found", chapterID), err)
			}
			return new500Error("error updating chapter", err)
		}

		a.rndr.JSON(w, http.StatusOK, &APIResponse{Message: "Chapter updated successfully"})
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
		_, bookReview, sErr := processChapterReq(req, db)
		if sErr != nil {
			return sErr
		}
		params := GetParamsObj(req)
		chapterID := params.ByName("cid")

		err := bookReview.DeleteChapter(db, chapterID)
		if err != nil {
			if err == grepbook.ErrNoRows {
				return new404Error("no such chapter", err)
			}
			return new500Error("error deleting chapter", err)
		}

		a.rndr.JSON(w, http.StatusOK, &APIResponse{Message: "Chapter deleted successfully"})
		return nil
	}
}

func processChapterReq(req *http.Request, db grepbook.BookReviewDB) (jsonBody []byte, bookReview *grepbook.BookReview, sErr *StatusError) {
	params := GetParamsObj(req)
	uid := params.ByName("id")
	br, err := db.GetBookReview(uid)
	if err != nil {
		if err == grepbook.ErrNoRows {
			return []byte{}, nil, new404Error("no book review with that uid found", err)
		}
		return []byte{}, nil, new500Error("error retrieving book review", err)
	}
	jb, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return []byte{}, nil, new500Error("error reading request body", err)
	}
	return jb, br, nil
}
