// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier

import (
	"context"
	"errors"
	"fmt"

	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/user"
)

// DefaultNotifier is the default implementation of the Notifier interface
// It routes notifications to the appropriate transport based on the user's preferences
type DefaultNotifier struct {
	userStore  user.Store
	transports []Transport
}

func (d *DefaultNotifier) Push(ctx context.Context, notification common.Notification) error {
	var errs []error

	recipientUser, err := d.userStore.Find(notification.Recipient)
	if err != nil {
		return fmt.Errorf("failed to find recipient user: %w", err)
	}

	for _, transport := range d.transports {
		if !recipientUser.Wants(notification.Type, transport.ID()) {
			continue
		}

		if err = transport.Push(ctx, notification); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

var _ Notifier = &DefaultNotifier{}

func New(userStore user.Store, transports ...Transport) *DefaultNotifier {
	return &DefaultNotifier{
		userStore:  userStore,
		transports: transports,
	}
}
