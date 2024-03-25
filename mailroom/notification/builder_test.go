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

	builder := notification.NewBuilder("com.example.test")

	empty := builder.Build()
	assert.Equal(t, common.EventType("com.example.test"), empty.Type())
	assert.Empty(t, empty.Recipients().ToList())
	assert.Empty(t, empty.Render("email"))

	builderWithRecipient := builder.WithRecipient(identifier.New(identifier.GenericUsername, "codell"))

	withRecipient := builderWithRecipient.Build()
	assert.Equal(t, common.EventType("com.example.test"), withRecipient.Type())
	assert.Len(t, withRecipient.Recipients().ToList(), 1)
	assert.Equal(t, "codell", withRecipient.Recipients().MustGet(identifier.GenericUsername))
	assert.Empty(t, withRecipient.Render("email"))

	builderWithRecipients := builder.WithRecipients(identifier.NewCollection(
		identifier.New(identifier.GenericUsername, "rufus"),
		identifier.New(identifier.GenericEmail, "rufus@seatgeek.com"),
	))

	withRecipients := builderWithRecipients.Build()
	assert.Equal(t, common.EventType("com.example.test"), withRecipients.Type())
	assert.Len(t, withRecipient.Recipients().ToList(), 2)
	assert.Equal(t, "rufus", withRecipients.Recipients().MustGet(identifier.GenericUsername))
	assert.Equal(t, "rufus@seatgeek.com", withRecipients.Recipients().MustGet(identifier.GenericEmail))
	assert.Empty(t, withRecipients.Render("email"))

	builderWithDefaultMessage := builder.WithDefaultMessage("Hello, world!")

	withDefaultMessage := builderWithDefaultMessage.Build()
	assert.Equal(t, common.EventType("com.example.test"), withDefaultMessage.Type())
	assert.Len(t, withRecipient.Recipients().ToList(), 2)
	assert.Equal(t, "Hello, world!", withDefaultMessage.Render("email"))
	assert.Equal(t, "Hello, world!", withDefaultMessage.Render("slack"))

	builderWithMessageForTransport := builder.WithMessageForTransport("email", "Hello, email!")

	withMessageForTransport := builderWithMessageForTransport.Build()
	assert.Equal(t, common.EventType("com.example.test"), withMessageForTransport.Type())
	assert.Len(t, withRecipient.Recipients().ToList(), 2)
	assert.Equal(t, "Hello, email!", withMessageForTransport.Render("email"))
	assert.Equal(t, "Hello, world!", withMessageForTransport.Render("slack"))

	builderWithSlackOpts := builder.WithSlackMessage(
		slack.MsgOptionText("I'm a Slack option!", false),
	)

	withSlackOpts := builderWithSlackOpts.Build()
	assert.Equal(t, common.EventType("com.example.test"), withSlackOpts.Type())
	assert.Len(t, withRecipient.Recipients().ToList(), 2)
	assert.Equal(t, "Hello, email!", withSlackOpts.Render("email"))
	assert.Equal(t, "Hello, world!", withSlackOpts.Render("slack"))
	assert.Len(t, withSlackOpts.RenderSlack(), 1)
}
