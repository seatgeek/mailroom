// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier

import (
	"context"
	"log/slog"

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
)

// BlackholeTransport is a transport that discards notifications instead of delivering them.
// It is designed to handle notifications that contain blackhole identifiers, ensuring
// that no error is returned when a notification would otherwise go undelivered.
type BlackholeTransport struct {
	key event.TransportKey
}

var _ Transport = &BlackholeTransport{}

func (b *BlackholeTransport) Key() event.TransportKey {
	return b.key
}

func (b *BlackholeTransport) Push(ctx context.Context, n event.Notification) error {
	// Check if the notification recipient contains a blackhole identifier
	hasBlackhole := false
	hasOtherIdentifiers := false
	
	for _, id := range n.Recipient().ToList() {
		if id.NamespaceAndKind.Kind() == identifier.KindBlackhole {
			hasBlackhole = true
		} else {
			hasOtherIdentifiers = true
		}
	}

	// Only handle notifications that contain blackhole identifiers
	// If there are other identifiers, let other transports handle them
	if hasBlackhole {
		if hasOtherIdentifiers {
			slog.DebugContext(ctx, "notification has blackhole and other identifiers, will be handled by other transports", 
				"id", n.Context().ID, 
				"recipient", n.Recipient().String(),
				"transport", b.key)
		} else {
			slog.InfoContext(ctx, "notification discarded by blackhole transport", 
				"id", n.Context().ID, 
				"type", n.Context().Type, 
				"recipient", n.Recipient().String(),
				"transport", b.key)
		}
		return nil // Successfully "delivered" to nowhere
	}

	// If no blackhole identifier is found, this transport should not handle the notification
	slog.DebugContext(ctx, "notification does not contain blackhole identifier, skipping", 
		"id", n.Context().ID, 
		"recipient", n.Recipient().String(),
		"transport", b.key)
	return nil
}

// NewBlackholeTransport creates a new BlackholeTransport with the given key.
func NewBlackholeTransport(key event.TransportKey) *BlackholeTransport {
	return &BlackholeTransport{key: key}
}