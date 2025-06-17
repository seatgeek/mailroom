// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier_test

import (
	"context"
	"errors"
	"testing"

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/notification"
	"github.com/seatgeek/mailroom/pkg/notifier"
	"github.com/seatgeek/mailroom/pkg/notifier/preference"
	"github.com/seatgeek/mailroom/pkg/user"
	"github.com/stretchr/testify/assert"
)

const someEventType = "com.example.one"

type wantSent struct {
	event     event.Type
	transport event.TransportKey
}

func TestDefaultNotifier_Push(t *testing.T) {
	t.Parallel()

	someUser := user.New(
		"rufus",
		user.WithIdentifier(identifier.New(identifier.GenericUsername, "rufus")),
	)

	prefs := preference.Map{
		someEventType: {
			"email": true,
			"slack": false,
		},
	}

	tests := []struct {
		name         string
		notification event.Notification
		transports   []notifier.Transport
		wantSent     []wantSent
		wantErrs     []error
	}{
		{
			name:         "delivers to all transports enabled by preferences",
			notification: notificationFor(someEventType, someUser.Identifiers),
			transports: []notifier.Transport{
				&fakeTransport{key: "email"},
				&fakeTransport{key: "slack"},
				&fakeTransport{key: "sms"},
			},
			wantSent: []wantSent{
				{event: someEventType, transport: "email"}, // explicitly enabled
				{event: someEventType, transport: "sms"},   // not explicitly disabled
			},
		},
		{
			name:         "transport fails",
			notification: notificationFor(someEventType, someUser.Identifiers),
			transports: []notifier.Transport{
				&fakeTransport{key: "email", returns: errSomethingFailed},
			},
			wantSent: []wantSent{},
			wantErrs: []error{errSomethingFailed},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			notifier := notifier.New(tc.transports, prefs)

			errs := notifier.Push(t.Context(), tc.notification)

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

func assertSent(t *testing.T, want []wantSent, transports []notifier.Transport) {
	t.Helper()

	for _, w := range want {
		matched := false
		for _, transport := range transports {
			if transport.Key() == w.transport {
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

var errSomethingFailed = errors.New("some transport error occurred")

type fakeTransport struct {
	key     event.TransportKey
	sent    []event.Type
	returns error
}

var _ notifier.Transport = (*fakeTransport)(nil)

func (f *fakeTransport) Key() event.TransportKey {
	return f.key
}

func (f *fakeTransport) Push(_ context.Context, notification event.Notification) error {
	if f.returns != nil {
		return f.returns
	}

	f.sent = append(f.sent, notification.Context().Type)

	return nil
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
