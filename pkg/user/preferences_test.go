// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user_test

import (
	"errors"
	"testing"

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/notification"
	"github.com/seatgeek/mailroom/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPreferenceProvider_Wants(t *testing.T) {
	t.Parallel()

	someUser := user.New(
		"rufus",
		user.WithIdentifier(identifier.New(identifier.GenericUsername, "rufus")),
		user.WithPreference("com.example.one", "slack", true),
		user.WithPreference("com.example.two", "email", false),
	)

	tests := []struct {
		name         string
		notification event.Notification
		transport    event.TransportKey
		userStore    user.Store
		wants        *bool
	}{
		{
			name:         "known user; wants",
			notification: notificationFor("com.example.one", someUser.Identifiers),
			transport:    "slack",
			userStore:    user.NewInMemoryStore(someUser),
			wants:        ptrTo(true),
		},
		{
			name:         "known user; does not want",
			notification: notificationFor("com.example.two", someUser.Identifiers),
			transport:    "email",
			userStore:    user.NewInMemoryStore(someUser),
			wants:        ptrTo(false),
		},
		{
			name:         "known user; no preference",
			notification: notificationFor("com.example.three", someUser.Identifiers),
			transport:    "smoke-signal",
			userStore:    user.NewInMemoryStore(someUser),
			wants:        nil,
		},
		{
			name:         "unknown user",
			notification: notificationFor("com.example.one", someUser.Identifiers),
			transport:    "email",
			userStore:    user.NewInMemoryStore(), // empty store
			wants:        nil,
		},
		{
			name:         "no user store provided for some reason",
			notification: notificationFor("com.example.one", someUser.Identifiers),
			transport:    "email",
			userStore:    nil,
			wants:        nil, // should not panic, just return nil
		},
		{
			name:         "user store returns error",
			notification: notificationFor("com.example.one", someUser.Identifiers),
			transport:    "email",
			userStore:    userStoreThatErrors(t, errors.New("database connection error")),
			wants:        nil, // should not panic, just return nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			provider := user.NewPreferenceProvider(tt.userStore)

			wants := provider.Wants(t.Context(), tt.notification, tt.transport)
			assert.Equal(t, tt.wants, wants)
		})
	}
}

func notificationFor(eventType event.Type, identifiers identifier.Set) event.Notification {
	return notification.NewBuilder(
		event.Context{
			ID:   event.ID(eventType),
			Type: eventType,
		}).
		WithRecipient(identifiers).
		WithDefaultMessage("hello world").
		Build()
}

func ptrTo(b bool) *bool {
	return &b
}

func userStoreThatErrors(t *testing.T, err error) user.Store {
	t.Helper()

	store := user.NewMockStore(t)
	store.EXPECT().Find(mock.Anything, mock.Anything).Return(nil, err).Maybe()

	return store
}
