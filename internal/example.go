// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/go-playground/webhooks/v6/gitlab"
	"github.com/lmittmann/tint"
	"github.com/seatgeek/mailroom/mailroom"
	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/identifier"
	"github.com/seatgeek/mailroom/mailroom/notifier"
	"github.com/seatgeek/mailroom/mailroom/source"
	"github.com/seatgeek/mailroom/mailroom/source/webhooks"
	"github.com/seatgeek/mailroom/mailroom/user"
)

// TemporaryNotificationGenerator is a temporary implementation of the source.NotificationGenerator interface
// TODO: Replace this with a real implementation
type TemporaryNotificationGenerator struct{}

// Generate returns some dummy notifications
func (t *TemporaryNotificationGenerator) Generate(payload any) ([]common.Notification, error) {
	return []common.Notification{}, nil
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
				"gitlab",
				webhooks.NewAdapter(
					webhooks.Must(gitlab.New(gitlab.Options.Secret("SomeSecretToValidatePayloads"))),
					gitlab.MergeRequestEvents,
				),
				&TemporaryNotificationGenerator{},
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
			// slack.NewTransport(
			//	"slack",
			//	"xoxb-1234567890-1234567890123-AbCdEfGhIjKlMnOpQrStUvWx",
			// ),
		),
		mailroom.WithUserStore(
			user.NewInMemoryStore(
				user.New(
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
