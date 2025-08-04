// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/lmittmann/tint"
	"github.com/seatgeek/mailroom"
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/notification"
	"github.com/seatgeek/mailroom/pkg/notifier"
	"github.com/seatgeek/mailroom/pkg/user"
)

type BlackholeExampleEvent struct {
	Message      string `json:"message"`
	SendToEmail  string `json:"send_to_email,omitempty"`  // Optional: send to specific email
	UseBlackhole bool   `json:"use_blackhole,omitempty"` // If true, use blackhole identifier
}

var blackholeExampleType = event.Type("com.example.blackhole_demo")

type BlackholeExampleParser struct{}

func (p *BlackholeExampleParser) EventTypes() []event.TypeDescriptor {
	return []event.TypeDescriptor{
		{
			Key:         blackholeExampleType,
			Title:       "Blackhole Demo",
			Description: "Demonstrates blackhole identifier functionality",
		},
	}
}

func (p *BlackholeExampleParser) Parse(req *http.Request) (*event.Event, error) {
	payload := BlackholeExampleEvent{}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		return nil, err
	}

	evt := &event.Event{
		Context: event.Context{
			ID:     event.ID(uuid.New().String()),
			Source: event.MustSource("/blackhole/example"),
			Type:   blackholeExampleType,
		},
		Data: payload,
	}
	return evt, nil
}

type BlackholeNotificationGenerator struct{}

func (p *BlackholeNotificationGenerator) Process(ctx context.Context, evt event.Event, notifications []event.Notification) ([]event.Notification, error) {
	payload, ok := evt.Data.(BlackholeExampleEvent)
	if !ok {
		return nil, fmt.Errorf("unexpected event data type for BlackholeNotificationGenerator: %T", evt.Data)
	}

	builder := notification.NewBuilder(evt.Context).
		WithDefaultMessage(payload.Message)

	// If a specific email is provided, use it
	if payload.SendToEmail != "" {
		builder = builder.WithRecipientIdentifiers(identifier.New("email", payload.SendToEmail))
	}

	// If blackhole is requested, add blackhole identifier
	if payload.UseBlackhole {
		builder = builder.WithRecipientIdentifiers(identifier.New("blackhole", "discard"))
	}

	// If neither email nor blackhole is specified, default to blackhole
	if payload.SendToEmail == "" && !payload.UseBlackhole {
		builder = builder.WithRecipientIdentifiers(identifier.New("blackhole", "discard"))
	}

	newNotification := builder.Build()
	return append(notifications, newNotification), nil
}

// This example demonstrates the blackhole identifier/transport functionality.
func main() {
	// Turn on pretty logging
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level: slog.LevelDebug,
		}),
	))

	userStore := user.NewInMemoryStore(
		user.New(
			"codell",
			user.WithIdentifier(identifier.New("email", "codell@seatgeek.com")),
		),
	)

	app := mailroom.New(
		mailroom.WithParserAndGenerator("blackhole-example", &BlackholeExampleParser{}, &BlackholeNotificationGenerator{}),
		mailroom.WithProcessors(
			user.NewIdentifierEnrichmentProcessor(userStore),
		),
		mailroom.WithTransports(
			notifier.NewWriterNotifier("console", os.Stderr),
			// Add the blackhole transport - it should typically be last in the list
			notifier.NewBlackholeTransport("blackhole"),
		),
		mailroom.WithUserStore(userStore),
	)

	slog.Info("Starting blackhole example server on :8080")
	slog.Info("Try these example requests:")
	slog.Info(`curl -X POST http://localhost:8080/blackhole-example -H "Content-Type: application/json" -d '{"message": "This will be discarded", "use_blackhole": true}'`)
	slog.Info(`curl -X POST http://localhost:8080/blackhole-example -H "Content-Type: application/json" -d '{"message": "This will go to console", "send_to_email": "codell@seatgeek.com"}'`)
	slog.Info(`curl -X POST http://localhost:8080/blackhole-example -H "Content-Type: application/json" -d '{"message": "Default behavior - blackhole"}'`)

	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}