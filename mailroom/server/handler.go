// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/seatgeek/mailroom/mailroom/notifier"
	"github.com/seatgeek/mailroom/mailroom/source"
)

// CreateHandler returns a handlerFunc that can be used to handle incoming webhooks
// It choreographs the parsing of the incoming request, the generation of notifications, dispatching the notifications
// to the notifier, and returning a success or error response to the client.
func CreateHandler(ctx context.Context, s *source.Source, n notifier.Notifier) handlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) error {
		payload, err := s.Parser.Parse(request)
		if err != nil {
			slog.Error("failed to parse payload", "source", s.ID, "error", err)
			return err
		}

		if payload == nil {
			slog.Debug("ignoring uninteresting event", "source", s.ID)
			writer.WriteHeader(200)
			_, _ = writer.Write([]byte("thanks but we're not interested in that event"))
			return nil
		}

		notifications, err := s.Generator.Generate(*payload)
		if err != nil {
			slog.Error("failed to generate notifications", "source", s.ID, "error", err)
			return fmt.Errorf("failed to generate notifications: %w", err)
		}

		return n.Push(ctx, notifications...)
	}
}
