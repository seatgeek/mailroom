// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier

import (
	"context"

	"github.com/seatgeek/mailroom/pkg/event"
)

// notifier provides a shortcut for using a function as a Notifier
type notifier struct {
	pushFunc Func
}

func (n notifier) Push(ctx context.Context, notification event.Notification) error {
	return n.pushFunc(ctx, notification)
}

// NewNotifier creates a new Notifier from a push function
func NewNotifier(pushFunc Func) Notifier {
	return &notifier{
		pushFunc: pushFunc,
	}
}

// transport provides a shortcut for using a function as a Transport
type transport struct {
	key      event.TransportKey
	pushFunc Func
}

func (t transport) Key() event.TransportKey {
	return t.key
}

func (t transport) Push(ctx context.Context, notification event.Notification) error {
	return t.pushFunc(ctx, notification)
}

// NewTransport creates a new Transport from a key and a push function
func NewTransport(key event.TransportKey, pushFunc Func) Transport {
	return &transport{
		key:      key,
		pushFunc: pushFunc,
	}
}
