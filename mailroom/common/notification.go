// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package common

import "github.com/seatgeek/mailroom/mailroom/identifier"

// EventType describes the type of event related to the originating occurrence.
// It may be used for routing, observability, etc. It must comply with CloudEvent `type` spec:
// https://github.com/cloudevents/spec/blob/main/cloudevents/spec.md#type
// Basically, it should be a non-empty string containing a reverse-DNS name.
// For example: "com.gitlab.push"
type EventType string

// Notification is a notification that should be sent
type Notification struct {
	Type      EventType
	Message   Renderer
	Initiator identifier.Collection
	Recipient identifier.Collection
}

// Renderer is a type that can render a message, potentially customizing it for a given transport
// For example, a Slack message might include :emoji: or Markdown formatting, while an email message might use HTML.
// If the given transport is not recognized, the Renderer should return a plain text message suitable for any transport.
type Renderer interface {
	Render(transport TransportID) string
}

// RendererFunc is a shortcut for quickly creating a Renderer
type RendererFunc func(transport TransportID) string

func (f RendererFunc) Render(transport TransportID) string {
	return f(transport)
}

// TransportID is a type that identifies a specific type of transport for sending notifications
type TransportID string // eg. "slack"; "email"
