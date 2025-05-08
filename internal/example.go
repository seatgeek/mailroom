// Copyright 2024 SeatGeek, Inc.
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

type MessageSentEvent struct {
	AuthorName     string `json:"from"`
	RecipientEmail string `json:"to"`
	Comment        string `json:"comment"`
}

var messageSentType = event.Type("com.example.message_sent")

type ExampleParser struct{}

func (p *ExampleParser) Key() string {
	return "example"
}

func (p *ExampleParser) EventTypes() []event.TypeDescriptor {
	return []event.TypeDescriptor{
		{
			Key:         messageSentType,
			Title:       "Message",
			Description: "A message sent from one user to another",
		},
	}
}

func (p *ExampleParser) Parse(req *http.Request) (*event.Event, error) {
	payload := MessageSentEvent{}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		return nil, err
	}

	evt := &event.Event{
		Context: event.Context{
			ID:     event.ID(uuid.New().String()),
			Source: event.MustSource("/example/parser"),
			Type:   messageSentType,
		},
		Data: payload,
	}
	return evt, nil
}

type NotificationGenerator struct{}

func (p *NotificationGenerator) Process(ctx context.Context, evt event.Event, notifications []event.Notification) ([]event.Notification, error) {
	payload, ok := evt.Data.(MessageSentEvent)
	if !ok {
		return nil, fmt.Errorf("unexpected event data type for NotificationGenerator: %T", evt.Data)
	}

	newNotification := notification.NewBuilder(evt.Context).
		WithDefaultMessage(fmt.Sprintf("%s sent you a message: '%s'", payload.AuthorName, payload.Comment)).
		WithRecipientIdentifiers(identifier.New("email", payload.RecipientEmail)).
		Build()

	return append(notifications, newNotification), nil
}

// This is an example of how to configure and run mailroom.
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
			user.WithIdentifier(identifier.New("gitlab.com/id", "123")),
			user.WithIdentifier(identifier.New("slack.com/id", "U4567")),
			user.WithPreference("com.example.notification", "console", true),
		),
	)

	app := mailroom.New(
		mailroom.WithEventSource(&ExampleParser{}, &NotificationGenerator{}, user.NewIdentifierEnrichmentProcessor(userStore)),
		// TODO: Add other EventSources and their specific processors, e.g.:
		// argoParser := argo.NewParser(...)
		// argoNotificationGenerator := argo.NewNotificationGeneratorProc(...)
		// mailroom.WithEventSource(argoParser, argoNotificationGenerator),

		mailroom.WithTransports(
			notifier.NewWriterNotifier("console", os.Stderr),
			// notifier.WithRetry(
			//	notifier.WithTimeout(
			//		slack.NewTransport(
			//			"slack",
			//			"xoxb-1234567890-1234567890123-AbCdEfGhIjKlMnOpQrStUvWx",
			//		),
			//		5*time.Second,
			//	),
			//	3,
			//	func() notifier.BackOff {
			//		return backoff.NewExponentialBackOff()
			//	},
			// ),
		),
		mailroom.WithUserStore(userStore),
	)

	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
