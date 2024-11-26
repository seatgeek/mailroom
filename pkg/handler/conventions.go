// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

// Generating notifications from incoming requests
package handler

import (
	"net/http"

	"github.com/seatgeek/mailroom/pkg/common"
	"github.com/seatgeek/mailroom/pkg/event"
)

// Handler is anything capable of generating notifications from an incoming HTTP request
type Handler interface {
	// Key is both a unique identifier for the handler, and the endpoint that it listens on
	Key() string
	// Process handles incoming webhooks, verifying the payloads and generating any number of notifications (or an error)
	Process(req *http.Request) ([]common.Notification, error)
	// EventTypes returns descriptors for all EventTypes that the handler may emit
	EventTypes() []event.TypeDescriptor
}
