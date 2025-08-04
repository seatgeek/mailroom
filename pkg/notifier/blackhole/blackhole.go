// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

// Package blackhole provides a notifier.Transport implementation that discards notifications
package blackhole

import (
	"context"
	"errors"
	"log/slog"

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/notifier"
)

// Transport supports "sending" messages to nowhere, effectively discarding them.
// This is useful when you want to create notifications but don't have a specific recipient,
// and want to avoid errors from undeliverable messages.
type Transport struct {
	key event.TransportKey
}

// Push "delivers" a notification to the black hole, effectively discarding it.
// It only accepts notifications with the BlackHoleDiscard identifier.
func (b *Transport) Push(ctx context.Context, notification event.Notification) error {
	if _, ok := notification.Recipient().Get(identifier.BlackHoleDiscard); !ok {
		return notifier.Permanent(errors.New("recipient does not have a black hole discard identifier"))
	}

	slog.DebugContext(ctx, "notification discarded to black hole",
		"id", notification.Context().ID,
		"type", notification.Context().Type,
		"transport", b.key)

	return nil
}

func (b *Transport) Key() event.TransportKey {
	return b.key
}

var _ notifier.Transport = &Transport{}

// NewTransport creates a new Black Hole Transport that discards notifications
func NewTransport(key event.TransportKey) *Transport {
	return &Transport{
		key: key,
	}
}
