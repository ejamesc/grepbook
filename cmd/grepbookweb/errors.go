package main

import "fmt"

// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
}

// StatusError represents an error with an associated HTTP status code.
type StatusError struct {
	Code int
	Err  error
}

// Allows StatusError to satisfy the error interface.
func (se StatusError) Error() string {
	return se.Err.Error()
}

// Returns our HTTP status code.
func (se StatusError) Status() int {
	return se.Code
}

func newError(code int, msg string, err error) *StatusError {
	if err != nil {
		return &StatusError{Code: code, Err: fmt.Errorf(msg+": %s", err)}
	} else {
		return &StatusError{Code: code, Err: fmt.Errorf(msg)}
	}
}

func newSessionSaveError(err error) *StatusError {
	return &StatusError{Code: 500, Err: fmt.Errorf("problem saving to cookie store: %s", err)}
}

func newRenderErrMsg(err error) string {
	return fmt.Sprintf("error rendering HTML: %s", err)
}
