// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/notifier/preference"
)

// DefaultNotifier is the default implementation of the Notifier interface
type DefaultNotifier struct {
	transports  []Transport
	preferences preference.Preferences
}

func (d *DefaultNotifier) Push(ctx context.Context, notification event.Notification) error {
	results := make([]error, 0, len(d.transports))

	for _, transport := range d.transports {
		wants := d.preferences.Wants(ctx, notification, transport.Key())
		if wants == nil {
			slog.DebugContext(ctx, "no explicit preference for notification; sending anyway", "id", notification.Context().ID, "type", notification.Context().Type, "recipient", notification.Recipient().String(), "transport", transport.Key())
			// No explicit preference, we assume the user wants it
		} else if !*wants {
			slog.DebugContext(ctx, "user does not want this notification via this transport", "id", notification.Context().ID, "type", notification.Context().Type, "recipient", notification.Recipient().String(), "transport", transport.Key())
			continue // User does not want this transport
		}

		slog.InfoContext(ctx, "pushing notification to transport", "id", notification.Context().ID, "type", notification.Context().Type, "recipient", notification.Recipient().String(), "transport", transport.Key())
		if err := transport.Push(ctx, notification); err != nil {
			slog.ErrorContext(ctx, "failed to push notification via transport", "id", notification.Context().ID, "recipient", notification.Recipient().String(), "transport", transport.Key(), "error", err)
			results = append(results, fmt.Errorf("transport %s failed for notification %s: %w", transport.Key(), notification.Context().ID, err))
		} else {
			results = append(results, nil)
		}
	}

	if len(results) == 0 {
		slog.WarnContext(ctx, "notification not sent to any transport", "id", notification.Context().ID, "recipient", notification.Recipient().String())
	}

	return errors.Join(results...)
}

var _ Notifier = &DefaultNotifier{}

// New creates a new DefaultNotifier
func New(transports []Transport, preferences preference.Preferences) *DefaultNotifier {
	return &DefaultNotifier{
		transports:  transports,
		preferences: preferences,
	}
}
