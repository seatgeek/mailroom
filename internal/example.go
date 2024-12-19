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
	"github.com/seatgeek/mailroom/pkg/common"
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/handler"
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

type ExampleHandler struct{}

var _ handler.Handler = &ExampleHandler{}

func (h *ExampleHandler) Key() string {
	return "example"
}

func (h *ExampleHandler) Process(req *http.Request) ([]common.Notification, error) {
	payload := MessageSentEvent{}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return []common.Notification{
		notification.NewBuilder(
			event.Context{
				ID:   event.ID(uuid.New().String()),
				Type: "local.playground.message",
			}).
			WithDefaultMessage(fmt.Sprintf("%s sent you a message: '%s'", payload.AuthorName, payload.Comment)).
			WithRecipientIdentifiers(identifier.New("email", payload.RecipientEmail)).
			Build(),
	}, nil
}

func (h *ExampleHandler) EventTypes() []event.TypeDescriptor {
	return []event.TypeDescriptor{
		{
			Key:         messageSentType,
			Title:       "Message",
			Description: "A message sent from one user to another",
		},
	}
}

// This is an example of how to configure and run mailroom.
// Code should be un-commented as the features are implemented.
func main() {
	// Turn on pretty logging
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level: slog.LevelDebug,
		}),
	))

	app := mailroom.New(
		mailroom.WithHandlers(
			&ExampleHandler{},
			// handler.New[ArgoEventPayload](
			//  "argocd",
			//	argocd.NewPayloadParser(
			//		argocd.WithEvents(argocd.AppSyncFailedEvent, argocd.AppSyncSucceededEvent),
			//	),
			//	argocd.NewNotificationGenerator(),
			// ),
		),
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
		mailroom.WithUserStore(
			user.NewInMemoryStore(
				user.New(
					"codell",
					user.WithIdentifier(identifier.New("email", "codell@seatgeek.com")),
					user.WithIdentifier(identifier.New("gitlab.com/id", "123")),
					user.WithIdentifier(identifier.New("slack.com/id", "U4567")),
					user.WithPreference("com.example.notification", "console", true),
				),
			),
		),
	)

	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
