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
	id2 := identifier.New("gitlab.com/email", "colin.odell@seatgeek.com")
	id3 := identifier.New("email", "zhammer@seatgeek.com")

	userA := New(WithIdentifier(id1), WithIdentifier(id2))
	userB := New(WithIdentifier(id3))

	store := NewInMemoryStore(userA, userB)

	tests := []struct {
		name    string
		input   identifier.Identifier
		want    *User
		wantErr error
	}{
		{
			name:  "exact match (test 1)",
			input: id1,
			want:  userA,
		},
		{
			name:  "exact match (test 2)",
			input: id2,
			want:  userA,
		},
		{
			name:  "exact match (test 3)",
			input: id3,
			want:  userB,
		},
		{
			name:    "no match",
			input:   identifier.New("email", "rufus@seatgeek.com"),
			wantErr: ErrUserNotFound,
		},
		{
			name:  "fallback to any email",
			input: identifier.New("slack.com/email", "colin.odell@seatgeek.com"),
			want:  userA,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := store.Get(tc.input)

			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestInMemoryStore_Find(t *testing.T) {
	t.Parallel()

	id1 := identifier.New("email", "codell@seatgeek.com")
	id2 := identifier.New("gitlab.com/email", "colin.odell@seatgeek.com")
	id3 := identifier.New("email", "zhammer@seatgeek.com")

	userA := New(WithIdentifier(id1), WithIdentifier(id2))
	userB := New(WithIdentifier(id3))

	store := NewInMemoryStore(userA, userB)

	tests := []struct {
		name    string
		input   identifier.Collection
		want    *User
		wantErr error
	}{
		{
			name:  "exact match (test 1)",
			input: identifier.NewCollection(id1),
			want:  userA,
		},
		{
			name:  "exact match (test 2)",
			input: identifier.NewCollection(id2),
			want:  userA,
		},
		{
			name:  "exact match (test 3)",
			input: identifier.NewCollection(id3),
			want:  userB,
		},
		{
			name: "exact match (multiple inputs)",
			input: identifier.NewCollection(
				identifier.New("email", "foo@example.com"),
				identifier.New("email", "bar@example.com"),
				id1,
			),
			want: userA,
		},
		{
			name:    "no match",
			input:   identifier.NewCollection(identifier.New("email", "foo@example.com")),
			wantErr: ErrUserNotFound,
		},
		{
			name:  "fallback to any email",
			input: identifier.NewCollection(identifier.New("slack.com/email", "colin.odell@seatgeek.com")),
			want:  userA,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := store.Find(tc.input)

			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
