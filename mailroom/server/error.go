// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package server

import (
	"errors"
	"fmt"
	"net/http"
)

// Error is a custom error type that includes an HTTP status code
type Error struct {
	Code   int
	Reason error
}

func (h *Error) Error() string {
	if h.Reason != nil {
		return fmt.Sprintf("internal %d: %v", h.Code, h.Reason.Error())
	}
	return fmt.Sprintf("internal %d: something happened, perhaps", h.Code)
}

// handlerFunc is basically a http.HandlerFunc that can return an error
// This allows our handlers to return the special Error type above, which HandleErr would then recognize
// and generate a consistent error response accordingly. Without this, each handler would have to write
// its own HTTP error response, which could be error-prone and inconsistent.
type handlerFunc func(http.ResponseWriter, *http.Request) error

// HandleErr wraps an HTTP handlerFunc and returns a new one that
// handles errors by writing an HTTP response with the error message and status
func HandleErr(h handlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if err := h(writer, request); err != nil {
			code := 500
			if he := (*Error)(nil); errors.As(err, &he) {
				code = he.Code
			}
			writer.WriteHeader(code)
			_, _ = fmt.Fprintf(writer, "%v\n", err)
		}
	}
}
