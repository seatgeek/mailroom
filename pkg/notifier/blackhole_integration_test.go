// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/notification"
	"github.com/seatgeek/mailroom/pkg/notifier"
	"github.com/seatgeek/mailroom/pkg/notifier/preference"
)

func TestBlackholeIntegration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("blackhole transport discards notification with only blackhole identifier", func(t *testing.T) {
		t.Parallel()

		// Create a buffer writer for the regular transport
		buffer := &bytes.Buffer{}
		writerTransport := notifier.NewWriterNotifier("writer", buffer)
		
		// Create a blackhole transport
		blackholeTransport := notifier.NewBlackholeTransport("blackhole")

		// Create a notifier with both transports
		defaultNotifier := notifier.New(
			[]notifier.Transport{writerTransport, blackholeTransport},
			preference.Default(true), // Allow all notifications
		)

		// Create a notification with ONLY a blackhole identifier
		eventCtx := event.Context{
			ID:     "test-blackhole-integration",
			Source: event.MustSource("/test"),
			Type:   "test.event",
		}
		
		notif := notification.NewBuilder(eventCtx).
			WithDefaultMessage("This should be discarded").
			WithRecipientIdentifiers(
				identifier.New("blackhole", "discard"),
			).
			Build()

		// Push the notification
		err := defaultNotifier.Push(ctx, notif)
		
		// Should succeed without error
		require.NoError(t, err)
		
		// Writer transport should also have received it (but it should write it normally)
		// The blackhole transport will also handle it by discarding
		assert.Contains(t, buffer.String(), "This should be discarded")
		assert.Contains(t, buffer.String(), "test-blackhole-integration")
	})

	t.Run("blackhole transport lets other transports handle notifications with mixed identifiers", func(t *testing.T) {
		t.Parallel()

		// Create a blackhole transport
		blackholeTransport := notifier.NewBlackholeTransport("blackhole")

		// Create a notifier with only blackhole transport
		defaultNotifier := notifier.New(
			[]notifier.Transport{blackholeTransport},
			preference.Default(true), // Allow all notifications
		)

		// Create a notification with mixed identifiers including blackhole
		eventCtx := event.Context{
			ID:     "test-mixed-identifiers",
			Source: event.MustSource("/test"),
			Type:   "test.event",
		}
		
		notif := notification.NewBuilder(eventCtx).
			WithDefaultMessage("This has mixed identifiers").
			WithRecipientIdentifiers(
				identifier.New("email", "user@example.com"),
				identifier.New("blackhole", "discard"),
				identifier.New("slack.com/id", "U123456"),
			).
			Build()

		// Push the notification
		err := defaultNotifier.Push(ctx, notif)
		
		// Should succeed without error (blackhole handles it but doesn't actively discard)
		require.NoError(t, err)
	})

	t.Run("normal transport handles notification without blackhole identifier", func(t *testing.T) {
		t.Parallel()

		// Create a buffer writer for the regular transport
		buffer := &bytes.Buffer{}
		writerTransport := notifier.NewWriterNotifier("writer", buffer)
		
		// Create a blackhole transport
		blackholeTransport := notifier.NewBlackholeTransport("blackhole")

		// Create a notifier with both transports
		defaultNotifier := notifier.New(
			[]notifier.Transport{writerTransport, blackholeTransport},
			preference.Default(true), // Allow all notifications
		)

		// Create a notification WITHOUT blackhole identifier
		eventCtx := event.Context{
			ID:     "test-normal-notification",
			Source: event.MustSource("/test"),
			Type:   "test.event",
		}
		
		notif := notification.NewBuilder(eventCtx).
			WithDefaultMessage("This should go to writer").
			WithRecipientIdentifiers(
				identifier.New("email", "user@example.com"),
			).
			Build()

		// Push the notification
		err := defaultNotifier.Push(ctx, notif)
		
		// Should succeed without error
		require.NoError(t, err)
		
		// Writer transport should have received the notification
		assert.Contains(t, buffer.String(), "This should go to writer")
		assert.Contains(t, buffer.String(), "test-normal-notification")
	})
}