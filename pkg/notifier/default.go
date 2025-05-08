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
	"github.com/seatgeek/mailroom/pkg/user"
)

// DefaultNotifier is the default implementation of the Notifier interface
type DefaultNotifier struct {
	userStore  user.Store
	transports []Transport
}

func (d *DefaultNotifier) Push(ctx context.Context, notification event.Notification) error {
	results := make([]error, 0, len(d.transports))

	recipientUser := d.getRecipientUserForNotification(ctx, notification)

	for _, transport := range d.transports {
		// todo: should a processor handle this instead?
		if recipientUser != nil && !recipientUser.Wants(notification.Context().Type, transport.Key()) {
			slog.DebugContext(ctx, "user does not want this notification via this transport", "id", notification.Context().ID, "recipient", recipientUser.String(), "transport", transport.Key())
			continue
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

func (d *DefaultNotifier) getRecipientUserForNotification(ctx context.Context, notification event.Notification) *user.User {
	if d.userStore == nil {
		return nil
	}

	usr, err := d.userStore.Find(ctx, notification.Recipient())
	if err == nil {
		return usr
	}
	if !errors.Is(err, user.ErrUserNotFound) {
		slog.WarnContext(ctx, "failed to find user for preference lookup (will proceed without specific prefs)", "recipient", notification.Recipient().String(), "error", err)
	}

	return nil
}

var _ Notifier = &DefaultNotifier{}

// New creates a new DefaultNotifier
func New(userStore user.Store, transports ...Transport) *DefaultNotifier {
	return &DefaultNotifier{
		userStore:  userStore,
		transports: transports,
	}
}
