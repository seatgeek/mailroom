// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"context"
	"errors"
	"log/slog"

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/notifier/preference"
)

type PreferenceProvider struct {
	userStore Store
}

var _ preference.Provider = (*PreferenceProvider)(nil)

func NewPreferenceProvider(userStore Store) *PreferenceProvider {
	return &PreferenceProvider{
		userStore: userStore,
	}
}

func (p PreferenceProvider) Wants(ctx context.Context, notification event.Notification, transport event.TransportKey) *bool {
	recipientUser := p.getRecipientUserForNotification(ctx, notification)
	if recipientUser == nil {
		return nil
	}

	return recipientUser.Preferences.Wants(ctx, notification, transport)
}

func (p PreferenceProvider) getRecipientUserForNotification(ctx context.Context, notification event.Notification) *User {
	if p.userStore == nil {
		return nil
	}

	usr, err := p.userStore.Find(ctx, notification.Recipient())
	if err == nil {
		return usr
	}
	if !errors.Is(err, ErrUserNotFound) {
		slog.WarnContext(ctx, "failed to find user for preference lookup", "recipient", notification.Recipient().String(), "error", err)
	}

	return nil
}
