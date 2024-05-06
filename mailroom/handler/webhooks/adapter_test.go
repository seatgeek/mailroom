// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package webhooks

import (
	"net/http/httptest"
	"testing"

	"github.com/go-playground/webhooks/v6/gitlab"
	"github.com/seatgeek/mailroom/mailroom/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdapter_Parse(t *testing.T) {
	t.Parallel()

	somePayload := struct{}{}

	tests := []struct {
		name      string
		hook      hook[string]
		wantEvent *event.Event[any]
		wantErr   error
	}{
		{
			name: "ok",
			hook: hookThatReturns(t, somePayload, nil),
			wantEvent: &event.Event[any]{
				Data: somePayload,
			},
			wantErr: nil,
		},
		{
			name:      "error",
			hook:      hookThatReturns(t, nil, gitlab.ErrParsingPayload),
			wantEvent: nil,
			wantErr:   gitlab.ErrParsingPayload,
		},
		{
			name:      "event not allowlisted",
			hook:      hookThatReturns(t, nil, gitlab.ErrEventNotFound),
			wantEvent: nil,
			wantErr:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewAdapter(tc.hook)

			payload, err := adapter.Parse(httptest.NewRequest("POST", "/webhook", nil))

			if tc.wantEvent == nil {
				assert.Nil(t, payload)
			} else {
				assert.Equal(t, tc.wantEvent.Data, payload.Data)
			}

			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func hookThatReturns(t *testing.T, event interface{}, err error) hook[string] {
	t.Helper()

	h := NewMockhook[string](t)
	h.EXPECT().Parse(mock.Anything, mock.Anything).Return(event, err)

	return h
}
