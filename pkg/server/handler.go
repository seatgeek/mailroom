// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

// Package server provides the HTTP server for incoming events
package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/notifier"
)

// CreateEventProcessingHandler returns a handlerFunc that can be used to handle incoming webhooks.
// It choreographs the parsing of the incoming request, the generation of notifications, dispatching the notifications
// to the notifier, and returning a success or error response to the client.
func CreateEventProcessingHandler(parserKey string, parser event.Parser, processors []event.Processor, ntfr notifier.Notifier) http.HandlerFunc { //nolint:revive
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Debug("handling incoming webhook", "parser", parserKey, "path", request.URL.Path)

		evt, err := parser.Parse(request)
		if err != nil {
			logAndSendErrorResponse(writer, parserKey, "failed to parse event", err)
			return
		}

		if evt == nil { // Event is ignorable
			slog.Debug("ignoring uninteresting event", "parser", parserKey)
			http.Error(writer, "thanks but we're not interested in that event", 200)
			return
		}

		notifications := []event.Notification{}

		for _, processor := range processors {
			notifications, err = processor.Process(request.Context(), *evt, notifications)
			if err != nil {
				logAndSendErrorResponse(writer, parserKey, fmt.Sprintf("failed during processing (processor %T)", processor), err)
				return
			}
		}

		if len(notifications) == 0 {
			slog.Debug("no notifications to send after processing", "parser", parserKey, "eventID", evt.ID)
			http.Error(writer, "no notifications to send", 200)
			return
		}

		slog.Debug("dispatching notifications", "eventID", evt.ID, "parser", parserKey, "notifications_count", len(notifications))

		errorCount := 0
		for _, n := range notifications {
			if err = ntfr.Push(request.Context(), n); err != nil {
				errorCount++
				slog.WarnContext(request.Context(), "failed to push notification", "eventID", evt.ID, "notification_recipient", n.Recipient().String(), "parser", parserKey, "error", err)
			}
		}

		if errorCount > 0 {
			slog.WarnContext(request.Context(), "some notifications failed to send", "eventID", evt.ID, "parser", parserKey, "total_notifications", len(notifications), "failed_count", errorCount)
		}

		writer.WriteHeader(http.StatusAccepted)
		_, _ = writer.Write([]byte("Notifications enqueued\n"))
	}
}

func logAndSendErrorResponse(writer http.ResponseWriter, parserKey string, errorPrefix string, err error) {
	statusCode := 500

	var httpError *Error
	if errors.As(err, &httpError) {
		statusCode = httpError.Code
		err = httpError.Reason
	}

	if statusCode < 500 {
		slog.Warn(errorPrefix, "parser", parserKey, "error", err)
	} else {
		slog.Error(errorPrefix, "parser", parserKey, "error", err)
	}

	http.Error(writer, fmt.Sprintf("%s: %v", errorPrefix, err), statusCode)
}
