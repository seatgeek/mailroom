// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notification_test

import (
	"testing"

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/notification"
	slack2 "github.com/seatgeek/mailroom/pkg/notifier/slack"
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

func TestNotificationWithRecipient(t *testing.T) {
	t.Parallel()

	originalNotification := notification.NewBuilder(event.Context{
		ID:   "test-id",
		Type: "test-type",
		Labels: map[string]string{
			"env": "test",
		},
	}).
		WithRecipientIdentifiers(identifier.New(identifier.GenericUsername, "original-user")).
		WithDefaultMessage("Hello message").
		WithMessageForTransport("email", "Email message").
		WithSlackOptions(slack.MsgOptionText("slack text", false)).
		Build()

	// WithRecipient mutates the original
	newRecipient := identifier.NewSet(identifier.New(identifier.GenericUsername, "new-user"))
	modifiedNotification := originalNotification.WithRecipient(newRecipient)

	assert.Same(t, originalNotification, modifiedNotification)

	assert.Equal(t, "new-user", originalNotification.Recipient().MustGet(identifier.GenericUsername))
	assert.Equal(t, "new-user", modifiedNotification.Recipient().MustGet(identifier.GenericUsername))

	assert.Equal(t, "Hello message", originalNotification.Render("default"))
	assert.Equal(t, "Email message", originalNotification.Render("email"))
	assert.Len(t, originalNotification.GetSlackOptions(), 1)
}

func TestNotificationCopy(t *testing.T) {
	t.Parallel()

	originalNotification := notification.NewBuilder(event.Context{
		ID:   "test-id",
		Type: "test-type",
		Labels: map[string]string{
			"environment": "test",
			"version":     "1.0.0",
		},
	}).
		WithRecipientIdentifiers(identifier.New(identifier.GenericUsername, "original-user")).
		WithDefaultMessage("Default message").
		WithMessageForTransport("email", "Email message").
		WithMessageForTransport("slack", "Slack message").
		WithSlackOptions(
			slack.MsgOptionText("slack text", false),
			slack.MsgOptionAttachments(slack.Attachment{Title: "Test", Text: "Attachment"}),
		).
		Build()

	clonedNotification := originalNotification.Copy()

	assert.NotSame(t, originalNotification, clonedNotification)

	assert.Equal(t, originalNotification.Context().ID, clonedNotification.Context().ID)
	assert.Equal(t, originalNotification.Context().Type, clonedNotification.Context().Type)
	assert.Equal(t, originalNotification.Context().Labels, clonedNotification.Context().Labels)
	assert.Equal(t, originalNotification.Recipient().MustGet(identifier.GenericUsername),
		clonedNotification.Recipient().MustGet(identifier.GenericUsername))
	assert.Equal(t, originalNotification.Render("email"), clonedNotification.Render("email"))
	assert.Equal(t, originalNotification.Render("slack"), clonedNotification.Render("slack"))
	assert.Equal(t, originalNotification.Render("default"), clonedNotification.Render("default"))

	richCloned, ok := clonedNotification.(slack2.RichNotification)
	assert.True(t, ok, "cloned notification should implement RichNotification")
	assert.Len(t, richCloned.GetSlackOptions(), 2)

	newRecipient := identifier.NewSet(identifier.New(identifier.GenericUsername, "modified-user"))
	originalNotification.WithRecipient(newRecipient)

	assert.Equal(t, "modified-user", originalNotification.Recipient().MustGet(identifier.GenericUsername))
	assert.Equal(t, "original-user", clonedNotification.Recipient().MustGet(identifier.GenericUsername))

	originalLabels := originalNotification.Context().Labels
	clonedLabels := clonedNotification.Context().Labels

	assert.Equal(t, originalLabels, clonedLabels)

	originalLabels["new-key"] = "new-value"
	assert.NotEqual(t, originalLabels, clonedLabels)
	assert.NotContains(t, clonedLabels, "new-key")
}

func TestNotificationCopyComplexScenarios(t *testing.T) {
	t.Parallel()

	// Test cloning with nil/empty values
	emptyNotification := notification.NewBuilder(event.Context{
		ID:   "empty-test",
		Type: "empty-type",
	}).Build()

	clonedEmpty := emptyNotification.Copy()
	assert.NotSame(t, emptyNotification, clonedEmpty)
	assert.Empty(t, clonedEmpty.Recipient().ToList())
	assert.Empty(t, clonedEmpty.Render("any-transport"))

	richClonedEmpty, ok := clonedEmpty.(slack2.RichNotification)
	assert.True(t, ok, "cloned empty notification should implement RichNotification")
	assert.Empty(t, richClonedEmpty.GetSlackOptions())

	contextWithNilLabels := event.Context{
		ID:     "nil-labels-test",
		Type:   "nil-type",
		Labels: nil,
	}

	notificationWithNilLabels := notification.NewBuilder(contextWithNilLabels).Build()
	clonedNilLabels := notificationWithNilLabels.Copy()

	assert.Nil(t, clonedNilLabels.Context().Labels)
	assert.Equal(t, notificationWithNilLabels.Context().Labels, clonedNilLabels.Context().Labels)
}
