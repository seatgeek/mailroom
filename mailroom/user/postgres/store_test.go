// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package postgres_test

import (
	"errors"
	"os"
	"testing"

	"github.com/seatgeek/mailroom/mailroom/identifier"
	"github.com/seatgeek/mailroom/mailroom/user"
	"github.com/seatgeek/mailroom/mailroom/user/postgres"
	"github.com/stretchr/testify/assert"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	dsn = os.Getenv("POSTGRES_DSN")
)

func TestPostgresStore_Get(t *testing.T) {
	t.Parallel()

	if dsn == "" {
		t.Skip("POSTGRES_DSN is not set")
	}

	db, err := gorm.Open(pg.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	store := postgres.NewPostgresStore(db)

	tests := []struct {
		name     string
		key      string
		expected *user.User
	}{
		{
			name: "user with preferences and identifiers",
			key:  "codell",
			expected: user.New(
				"codell",
				user.WithIdentifier(identifier.New("email", "codell@seatgeek.com")),
				user.WithIdentifier(identifier.New("gitlab.com/email", "colin.odell@seatgeek.com")),
				user.WithPreference("com.gitlab.push", "email", false),
				user.WithPreference("com.gitlab.push", "slack", true),
				user.WithPreference("com.argocd.sync-succeeded", "email", true),
				user.WithPreference("com.argocd.sync-succeeded", "slack", true),
			),
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err = store.Add(tc.expected)
			assert.NoError(t, err)

			got, err := store.Get(tc.key)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestPostgresStore_Find(t *testing.T) {
	t.Parallel()

	if dsn == "" {
		t.Skip("POSTGRES_DSN is not set")
	}

	db, err := gorm.Open(pg.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	store := postgres.NewPostgresStore(db)

	tests := []struct {
		name     string
		arg      identifier.Collection
		expected *user.User
		wantErr  error
	}{
		{
			name: "find user by email",
			arg:  identifier.NewCollection(identifier.New("email", "bbecker@seatgeek.com")),
			expected: user.New(
				"bckr",
				user.WithIdentifier(identifier.New("email", "bbecker@seatgeek.com")),
				user.WithIdentifier(identifier.New("gitlab.com/email", "bbecker@seatgeek.com")),
			),
		},
		{
			name:     "user not found",
			arg:      identifier.NewCollection(identifier.New("email", "bbecker")),
			expected: nil,
			wantErr:  user.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.expected != nil {
				err := store.Add(tc.expected)
				assert.NoError(t, err)
			}

			got, err := store.Find(tc.arg)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestPostgresStore_Find_duplicate(t *testing.T) {
	t.Parallel()

	if dsn == "" {
		t.Skip("POSTGRES_DSN is not set")
	}

	db, err := gorm.Open(pg.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	store := postgres.NewPostgresStore(db)

	duplicateIdentifier := identifier.New("email", "dup@dup.com")

	err = store.Add(user.New(
		"duplicateA",
		user.WithIdentifier(duplicateIdentifier),
	))
	assert.NoError(t, err)

	err = store.Add(user.New(
		"duplicateb",
		user.WithIdentifier(duplicateIdentifier),
	))
	assert.NoError(t, err)

	got, err := store.Find(identifier.NewCollection(duplicateIdentifier))

	wantErr := errors.New("found multiple users with identifiers: [email:dup@dup.com]")
	assert.ErrorAs(t, err, &wantErr)
	assert.Nil(t, got)
}

func TestPostgresStore_SetPreferences(t *testing.T) {
	t.Parallel()

	if dsn == "" {
		t.Skip("POSTGRES_DSN is not set")
	}

	db, err := gorm.Open(pg.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	store := postgres.NewPostgresStore(db)

	// Add a user with  no preferences so we can test setting them later
	userWithNoPreferences := user.New(
		"zach",
		user.WithIdentifier(identifier.New("email", "zhammer@seatgeek.com")),
	)
	err = store.Add(userWithNoPreferences)
	assert.NoError(t, err)

	// Set preferences
	expectedPreferences := user.Preferences{
		"com.example.notification": {
			"email": true,
			"slack": false,
		},
	}
	err = store.SetPreferences(userWithNoPreferences.Key, expectedPreferences)
	assert.NoError(t, err)

	// Check if set
	expectedUser := user.New(
		userWithNoPreferences.Key,
		user.WithIdentifiers(userWithNoPreferences.Identifiers),
		user.WithPreferences(expectedPreferences),
	)
	got, err := store.Get(userWithNoPreferences.Key)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, got)
}
