// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier

import (
	"context"
	"fmt"
	"testing"

	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/identifier"
	"github.com/seatgeek/mailroom/mailroom/user"
	"github.com/stretchr/testify/assert"
)

type wantSent struct {
	event     common.EventType
	transport common.TransportID
}

func TestDefaultNotifier_Push(t *testing.T) {
	t.Parallel()

	unknownUser := identifier.Collection{}

	knownUser := user.New(
		user.WithIdentifier(identifier.New(identifier.GenericUsername, "rufus")),
		user.WithPreference("com.example.one", "email", true),
		user.WithPreference("com.example.one", "slack", true),
		user.WithPreference("com.example.two", "email", true),
		user.WithPreference("com.example.two", "slack", false),
	)

	store := user.NewInMemoryStore(knownUser)

	tests := []struct {
		name         string
		notification common.Notification
		transports   []Transport
		wantSent     []wantSent
		wantErrs     []error
	}{
		{
			name: "user wants all",
			notification: common.Notification{
				Type:      "com.example.one",
				Recipient: knownUser.Identifiers,
			},
			transports: []Transport{
				&fakeTransport{id: "email"},
				&fakeTransport{id: "slack"},
			},
			wantSent: []wantSent{
				{event: "com.example.one", transport: "email"},
				{event: "com.example.one", transport: "slack"},
			},
		},
		{
			name: "user wants some",
			notification: common.Notification{
				Type:      "com.example.two",
				Recipient: knownUser.Identifiers,
			},
			transports: []Transport{
				&fakeTransport{id: "email"},
				&fakeTransport{id: "slack"},
			},
			wantSent: []wantSent{
				{event: "com.example.two", transport: "email"},
			},
		},
		{
			name: "user wants none",
			notification: common.Notification{
				Type:      "com.example.three",
				Recipient: knownUser.Identifiers,
			},
			transports: []Transport{
				&fakeTransport{id: "slack"},
			},
			wantSent: []wantSent{},
		},
		{
			name: "unknown user",
			notification: common.Notification{
				Type:      "com.example.one",
				Recipient: unknownUser,
			},
			wantSent: []wantSent{},
			wantErrs: []error{user.ErrUserNotFound, user.ErrUserNotFound},
		},
		{
			name: "transport fails",
			notification: common.Notification{
				Type:      "com.example.one",
				Recipient: knownUser.Identifiers,
			},
			transports: []Transport{
				&fakeTransport{id: "email", returns: errSomethingFailed},
			},
			wantSent: []wantSent{},
			wantErrs: []error{errSomethingFailed},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			notifier := New(store, tc.transports...)

			errs := notifier.Push(context.Background(), tc.notification)

			if len(tc.wantErrs) == 0 {
				assert.NoError(t, errs)
			} else {
				for _, wantErr := range tc.wantErrs {
					assert.ErrorIs(t, errs, wantErr)
				}
			}

			assertSent(t, tc.wantSent, tc.transports)
		})
	}
}

func assertSent(t *testing.T, want []wantSent, transports []Transport) {
	t.Helper()

	for _, w := range want {
		matched := false
		for _, transport := range transports {
			if transport.ID() == w.transport {
				matched = true
				assert.Contains(t, transport.(*fakeTransport).sent, w.event)
				continue
			}
		}

		if !matched {
			assert.Failf(t, "expected transport not found", "transport %q not found", w.transport)
		}
	}
}

var errSomethingFailed = fmt.Errorf("some transport error occurred")

type fakeTransport struct {
	id      common.TransportID
	sent    []common.EventType
	returns error
}

var _ Transport = (*fakeTransport)(nil)

func (f *fakeTransport) ID() common.TransportID {
	return f.id
}

func (f *fakeTransport) Push(ctx context.Context, notification common.Notification) error {
	if f.returns != nil {
		return f.returns
	}

	f.sent = append(f.sent, notification.Type)

	return nil
}
