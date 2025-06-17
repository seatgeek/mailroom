// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package preference

import (
	"context"
	"errors"

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/validation"
)

// Preferences provides a mechanism for determining whether a user wants to receive some notification via some transport.
type Preferences interface {
	// Wants returns whether the user wants to receive the given event via the given transport.
	// It returns nil if there is no explicit preference.
	Wants(context.Context, event.Notification, event.TransportKey) *bool
}

// Func is a function type that implements the Preferences interface.
type Func func(context.Context, event.Notification, event.TransportKey) *bool

func (f Func) Wants(ctx context.Context, notification event.Notification, transport event.TransportKey) *bool {
	return f(ctx, notification, transport)
}

// Chain is a sequence of Preferences that will be checked in order until one returns a non-nil value.
type Chain []Preferences

var _ validation.Validator = (*Chain)(nil)

func (c Chain) Wants(ctx context.Context, notification event.Notification, transport event.TransportKey) *bool {
	for _, pref := range c {
		if result := pref.Wants(ctx, notification, transport); result != nil {
			return result
		}
	}

	// If no preferences matched, return nil to indicate no explicit preference.
	return nil
}

func (c Chain) Validate(ctx context.Context) error {
	errs := make([]error, 0, len(c))
	for _, pref := range c {
		if v, ok := pref.(validation.Validator); ok {
			if err := v.Validate(ctx); err != nil {
				errs = append(errs, err)
				continue
			}
		}
	}

	return errors.Join(errs...)
}

// Map defines preferences by event type and transport.
// For example, a user may want to receive PR review request notifications via Slack but not email.
type Map map[event.Type]map[event.TransportKey]bool

func (p Map) Wants(_ context.Context, notification event.Notification, transport event.TransportKey) *bool {
	eventType := notification.Context().Type

	if _, exists := p[eventType]; !exists {
		// No preference set for this event
		return nil
	}

	if want, exists := p[eventType][transport]; exists {
		return &want
	}

	// No preference set for this transport
	return nil
}

// Default returns a Preferences implementation that always returns the given boolean value.
func Default(wants bool) Func {
	return func(_ context.Context, _ event.Notification, _ event.TransportKey) *bool {
		return &wants
	}
}
