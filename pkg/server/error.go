// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package server

import (
	"errors"
	"fmt"
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

func (h *Error) Is(target error) bool {
	var err *Error
	if ok := errors.As(target, &err); !ok {
		return false
	}

	return err.Code == h.Code && (errors.Is(err.Reason, h.Reason) || err.Reason.Error() == h.Reason.Error())
}
