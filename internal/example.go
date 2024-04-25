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
	"strings"

	"github.com/lmittmann/tint"
	"github.com/seatgeek/mailroom/mailroom"
	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/identifier"
	"github.com/seatgeek/mailroom/mailroom/notification"
	"github.com/seatgeek/mailroom/mailroom/notifier"
	"github.com/seatgeek/mailroom/mailroom/source"
	"github.com/seatgeek/mailroom/mailroom/user"
)

type PlaygroundPayload struct {
	// identifier, e.g. "gitlab.com/id:123"
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
}

type PlaygroundParser struct{}

func (t *PlaygroundParser) Parse(req *http.Request) (any, error) {
	payload := PlaygroundPayload{}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return payload, nil
}

// PlaygroundGenerator is...
type PlaygroundGenerator struct{}

// Generate returns some dummy notifications
func (p *PlaygroundGenerator) Generate(payload any) ([]common.Notification, error) {
	body := payload.(PlaygroundPayload)
	ident := strings.Split(body.Recipient, ":")
	if len(ident) != 2 {
		return nil, fmt.Errorf("invalid recipient identifier: %s", body.Recipient)
	}

	return []common.Notification{
		notification.NewBuilder("a1c11a53-c4be-488f-89b6-f83bf2d48dab", "local.playground.message").WithDefaultMessage(body.Message).WithRecipientIdentifiers(identifier.New(ident[0], ident[1])).Build(),
	}, nil
}

func (p *PlaygroundGenerator) EventTypes() []common.EventTypeDescriptor {
	return []common.EventTypeDescriptor{
		{
			Key:         "local.playground.message",
			Title:       "Message",
			Description: "A message sent from to local playground",
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
		mailroom.WithSources(
			source.New(
				"playground",
				&PlaygroundParser{},
				&PlaygroundGenerator{},
			),
			// source.New(
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
			//	backoff.WithMaxElapsedTime(30*time.Second),
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
