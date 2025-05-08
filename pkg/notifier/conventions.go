// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

// Package notifier encapsulates the concept of sending a common.Notification via some Transport
package notifier

import (
	"context"

	"github.com/cenkalti/backoff/v5"
	"github.com/seatgeek/mailroom/pkg/event"
)

// Permanent wraps the given err as a permanent error which should not be retried
func Permanent(err error) error {
	return backoff.Permanent(err)
}

// Func is a function that sends a notification
type Func func(context.Context, event.Notification) error

// Notifier is an interface for sending notifications
// Implementations of this interface may send them to a Transport, enqueue them for later,
// batch multiple notifications together, etc.
// Using the decorator design pattern to add functionality to a Notifier is encouraged!
// (https://refactoring.guru/design-patterns/decorator)
type Notifier interface {
	// Push sends a notification, either immediately or enqueued for later delivery
	// It SHOULD return an error wrapped by Permanent if we know that retries will never succeed
	Push(context.Context, event.Notification) error
}

// Transport is any notifier with a distinct, named key.
// The key is used to route notifications to the correct transport.
// You will typically have one Transport implementation per notification service. For example: one for Slack, another for email, etc.
type Transport interface {
	Notifier
	// Key returns a unique identifier for this transport, useful for routing purposes
	Key() event.TransportKey
}
