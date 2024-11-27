// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

// Package server provides the HTTP server for incoming events
package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/seatgeek/mailroom/pkg/handler"
	"github.com/seatgeek/mailroom/pkg/notifier"
)

// CreateEventHandler returns a handlerFunc that can be used to handle incoming webhooks
// It choreographs the parsing of the incoming request, the generation of notifications, dispatching the notifications
// to the notifier, and returning a success or error response to the client.
func CreateEventHandler(s handler.Handler, n notifier.Notifier) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Debug("handling incoming webhook", "handler", s.Key(), "path", request.URL.Path)

		notifications, err := s.Process(request)
		if err != nil {
			slog.Error("failed to generate notifications", "handler", s.Key(), "error", err)
			http.Error(writer, fmt.Sprintf("failed to generate notifications: %v", err), 500)
			return
		}

		if len(notifications) == 0 {
			slog.Debug("no notifications to send", "handler", s.Key())
			http.Error(writer, "thanks but we're not interested in that event", 200)
			return
		}

		id := notifications[0].Context().ID
		slog.Debug("dispatching notifications", "id", id, "handler", s.Key(), "notifications", len(notifications))

		var errs []error
		for _, notification := range notifications {
			err = n.Push(request.Context(), notification)
			if err != nil {
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			http.Error(writer, fmt.Sprintf("failed to send notifications: %v", errs), 500)
			return
		}
	}
}
