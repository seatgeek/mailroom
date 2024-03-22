// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier

import (
	"context"
	"errors"

	"github.com/seatgeek/mailroom/mailroom/common"
)

var ErrPermanentFailure = errors.New("permanent failure")

// Notifier is an interface for sending notifications
// Implementations of this interface may send them to a Transport, enqueue them for later,
// batch multiple notifications together, etc.
// Using the decorator design pattern to add functionality to a Notifier is encouraged!
// (https://refactoring.guru/design-patterns/decorator)
type Notifier interface {
	// Push sends a notification, either immediately or enqueued for later delivery
	// It SHOULD return an error containing ErrPermanentFailure if we know that retries will never succeed
	Push(context.Context, common.Notification) error
}

// Transport is any notifier with a distinct, named ID
type Transport interface {
	Notifier
	// ID returns a unique identifier for this transport, useful for routing purposes
	ID() common.TransportID
}
