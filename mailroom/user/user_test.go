// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"testing"

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
	)

	wantIdentifiers := identifier.Collection{
		identifier.For("username"): "rufus",
		identifier.For("email"):    "rufus@seatgeek.com",
	}

	assert.Equal(t, wantIdentifiers, user.Identifiers)
}
