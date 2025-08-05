// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

// Package server provides the HTTP server for incoming events
package server

import (
	"context"
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
		logger := slog.With(
			slog.String("parser", parserKey),
			slog.String("path", request.URL.Path),
		)

		logger.DebugContext(request.Context(), "handling incoming webhook")

		evt, err := parser.Parse(request)
		if err != nil {
			logAndSendErrorResponse(request.Context(), logger, writer, "failed to parse event", err)
			return
		}

		if evt == nil { // Event is ignorable
			logger.DebugContext(request.Context(), "ignoring uninteresting event")
			http.Error(writer, "thanks but we're not interested in that event", 200)
			return
		}

		logger = logger.With(slog.String("event_id", string(evt.ID)))

		notifications := []event.Notification{}

		for _, processor := range processors {
			notifications, err = processor.Process(request.Context(), *evt, notifications)
			if err != nil {
				logAndSendErrorResponse(request.Context(), logger, writer, fmt.Sprintf("failed during processing (processor %T)", processor), err)
				return
			}
		}

		if len(notifications) == 0 {
			logger.DebugContext(request.Context(), "no notifications to send after processing")
			http.Error(writer, "no notifications to send", 200)
			return
		}

		logger = logger.With(slog.Int("total_notifications", len(notifications)))

		logger.DebugContext(request.Context(), "dispatching notifications")

		errorCount := 0
		for _, n := range notifications {
			if err = ntfr.Push(request.Context(), n); err != nil {
				errorCount++
				logger.WarnContext(request.Context(), "failed to push notification", "notification_recipient", n.Recipient().String(), "error", err)
			}
		}

		if errorCount > 0 {
			logger.WarnContext(request.Context(), "some notifications failed to send", "failed_count", errorCount)
		}

		writer.WriteHeader(http.StatusAccepted)
		_, _ = writer.Write([]byte("Notifications enqueued\n"))
	}
}

func logAndSendErrorResponse(ctx context.Context, logger *slog.Logger, writer http.ResponseWriter, errorPrefix string, err error) {
	statusCode := 500

	var httpError *Error
	if errors.As(err, &httpError) {
		statusCode = httpError.Code
		err = httpError.Reason
	}

	if statusCode < 500 {
		logger.WarnContext(ctx, errorPrefix, "error", err)
	} else {
		logger.ErrorContext(ctx, errorPrefix, "error", err)
	}

	http.Error(writer, fmt.Sprintf("%s: %v", errorPrefix, err), statusCode)
}
