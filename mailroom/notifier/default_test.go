// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier

import (
	"context"
	"errors"
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
		name          string
		notifications []*common.Notification
		transports    []Transport
		wantSent      []wantSent
		wantErrs      []error
	}{
		{
			name: "user wants all",
			notifications: []*common.Notification{
				{
					Type:      "com.example.one",
					Recipient: knownUser.Identifiers,
				},
				{
					Type:      "com.example.two",
					Recipient: knownUser.Identifiers,
				},
			},
			transports: []Transport{
				&fakeTransport{id: "email"},
				&fakeTransport{id: "slack"},
			},
			wantSent: []wantSent{
				{event: "com.example.one", transport: "email"},
				{event: "com.example.one", transport: "slack"},
				{event: "com.example.two", transport: "email"},
			},
		},
		{
			name:          "no notifications",
			notifications: []*common.Notification{},
			wantSent:      []wantSent{},
			wantErrs:      []error{},
		},
		{
			name: "unknown user",
			notifications: []*common.Notification{
				{
					Type:      "com.example.one",
					Recipient: unknownUser,
				},
				{
					Type:      "com.example.two",
					Recipient: unknownUser,
				},
			},
			wantSent: []wantSent{},
			wantErrs: []error{user.ErrUserNotFound, user.ErrUserNotFound},
		},
		{
			name: "user wants none",
			notifications: []*common.Notification{
				{
					Type:      "com.example.two",
					Recipient: knownUser.Identifiers,
				},
			},
			transports: []Transport{
				&fakeTransport{id: "slack"},
			},
			wantSent: []wantSent{},
		},
		{
			name: "transport fails",
			notifications: []*common.Notification{
				{
					Type:      "com.example.one",
					Recipient: knownUser.Identifiers,
				},
			},
			transports: []Transport{
				&fakeTransport{id: "email", returns: fmt.Errorf("failed to send email")},
			},
			wantSent: []wantSent{},
			wantErrs: []error{fmt.Errorf("failed to send email")},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			notifier := New(store, tc.transports...)

			errs := notifier.Push(context.Background(), tc.notifications...)

			if len(tc.wantErrs) == 0 {
				assert.NoError(t, errs)
			} else {
				assert.Equal(t, errors.Join(tc.wantErrs...), errs)
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

type fakeTransport struct {
	id      common.TransportID
	sent    []common.EventType
	returns error
}

var _ Transport = (*fakeTransport)(nil)

func (f *fakeTransport) ID() common.TransportID {
	return f.id
}

func (f *fakeTransport) Push(ctx context.Context, notifications ...*common.Notification) error {
	if f.returns != nil {
		return f.returns
	}

	for _, n := range notifications {
		f.sent = append(f.sent, n.Type)
	}

	return nil
}
