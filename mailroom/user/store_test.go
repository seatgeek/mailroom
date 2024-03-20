// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"testing"

	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryStore(t *testing.T) {
	t.Parallel()

	user1 := New(common.NewIdentifier("email", "codell@seatgeek.com"))
	user2 := New(common.NewIdentifier("email", "zhammer@seatgeek.com"))

	store := NewInMemoryStore(user1)
	store.Add(user2)

	// Test that we can retrieve the users
	retrievedUser1, err := store.Get(common.NewIdentifier("email", "codell@seatgeek.com"))
	assert.Equal(t, user1, retrievedUser1)
	assert.NoError(t, err)

	retrievedUser2, err := store.Get(common.NewIdentifier("email", "zhammer@seatgeek.com"))
	assert.Equal(t, user2, retrievedUser2)
	assert.NoError(t, err)

	// Test that we can't retrieve a user that doesn't exist
	_, err = store.Get(common.NewIdentifier("email", "rufus@seatgeek.com"))
	assert.Error(t, err)
}
