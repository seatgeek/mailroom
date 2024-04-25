// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notification_test

import (
	"testing"

	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/identifier"
	"github.com/seatgeek/mailroom/mailroom/notification"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	builder := notification.NewBuilder("a1c11a53-c4be-488f-89b6-f83bf2d48dab", "com.example.test")

	empty := builder.Build()
	assert.Equal(t, common.EventID("a1c11a53-c4be-488f-89b6-f83bf2d48dab"), empty.ID())
	assert.Equal(t, common.EventType("com.example.test"), empty.Type())
	assert.Empty(t, empty.Recipient().ToList())
	assert.Empty(t, empty.Render("email"))

	builderWithRecipient := builder.WithRecipientIdentifiers(identifier.New(identifier.GenericUsername, "codell"))

	withRecipient := builderWithRecipient.Build()
	assert.Equal(t, common.EventID("a1c11a53-c4be-488f-89b6-f83bf2d48dab"), empty.ID())
	assert.Equal(t, common.EventType("com.example.test"), withRecipient.Type())
	assert.Len(t, withRecipient.Recipient().ToList(), 1)
	assert.Equal(t, "codell", withRecipient.Recipient().MustGet(identifier.GenericUsername))
	assert.Empty(t, withRecipient.Render("email"))

	builderWithDefaultMessage := builder.WithDefaultMessage("Hello, world!")

	withDefaultMessage := builderWithDefaultMessage.Build()
	assert.Equal(t, common.EventID("a1c11a53-c4be-488f-89b6-f83bf2d48dab"), empty.ID())
	assert.Equal(t, common.EventType("com.example.test"), withDefaultMessage.Type())
	assert.Len(t, withRecipient.Recipient().ToList(), 1)
	assert.Equal(t, "Hello, world!", withDefaultMessage.Render("email"))
	assert.Equal(t, "Hello, world!", withDefaultMessage.Render("slack"))

	builderWithMessageForTransport := builder.WithMessageForTransport("email", "Hello, email!")

	withMessageForTransport := builderWithMessageForTransport.Build()
	assert.Equal(t, common.EventID("a1c11a53-c4be-488f-89b6-f83bf2d48dab"), empty.ID())
	assert.Equal(t, common.EventType("com.example.test"), withMessageForTransport.Type())
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
	assert.Equal(t, common.EventID("a1c11a53-c4be-488f-89b6-f83bf2d48dab"), empty.ID())
	assert.Equal(t, common.EventType("com.example.test"), withSlackOpts.Type())
	assert.Len(t, withRecipient.Recipient().ToList(), 1)
	assert.Equal(t, "Hello, email!", withSlackOpts.Render("email"))
	assert.Equal(t, "Hello, world!", withSlackOpts.Render("slack"))
	assert.Len(t, withSlackOpts.GetSlackOptions(), 1)
}
