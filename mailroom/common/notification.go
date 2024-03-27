// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package common

import (
	"context"

	"github.com/seatgeek/mailroom/mailroom/identifier"
)

// EventType describes the type of event related to the originating occurrence.
// It may be used for routing, observability, etc. It must comply with CloudEvent `type` spec:
// https://github.com/cloudevents/spec/blob/main/cloudevents/spec.md#type
// Basically, it should be a non-empty string containing a reverse-DNS name.
// For example: "com.gitlab.push"
type EventType string

// Notification is a notification that should be sent
type Notification interface {
	Type() EventType
	Recipient() identifier.Collection
	Render(TransportID) string
	AddRecipients(identifier.Collection)
}

// TransportID is a type that identifies a specific type of transport for sending notifications
type TransportID string // eg. "slack"; "email"

// Validator can be implemented by any parser, generator, transport, etc. to validate its configuration at runtime
// Errors returned by Validate are considered fatal
type Validator interface {
	Validate(ctx context.Context) error
}
