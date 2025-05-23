// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier

import (
	"context"
	"errors"
	"log/slog"

	"github.com/seatgeek/mailroom/pkg/common"
	"github.com/seatgeek/mailroom/pkg/user"
)

// DefaultNotifier is the default implementation of the Notifier interface
// It routes notifications to the appropriate transport based on the user's preferences
type DefaultNotifier struct {
	userStore  user.Store
	transports []Transport
}

func (d *DefaultNotifier) Push(ctx context.Context, notification common.Notification) error {
	var errs []error

	// The store may know of other identifiers for this user, so we merge those in
	recipientUser, err := d.userStore.Find(ctx, notification.Recipient())
	if err == nil {
		notification.Recipient().Merge(recipientUser.Identifiers)
	} else {
		slog.Debug("failed to find user", "id", notification.Context().ID, "user", notification.Recipient().String(), "error", err)
	}

	for _, transport := range d.transports {
		if recipientUser != nil && !recipientUser.Wants(notification.Context().Type, transport.Key()) {
			slog.Debug("user does not want this notification this way", "id", notification.Context().ID, "user", recipientUser.String(), "transport", transport.Key())
			continue
		}

		slog.Info("pushing notification", "id", notification.Context().ID, "type", notification.Context().Type, "user", recipientUser.String(), "transport", transport.Key())
		if err = transport.Push(ctx, notification); err != nil {
			slog.Error("failed to push notification", "id", notification.Context().ID, "user", recipientUser, "transport", transport.Key(), "error", err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
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
