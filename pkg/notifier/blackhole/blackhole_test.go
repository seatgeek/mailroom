// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package blackhole

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/notification"
	"github.com/seatgeek/mailroom/pkg/notifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransport_Key(t *testing.T) {
	t.Parallel()

	transport := NewTransport("test-blackhole")
	assert.Equal(t, event.TransportKey("test-blackhole"), transport.Key())
}

func TestTransport_Push_Success(t *testing.T) {
	t.Parallel()

	transport := NewTransport("blackhole")
	ctx := context.Background()

	eventCtx := event.Context{
		ID:     event.ID(uuid.New().String()),
		Source: event.MustSource("/test"),
		Type:   "test.event",
	}

	notif := notification.NewBuilder(eventCtx).
		WithDefaultMessage("test message").
		WithRecipientIdentifiers(identifier.New(identifier.BlackHoleDiscard, "any-value")).
		Build()

	err := transport.Push(ctx, notif)
	assert.NoError(t, err)
}

func TestTransport_Push_RejectsWrongIdentifier(t *testing.T) {
	t.Parallel()

	transport := NewTransport("blackhole")
	ctx := context.Background()

	eventCtx := event.Context{
		ID:     event.ID(uuid.New().String()),
		Source: event.MustSource("/test"),
		Type:   "test.event",
	}

	tests := []struct {
		name       string
		identifier identifier.Identifier
	}{
		{
			name:       "email identifier",
			identifier: identifier.New("email", "test@example.com"),
		},
		{
			name:       "slack identifier",
			identifier: identifier.New("slack.com/id", "U123456"),
		},
		{
			name:       "generic username",
			identifier: identifier.New(identifier.GenericUsername, "testuser"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			notif := notification.NewBuilder(eventCtx).
				WithDefaultMessage("test message").
				WithRecipientIdentifiers(tc.identifier).
				Build()

			err := transport.Push(ctx, notif)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "does not have a black hole discard identifier")
		})
	}
}

func TestTransport_Push_AcceptsMultipleIdentifiersWithBlackHole(t *testing.T) {
	t.Parallel()

	transport := NewTransport("blackhole")
	ctx := context.Background()

	eventCtx := event.Context{
		ID:     event.ID(uuid.New().String()),
		Source: event.MustSource("/test"),
		Type:   "test.event",
	}

	notif := notification.NewBuilder(eventCtx).
		WithDefaultMessage("test message").
		WithRecipientIdentifiers(
			identifier.New("email", "test@example.com"),
			identifier.New(identifier.BlackHoleDiscard, "discard"),
		).
		Build()

	err := transport.Push(ctx, notif)
	assert.NoError(t, err)
}

func TestTransport_ImplementsInterface(t *testing.T) {
	t.Parallel()

	var _ notifier.Transport = &Transport{}
}
