// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package common

import (
	"context"

	"github.com/seatgeek/mailroom/mailroom/event"
	"github.com/seatgeek/mailroom/mailroom/identifier"
)

// Notification is a notification that should be sent
type Notification interface {
	Context() event.Context
	Recipient() identifier.Set
	Render(TransportKey) string
}

// TransportKey is a type that identifies a specific type of transport for sending notifications
type TransportKey string // eg. "slack"; "email"

// Validator can be implemented by any parser, generator, transport, etc. to validate its configuration at runtime
// Errors returned by Validate are considered fatal
type Validator interface {
	Validate(ctx context.Context) error
}
