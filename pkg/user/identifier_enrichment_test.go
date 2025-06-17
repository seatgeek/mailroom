// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user_test

import (
	"errors"
	"testing"

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/user"
	"github.com/stretchr/testify/assert"
)

func TestIdentifierEnrichmentProcessor_Process(t *testing.T) {
	t.Parallel()

	evt := event.Event{Context: event.Context{ID: "test-event"}}

	id1 := identifier.New("email", "test@example.com")
	id2 := identifier.New("slack", "U123")
	id3 := identifier.New("github", "testuser")

	testCases := []struct {
		name           string
		notifications  []event.Notification
		mockStoreSetup func(mockStore *user.MockStore)
		expect         func(t *testing.T, result []event.Notification)
		expectedError  error
	}{
		{
			name: "user found and identifiers merged",
			notifications: []event.Notification{
				notificationFor("some-event", identifier.NewSet(id1)),
			},
			mockStoreSetup: func(mockStore *user.MockStore) {
				foundUser := user.New("test-user", user.WithIdentifier(id2), user.WithIdentifier(id3))
				mockStore.On("Find", t.Context(), identifier.NewSet(id1)).Return(foundUser, nil).Once()
			},
			expect: func(t *testing.T, result []event.Notification) {
				t.Helper()
				assert.Len(t, result, 1)
				assert.ElementsMatch(t, identifier.NewSet(id1, id2, id3).ToList(), result[0].Recipient().ToList())
			},
		},
		{
			name: "multiple notifications with mixed results",
			notifications: []event.Notification{
				notificationFor("some-event", identifier.NewSet(id1)),
				notificationFor("some-event", nil),
				notificationFor("some-event", identifier.NewSet(id2)),
				notificationFor("some-event", identifier.NewSet(id3)),
			},
			mockStoreSetup: func(mockStore *user.MockStore) {
				foundUser := user.New("test-user-1", user.WithIdentifier(identifier.New("new", "val1")))
				mockStore.On("Find", t.Context(), identifier.NewSet(id1)).Return(foundUser, nil).Once()

				mockStore.On("Find", t.Context(), identifier.NewSet(id2)).Return(nil, user.ErrUserNotFound).Once()
				mockStore.On("Find", t.Context(), identifier.NewSet(id3)).Return(nil, errors.New("some db error")).Once()
			},
			expect: func(t *testing.T, result []event.Notification) {
				t.Helper()
				assert.Len(t, result, 4)
				assert.ElementsMatch(t, identifier.NewSet(id1, identifier.New("new", "val1")).ToList(), result[0].Recipient().ToList())
				assert.Nil(t, result[1].Recipient())
				assert.ElementsMatch(t, identifier.NewSet(id2).ToList(), result[2].Recipient().ToList())
				assert.ElementsMatch(t, identifier.NewSet(id3).ToList(), result[3].Recipient().ToList())
			},
		},
		{
			name: "recipient is nil",
			notifications: []event.Notification{
				notificationFor("some-event", nil),
			},
			expect: func(t *testing.T, result []event.Notification) {
				t.Helper()
				assert.Len(t, result, 1)
				assert.Nil(t, result[0].Recipient())
			},
		},
		{
			name:          "no notifications",
			notifications: []event.Notification{},
			expect: func(t *testing.T, result []event.Notification) {
				t.Helper()
				assert.Empty(t, result)
			},
		},
		{
			name: "user not found",
			notifications: []event.Notification{
				notificationFor("some-event", identifier.NewSet(id1)),
			},
			mockStoreSetup: func(mockStore *user.MockStore) {
				mockStore.On("Find", t.Context(), identifier.NewSet(id1)).Return(nil, user.ErrUserNotFound).Once()
			},
			expect: func(t *testing.T, result []event.Notification) {
				t.Helper()
				assert.Len(t, result, 1)
				assert.ElementsMatch(t, identifier.NewSet(id1).ToList(), result[0].Recipient().ToList())
			},
		},
		{
			name: "error finding user",
			notifications: []event.Notification{
				notificationFor("some-event", identifier.NewSet(id1)),
			},
			mockStoreSetup: func(mockStore *user.MockStore) {
				mockStore.On("Find", t.Context(), identifier.NewSet(id1)).Return(nil, errors.New("some db error")).Once()
			},
			expect: func(t *testing.T, result []event.Notification) {
				t.Helper()
				assert.Len(t, result, 1)
				assert.ElementsMatch(t, identifier.NewSet(id1).ToList(), result[0].Recipient().ToList())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockStoreSetup := func(mockStore *user.MockStore) {}
			if tc.mockStoreSetup != nil {
				mockStoreSetup = tc.mockStoreSetup
			}

			mockUserStore := user.NewMockStore(t)
			mockStoreSetup(mockUserStore)

			processor := user.NewIdentifierEnrichmentProcessor(mockUserStore)

			resultNotifications, err := processor.Process(t.Context(), evt, tc.notifications)

			assert.Equal(t, tc.expectedError, err)
			tc.expect(t, resultNotifications)
			mockUserStore.AssertExpectations(t)
		})
	}
}

func TestNewIdentifierEnrichmentProcessor_NilStore(t *testing.T) {
	t.Parallel()

	assert.PanicsWithValue(t, "user.Store cannot be nil for IdentifierEnrichmentProcessor", func() {
		user.NewIdentifierEnrichmentProcessor(nil)
	})
}
