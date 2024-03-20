// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package main

import (
	"context"
	"os"

	"github.com/go-playground/webhooks/v6/gitlab"
	"github.com/seatgeek/mailroom/mailroom"
	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/identifier"
	"github.com/seatgeek/mailroom/mailroom/notifier"
	"github.com/seatgeek/mailroom/mailroom/source"
	"github.com/seatgeek/mailroom/mailroom/source/webhooks"
	"github.com/seatgeek/mailroom/mailroom/user"
)

type TemporaryNotificationGenerator struct{}

func (t *TemporaryNotificationGenerator) Generate(payload any) ([]*common.Notification, error) {
	return nil, nil
}

// This is an example of how to configure and run mailroom.
// Code should be un-commented as the features are implemented.
func main() {
	app := mailroom.New(
		mailroom.WithSources(
			source.New(
				"/gitlab",
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
				user.New(
					user.WithIdentifiers(identifier.Collection{
						identifier.GenericEmail:         "zhammer@seatgeek.com",
						identifier.For("gitlab.com/id"): "456",
						identifier.For("slack.com/id"):  "U7654",
					}),
					user.WithPreference("com.example.notification", "console", true),
				),
			),
		),
	)

	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
