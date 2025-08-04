// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/notification"
	"github.com/seatgeek/mailroom/pkg/notifier"
)

func TestBlackholeTransport_Key(t *testing.T) {
	t.Parallel()

	transport := notifier.NewBlackholeTransport("test-blackhole")
	assert.Equal(t, event.TransportKey("test-blackhole"), transport.Key())
}

func TestBlackholeTransport_Push(t *testing.T) {
	t.Parallel()

	transport := notifier.NewBlackholeTransport("blackhole")
	ctx := context.Background()

	t.Run("handles notification with mixed identifiers including blackhole", func(t *testing.T) {
		t.Parallel()

		// Create a notification with a blackhole identifier
		eventCtx := event.Context{
			ID:     "test-event-1",
			Source: event.MustSource("/test"),
			Type:   "test.event",
		}
		
		notification := notification.NewBuilder(eventCtx).
			WithDefaultMessage("test message").
			WithRecipientIdentifiers(
				identifier.New("email", "user@example.com"),
				identifier.New("blackhole", "discard"),
			).
			Build()

		err := transport.Push(ctx, notification)
		assert.NoError(t, err)
	})

	t.Run("discards notification with only blackhole identifier", func(t *testing.T) {
		t.Parallel()

		// Create a notification with only a blackhole identifier
		eventCtx := event.Context{
			ID:     "test-event-2",
			Source: event.MustSource("/test"),
			Type:   "test.event",
		}
		
		notification := notification.NewBuilder(eventCtx).
			WithDefaultMessage("test message").
			WithRecipientIdentifiers(
				identifier.New("blackhole", "discard"),
			).
			Build()

		err := transport.Push(ctx, notification)
		assert.NoError(t, err)
	})

	t.Run("handles notification without blackhole identifier", func(t *testing.T) {
		t.Parallel()

		// Create a notification without a blackhole identifier
		eventCtx := event.Context{
			ID:     "test-event-3",
			Source: event.MustSource("/test"),
			Type:   "test.event",
		}
		
		notification := notification.NewBuilder(eventCtx).
			WithDefaultMessage("test message").
			WithRecipientIdentifiers(
				identifier.New("email", "user@example.com"),
				identifier.New("slack.com/id", "U123456"),
			).
			Build()

		err := transport.Push(ctx, notification)
		assert.NoError(t, err) // Should still succeed but do nothing
	})

	t.Run("handles notification with namespaced blackhole identifier", func(t *testing.T) {
		t.Parallel()

		// Create a notification with a namespaced blackhole identifier
		eventCtx := event.Context{
			ID:     "test-event-4",
			Source: event.MustSource("/test"),
			Type:   "test.event",
		}
		
		notification := notification.NewBuilder(eventCtx).
			WithDefaultMessage("test message").
			WithRecipientIdentifiers(
				identifier.New("example.com/blackhole", "discard"),
			).
			Build()

		err := transport.Push(ctx, notification)
		assert.NoError(t, err)
	})
}