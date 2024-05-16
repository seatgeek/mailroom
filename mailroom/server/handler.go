// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/seatgeek/mailroom/mailroom/handler"
	"github.com/seatgeek/mailroom/mailroom/notifier"
)

// CreateEventHandler returns a handlerFunc that can be used to handle incoming webhooks
// It choreographs the parsing of the incoming request, the generation of notifications, dispatching the notifications
// to the notifier, and returning a success or error response to the client.
func CreateEventHandler(ctx context.Context, s handler.Handler, n notifier.Notifier) handlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) error {
		slog.Debug("handling incoming webhook", "handler", s.Key(), "path", request.URL.Path)

		notifications, err := s.Process(request)
		if err != nil {
			slog.Error("failed to generate notifications", "handler", s.Key(), "error", err)
			return fmt.Errorf("failed to generate notifications: %w", err)
		}

		if len(notifications) == 0 {
			slog.Debug("no notifications to send", "handler", s.Key())
			writer.WriteHeader(200)
			_, _ = writer.Write([]byte("thanks but we're not interested in that event"))
			return nil
		}

		id := notifications[0].Context().ID
		slog.Debug("dispatching notifications", "id", id, "handler", s.Key(), "notifications", len(notifications))

		var errs []error
		for _, notification := range notifications {
			err = n.Push(ctx, notification)
			if err != nil {
				errs = append(errs, err)
			}
		}

		return errors.Join(errs...)
	}
}
