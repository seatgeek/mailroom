// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"testing"

	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/event"
	"github.com/seatgeek/mailroom/mailroom/identifier"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	user := New(
		"rufus",
		WithIdentifier(identifier.New("username", "rufus")),
		WithIdentifiers(identifier.NewCollection(
			identifier.New("email", "rufus@seatgeek.com"),
		)),
		WithPreference("com.example.notification", "email", true),
	)

	wantIdentifiers := []identifier.Identifier{
		identifier.New("username", "rufus"),
		identifier.New("email", "rufus@seatgeek.com"),
	}

	wantPreferences := Preferences{
		"com.example.notification": {
			"email": true,
		},
	}

	assert.ElementsMatch(t, wantIdentifiers, user.Identifiers.ToList())
	assert.Equal(t, wantPreferences, user.Preferences)
}

func TestUser_Wants(t *testing.T) {
	t.Parallel()

	user := New(
		"rufus",
		WithPreference("com.example.notification", "email", true),
		WithPreference("com.example.notification", "slack", false),
	)

	tests := []struct {
		name      string
		event     event.Type
		transport common.TransportKey
		expected  bool
	}{
		{
			name:      "preference explicitly set to true",
			event:     "com.example.notification",
			transport: "email",
			expected:  true,
		},
		{
			name:      "preference explicitly set to false",
			event:     "com.example.notification",
			transport: "slack",
			expected:  false,
		},
		{
			name:      "preference not defined for transport",
			event:     "com.example.notification",
			transport: "smoke_signal",
			expected:  true,
		},
		{
			name:      "preference not defined for event",
			event:     "com.example.other",
			transport: "email",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, user.Wants(tt.event, tt.transport))
		})
	}
}
