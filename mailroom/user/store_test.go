// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"testing"

	"github.com/seatgeek/mailroom/mailroom/identifier"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryStore_Get(t *testing.T) {
	t.Parallel()

	id1 := identifier.New("email", "codell@seatgeek.com")
	id2 := identifier.New("email", "zhammer@seatgeek.com")

	user1 := New(WithIdentifier(id1))
	user2 := New(WithIdentifier(id2))

	store := NewInMemoryStore(user1, user2)

	// Test that we can retrieve the users
	retrievedUser1, err := store.Get(id1)
	assert.Equal(t, user1, retrievedUser1)
	assert.NoError(t, err)

	retrievedUser2, err := store.Get(id2)
	assert.Equal(t, user2, retrievedUser2)
	assert.NoError(t, err)

	// Test that we can't retrieve a user that doesn't exist
	_, err = store.Get(identifier.New("email", "rufus@seatgeek.com"))
	assert.Error(t, err)
}

func TestInMemoryStore_Find(t *testing.T) {
	t.Parallel()

	id1 := identifier.New("email", "codell@seatgeek.com")
	id2 := identifier.New("email", "zhammer@seatgeek.com")

	user1 := New(WithIdentifier(id1))
	user2 := New(WithIdentifier(id2))

	store := NewInMemoryStore(user1, user2)

	// Test that we can find the users
	retrievedUser1, err := store.Find(user1.Identifiers)
	assert.Equal(t, user1, retrievedUser1)
	assert.NoError(t, err)

	retrievedUser2, err := store.Find(user2.Identifiers)
	assert.Equal(t, user2, retrievedUser2)
	assert.NoError(t, err)

	// Test that we can't find a user that doesn't exist
	_, err = store.Find(identifier.NewCollection(identifier.New("email", "rufus@seatgeek.com")))
	assert.Error(t, err)
}
