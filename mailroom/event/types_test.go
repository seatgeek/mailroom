// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package event_test

import (
	"testing"

	"github.com/seatgeek/mailroom/mailroom/event"
	"github.com/stretchr/testify/assert"
)

func TestSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		uri     string
		wantNil bool
	}{
		{
			name: "valid uri with DNS authority",
			uri:  "https://www.example.com/foo",
		},
		{
			name: "mailto",
			uri:  "mailto:codell@seatgeek.com",
		},
		{
			name: "universally-unique URN with a UUID",
			uri:  "urn:uuid:6e8bc430-9c3a-11d9-9669-0800200c9a66",
		},
		{
			name: "application-specific identifier",
			uri:  "/cloudevents/spec/pull/123",
		},
		{
			name:    "empty uri",
			uri:     "",
			wantNil: true,
		},
		{
			name:    "invalid uri",
			uri:     ":This is not a URI:",
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			source := event.NewSource(tc.uri)

			if tc.wantNil {
				assert.Nil(t, source)
				return
			}

			assert.NotNil(t, source)
			assert.Equal(t, tc.uri, source.String())
		})
	}
}
