// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/seatgeek/mailroom/mailroom/identifier"
	"github.com/seatgeek/mailroom/mailroom/user"
	"github.com/seatgeek/mailroom/mailroom/user/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	pgtc "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPostgresStore_Add(t *testing.T) {
	t.Parallel()

	store := createDatastore(t)

	// Prove that our user doesn't exist in the database yet
	_, err := store.GetByIdentifier(identifier.New("email", "codell@seatgeek.com"))
	assert.ErrorIs(t, err, user.ErrUserNotFound)

	// Insert a new user
	u := user.New(
		"codell",
		user.WithIdentifier(identifier.New("email", "codell@seatgeek.com")),
	)

	err = store.Add(u)
	assert.NoError(t, err)

	// Check if inserted (by key)
	got, err := store.Get(u.Key)
	assert.NoError(t, err)
	assert.Equal(t, u, got)

	// Check if inserted (by identifier)
	got, err = store.GetByIdentifier(identifier.New("email", "codell@seatgeek.com"))
	assert.NoError(t, err)
	assert.Equal(t, u, got)

	// Update that same user object
	u.Identifiers.Add(identifier.New("gitlab.com/email", "codell@seatgeek.com"))

	err = store.Add(u)
	assert.NoError(t, err)

	// Check if updated (by key)
	got, err = store.Get(u.Key)
	assert.NoError(t, err)
	assert.Equal(t, u, got)

	// Check if updated (by identifier)
	got, err = store.GetByIdentifier(identifier.New("email", "codell@seatgeek.com"))
	assert.NoError(t, err)
	assert.Equal(t, u, got)

	// Update that user using a completely different object with the same key and different identifier
	u = user.New("codell", user.WithIdentifier(identifier.New("email", "codell@example.com")))

	err = store.Add(u)
	assert.NoError(t, err)

	// Check if updated (by key)
	got, err = store.Get(u.Key)
	assert.NoError(t, err)
	assert.Equal(t, u, got)

	// Check if updated (by identifier)
	got, err = store.GetByIdentifier(identifier.New("email", "codell@example.com"))
	assert.NoError(t, err)
	assert.Equal(t, u, got)

	// Check if old identifier is gone
	_, err = store.GetByIdentifier(identifier.New("email", "codell@seatgeek.com"))
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestPostgresStore_Get(t *testing.T) {
	t.Parallel()

	store := createDatastore(t)

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
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := store.Add(tc.expected)
			assert.NoError(t, err)

			got, err := store.Get(tc.key)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestPostgresStore_Find(t *testing.T) {
	t.Parallel()

	store := createDatastore(t)

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
			name: "find user by email (using fallback)",
			arg:  identifier.NewCollection(identifier.New("slack.com/email", "codell@seatgeek.com")),
			expected: user.New(
				"codell",
				user.WithIdentifier(identifier.New("gitlab.com/email", "codell@seatgeek.com")),
			),
		},
		{
			name:     "user not found",
			arg:      identifier.NewCollection(identifier.New("email", "bbecker")),
			expected: nil,
			wantErr:  user.ErrUserNotFound,
		},
		{
			name:     "user not found; no fallback emails",
			arg:      identifier.NewCollection(identifier.New("slack.com/id", "U123")),
			expected: nil,
			wantErr:  user.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
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

	store := createDatastore(t)

	duplicateIdentifier := identifier.New("email", "dup@dup.com")

	err := store.Add(user.New(
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
	//nolint:testifylint
	assert.ErrorAs(t, err, &wantErr)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
	assert.Nil(t, got)
}

func TestPostgresStore_GetByIdentifier(t *testing.T) {
	t.Parallel()

	store := createDatastore(t)

	tests := []struct {
		name     string
		arg      identifier.Identifier
		expected *user.User
		wantErr  error
	}{
		{
			name: "find user by email",
			arg:  identifier.New("email", "bbecker@seatgeek.com"),
			expected: user.New(
				"bckr",
				user.WithIdentifier(identifier.New("email", "bbecker@seatgeek.com")),
				user.WithIdentifier(identifier.New("gitlab.com/email", "bbecker@seatgeek.com")),
			),
		},
		{
			name: "find user by email (using fallback)",
			arg:  identifier.New("slack.com/email", "codell@seatgeek.com"),
			expected: user.New(
				"codell",
				user.WithIdentifier(identifier.New("gitlab.com/email", "codell@seatgeek.com")),
			),
		},
		{
			name:     "user not found",
			arg:      identifier.New("email", "bbecker"),
			expected: nil,
			wantErr:  user.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.expected != nil {
				err := store.Add(tc.expected)
				assert.NoError(t, err)
			}

			got, err := store.GetByIdentifier(tc.arg)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestPostgresStore_SetPreferences(t *testing.T) {
	t.Parallel()

	store := createDatastore(t)

	// Add a user with  no preferences so we can test setting them later
	userWithNoPreferences := user.New(
		"zach",
		user.WithIdentifier(identifier.New("email", "zhammer@seatgeek.com")),
	)
	err := store.Add(userWithNoPreferences)
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

func createDatastore(t *testing.T) *postgres.Store {
	t.Helper()

	ctx := context.Background()

	container, err := pgtc.RunContainer(ctx,
		testcontainers.WithImage("postgres:16.2"),
		pgtc.WithInitScripts("../../../test/initdb/init.sql"),
		pgtc.WithDatabase("mailroom"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	assert.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, container.Terminate(ctx))
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable", "application_name=test")
	assert.NoError(t, err)

	db, err := gorm.Open(pg.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)

	return postgres.NewPostgresStore(db)
}
