// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package webhooks

import (
	"net/http/httptest"
	"testing"

	"github.com/go-playground/webhooks/v6/gitlab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdapter_Parse(t *testing.T) {
	t.Parallel()

	somePayload := struct{}{}

	tests := []struct {
		name        string
		hook        hook[string]
		wantPayload *struct{}
		wantErr     error
	}{
		{
			name:        "ok",
			hook:        hookThatReturns(t, somePayload, nil),
			wantPayload: &somePayload,
			wantErr:     nil,
		},
		{
			name:        "error",
			hook:        hookThatReturns(t, nil, gitlab.ErrParsingPayload),
			wantPayload: nil,
			wantErr:     gitlab.ErrParsingPayload,
		},
		{
			name:        "event not allowlisted",
			hook:        hookThatReturns(t, nil, gitlab.ErrEventNotFound),
			wantPayload: nil,
			wantErr:     nil,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewAdapter(tc.hook)

			payload, err := adapter.Parse(httptest.NewRequest("POST", "/webhook", nil))

			assert.Equal(t, tc.wantPayload, payload)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func hookThatReturns(t *testing.T, event interface{}, err error) hook[string] {
	t.Helper()

	h := NewMockhook[string](t)
	h.On("Parse", mock.Anything).Return(event, err)

	return h
}
