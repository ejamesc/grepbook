package main

import (
	"net/http"

	"github.com/ejamesc/grepbook"
)

func (a *App) IndexHandler(db grepbook.BookReviewDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		return nil
	}
}
