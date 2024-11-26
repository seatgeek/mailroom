// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notification_test

import (
	"testing"

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/notification"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	builder := notification.NewBuilder(event.Context{
		ID:   "a1c11a53-c4be-488f-89b6-f83bf2d48dab",
		Type: "com.example.test",
	})

	empty := builder.Build()
	assert.Equal(t, event.ID("a1c11a53-c4be-488f-89b6-f83bf2d48dab"), empty.Context().ID)
	assert.Equal(t, event.Type("com.example.test"), empty.Context().Type)
	assert.Empty(t, empty.Recipient().ToList())
	assert.Empty(t, empty.Render("email"))

	builderWithRecipient := builder.WithRecipientIdentifiers(identifier.New(identifier.GenericUsername, "codell"))

	withRecipient := builderWithRecipient.Build()
	assert.Equal(t, event.ID("a1c11a53-c4be-488f-89b6-f83bf2d48dab"), empty.Context().ID)
	assert.Equal(t, event.Type("com.example.test"), withRecipient.Context().Type)
	assert.Len(t, withRecipient.Recipient().ToList(), 1)
	assert.Equal(t, "codell", withRecipient.Recipient().MustGet(identifier.GenericUsername))
	assert.Empty(t, withRecipient.Render("email"))

	builderWithDefaultMessage := builder.WithDefaultMessage("Hello, world!")

	withDefaultMessage := builderWithDefaultMessage.Build()
	assert.Equal(t, event.ID("a1c11a53-c4be-488f-89b6-f83bf2d48dab"), empty.Context().ID)
	assert.Equal(t, event.Type("com.example.test"), withDefaultMessage.Context().Type)
	assert.Len(t, withRecipient.Recipient().ToList(), 1)
	assert.Equal(t, "Hello, world!", withDefaultMessage.Render("email"))
	assert.Equal(t, "Hello, world!", withDefaultMessage.Render("slack"))

	builderWithMessageForTransport := builder.WithMessageForTransport("email", "Hello, email!")

	withMessageForTransport := builderWithMessageForTransport.Build()
	assert.Equal(t, event.ID("a1c11a53-c4be-488f-89b6-f83bf2d48dab"), empty.Context().ID)
	assert.Equal(t, event.Type("com.example.test"), withMessageForTransport.Context().Type)
	assert.Len(t, withRecipient.Recipient().ToList(), 1)
	assert.Equal(t, "Hello, email!", withMessageForTransport.Render("email"))
	assert.Equal(t, "Hello, world!", withMessageForTransport.Render("slack"))

	builderWithSlackOpts := builder.WithSlackOptions(
		slack.MsgOptionAttachments(
			slack.Attachment{
				Title: "Hello",
				Text:  "world!",
			},
		),
	)

	withSlackOpts := builderWithSlackOpts.Build()
	assert.Equal(t, event.ID("a1c11a53-c4be-488f-89b6-f83bf2d48dab"), empty.Context().ID)
	assert.Equal(t, event.Type("com.example.test"), withSlackOpts.Context().Type)
	assert.Len(t, withRecipient.Recipient().ToList(), 1)
	assert.Equal(t, "Hello, email!", withSlackOpts.Render("email"))
	assert.Equal(t, "Hello, world!", withSlackOpts.Render("slack"))
	assert.Len(t, withSlackOpts.GetSlackOptions(), 1)
}
