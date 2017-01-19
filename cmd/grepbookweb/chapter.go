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
			return new500Error("error unmarshalling jsonBody from update", err)
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
		jsonBody, bookReview, sErr := processChapterReq(req, db)
		if sErr != nil {
			return sErr
		}

		var cpd *struct {
			OldIndex int `json:"old_index"`
			NewIndex int `json:"new_index"`
		}
		err := json.Unmarshal(jsonBody, &cpd)
		if err != nil {
			return new500Error("error unmarshalling jsonBody from reorder", err)
		}

		if cpd.OldIndex == cpd.NewIndex || cpd.OldIndex <= 0 && cpd.NewIndex <= 0 {
			return newError(http.StatusBadRequest, "old index and new index cannot both be the same, or < 0", fmt.Errorf("bad indexes"))
		}

		err = bookReview.ReorderChapter(db, cpd.OldIndex, cpd.NewIndex)
		if err != nil {
			return new500Error("error reordering chapter", err)
		}

		a.rndr.JSON(w, http.StatusOK, &APIResponse{Message: "Chapter reordered successfully"})
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
		if err != nil && err != grepbook.ErrNoRows {
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
