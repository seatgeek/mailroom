// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package source

import (
	"net/http"

	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/event"
)

// PayloadParser is an interface for anything that parses incoming webhooks
type PayloadParser interface {
	// Parse verifies and parses incoming webhooks and returns a well-defined payload object or an error
	//
	// The payload return value contains the parsed payload struct, which will be passed to the NotificationGenerator.
	// Returning nil, nil is valid and indicates that the payload was parsed and determined to not be allowlisted, and
	// thus should be ignored.
	Parse(req *http.Request) (payload any, err error)
}

// NotificationGenerator is an interface for anything that generates notifications from a parsed payload
type NotificationGenerator interface {
	// Generate takes a payload and returns a list of Notifications to be sent
	//
	// Some payloads may result in multiple notifications, for example the creation of a new merge request in GitLab
	// might result in notifications to multiple reviewers.
	Generate(payload any) ([]common.Notification, error)
	// EventTypes returns descriptors for all EventTypes that the generator may emit
	EventTypes() []event.TypeDescriptor
}

// Source is a combination of a PayloadParser and a NotificationGenerator
// Both are required to be able to generate notifications from incoming webhooks, but they are kept separate to allow
// users to easily override the default generator with a custom one if needed.
type Source struct {
	// Key is both a unique identifier for the source, and the endpoint that it listens on
	Key       string
	Parser    PayloadParser
	Generator NotificationGenerator
}

// New returns a new Source, pairing a PayloadParser and a NotificationGenerator together with some key
func New(key string, parser PayloadParser, generator NotificationGenerator) *Source {
	return &Source{
		Key:       key,
		Parser:    parser,
		Generator: generator,
	}
}
