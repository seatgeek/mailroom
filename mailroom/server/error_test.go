// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrHandler(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName    string
		err         error
		wantStatus  int
		wantMessage string
	}{
		{
			testName:    "no error",
			err:         nil,
			wantStatus:  200,
			wantMessage: "OK",
		},
		{
			testName:    "internal http error",
			err:         &Error{Code: 403, Reason: errors.New("forbidden")},
			wantStatus:  403,
			wantMessage: "internal 403: forbidden",
		},
		{
			testName:    "other error",
			err:         fmt.Errorf("something went wrong"),
			wantStatus:  500,
			wantMessage: "something went wrong",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			responseRecorder := httptest.NewRecorder()

			handler := func(writer http.ResponseWriter, _ *http.Request) error {
				if tc.err != nil {
					return tc.err
				}

				writer.WriteHeader(200)
				_, _ = writer.Write([]byte("OK"))

				return nil
			}

			wrappedHandler := HandleErr(handler)
			wrappedHandler(responseRecorder, new(http.Request))

			assert.Equal(t, tc.wantStatus, responseRecorder.Code)
			assert.Equal(t, tc.wantMessage, strings.TrimSpace(responseRecorder.Body.String()))
		})
	}
}
