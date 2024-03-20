// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"testing"

	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/identifier"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	user := New(
		WithIdentifier(identifier.New("username", "rufus")),
		WithIdentifiers(identifier.Collection{
			identifier.For("email"): "rufus@seatgeek.com",
		}),
		WithPreference("com.example.notification", "email", true),
	)

	wantIdentifiers := identifier.Collection{
		identifier.For("username"): "rufus",
		identifier.For("email"):    "rufus@seatgeek.com",
	}

	wantPreferences := map[common.EventType]map[common.TransportID]bool{
		"com.example.notification": {
			"email": true,
		},
	}

	assert.Equal(t, wantIdentifiers, user.Identifiers)
	assert.Equal(t, wantPreferences, user.preferences)
}

func TestUser_Wants(t *testing.T) {
	t.Parallel()

	user := New(
		WithPreference("com.example.notification", "email", true),
	)

	assert.True(t, user.Wants("com.example.notification", "email"))
	assert.False(t, user.Wants("com.example.notification", "slack"))
	assert.False(t, user.Wants("com.example.other", "email"))
}
