// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package server

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "with reason",
			err:  &Error{Code: 403, Reason: errors.New("forbidden")},
			want: "internal 403: forbidden",
		},
		{
			name: "without reason",
			err:  &Error{Code: 403},
			want: "internal 403: something happened, perhaps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

func TestError_Is(t *testing.T) {
	t.Parallel()

	target := &Error{Code: 403, Reason: errors.New("forbidden")}

	tests := []struct {
		name  string
		other error
		want  bool
	}{
		{
			name:  "exact same object",
			other: target,
			want:  true,
		},
		{
			name:  "different object, same values",
			other: &Error{Code: 403, Reason: errors.New("forbidden")},
			want:  true,
		},
		{
			name:  "different code",
			other: &Error{Code: 400, Reason: errors.New("forbidden")},
			want:  false,
		},
		{
			name:  "different reason",
			other: &Error{Code: 403, Reason: errors.New("not allowed")},
			want:  false,
		},
		{
			name:  "different type",
			other: errors.New("forbidden"),
			want:  false,
		},
		{
			name:  "wrapped",
			other: fmt.Errorf("wrapped: %w", target),
			want:  true,
		},
		{
			name:  "inner reason",
			other: target.Reason,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, errors.Is(target, tt.other))
		})
	}
}
