// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package source

import (
	"net/http"

	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/event"
)

// Source is anything capable of generating notifications from an incoming HTTP request
type Source interface {
	// Key is both a unique identifier for the source, and the endpoint that it listens on
	Key() string
	// Parse verifies and parses incoming webhooks, returning any number of notifications or an error
	Parse(req *http.Request) ([]common.Notification, error)
	// EventTypes returns descriptors for all EventTypes that the source may emit
	EventTypes() []event.TypeDescriptor
}
