// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier

import (
	"context"

	"github.com/seatgeek/mailroom/mailroom/common"
)

// Notifier is an interface for sending notifications
// Implementations of this interface may be specific transports (eg. Slack, Email), or perhaps a queueing system that
// sends notifications to multiple transports, or anything else.
// Using the decorator design pattern to add functionality to a Notifier is encouraged!
// (https://refactoring.guru/design-patterns/decorator)
type Notifier interface {
	// Push sends one or more notifications, either immediately or enqueued for later delivery
	Push(context.Context, ...*common.Notification) error
}

type Transport interface {
	Notifier
	// ID returns a unique identifier for this transport
	ID() common.TransportID
}
