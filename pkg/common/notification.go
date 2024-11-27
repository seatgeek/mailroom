// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

// Package common contains common types used throughout the framework
package common

import (
	"context"

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
)

// Notification is a notification that should be sent
type Notification interface {
	// Context provides the metadata for the notification
	Context() event.Context
	// Recipient returns the intended recipient of the notification
	Recipient() identifier.Set
	// Render returns the message to be sent via the given transport
	Render(TransportKey) string
}

// TransportKey is a type that identifies a specific type of transport for sending notifications
type TransportKey string // eg. "slack"; "email"

// Validator can be implemented by any parser, generator, transport, etc. to validate its configuration at runtime
type Validator interface {
	// Validate should return an error if the configuration is invalid
	// Errors returned by Validate are considered fatal
	Validate(ctx context.Context) error
}
