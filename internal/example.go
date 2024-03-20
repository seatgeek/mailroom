// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package main

import (
	"context"

	"github.com/go-playground/webhooks/v6/gitlab"
	"github.com/seatgeek/mailroom/mailroom"
	"github.com/seatgeek/mailroom/mailroom/common"
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
		// mailroom.WithTransports(
		//	transport.New(
		//		"slack",
		//		slack.NewTransport(
		//			slack.WithToken("xoxb-1234567890-1234567890123-AbCdEfGhIjKlMnOpQrStUvWx"),
		//		),
		//	),
		// ),
		mailroom.WithUserStore(
			user.NewInMemoryStore(
				user.New(
					common.NewIdentifier("email", "codell@seatgeek.com"),
					common.NewIdentifier("gitlab.com/user_id", "123"),
					common.NewIdentifier("slack.com/user_id", "U4567"),
				),
				user.New(
					common.NewIdentifier("email", "zhammer@seatgeek.com"),
					common.NewIdentifier("gitlab.com/user_id", "999"),
					common.NewIdentifier("slack.com/user_id", "U9876"),
				),
			),
		),
	)

	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
