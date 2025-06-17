// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"testing"

	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/notifier/preference"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	user := New(
		"rufus",
		WithIdentifier(identifier.New("username", "rufus")),
		WithIdentifiers(identifier.NewSet(
			identifier.New("email", "rufus@seatgeek.com"),
		)),
		WithPreference("com.example.notification", "email", true),
	)

	wantIdentifiers := []identifier.Identifier{
		identifier.New("username", "rufus"),
		identifier.New("email", "rufus@seatgeek.com"),
	}

	wantPreferences := preference.Map{
		"com.example.notification": {
			"email": true,
		},
	}

	assert.ElementsMatch(t, wantIdentifiers, user.Identifiers.ToList())
	assert.Equal(t, wantPreferences, user.Preferences)
}

func TestUser_String(t *testing.T) {
	t.Parallel()

	user := New(
		"rufus",
		WithIdentifier(identifier.New("username", "rufus")),
		WithIdentifier(identifier.New("email", "rufus@seatgeek.com")),
	)

	assert.Equal(t, "[email:rufus@seatgeek.com username:rufus]", user.String())
}
