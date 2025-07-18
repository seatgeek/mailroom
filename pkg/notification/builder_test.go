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
	assert.Equal(t, event.ID("a1c11a53-c4be-488f-89b6-f83bf2d48dab"), withRecipient.Context().ID)
	assert.Equal(t, event.Type("com.example.test"), withRecipient.Context().Type)
	assert.Len(t, withRecipient.Recipient().ToList(), 1)
	assert.Equal(t, "codell", withRecipient.Recipient().MustGet(identifier.GenericUsername))
	assert.Empty(t, withRecipient.Render("email"))

	builderWithDefaultMessage := builder.WithDefaultMessage("Hello, world!")

	withDefaultMessage := builderWithDefaultMessage.Build()
	assert.Equal(t, event.ID("a1c11a53-c4be-488f-89b6-f83bf2d48dab"), withDefaultMessage.Context().ID)
	assert.Equal(t, event.Type("com.example.test"), withDefaultMessage.Context().Type)
	assert.Empty(t, withDefaultMessage.Recipient().ToList()) // This should be empty since we didn't add recipient to this builder
	assert.Equal(t, "Hello, world!", withDefaultMessage.Render("email"))
	assert.Equal(t, "Hello, world!", withDefaultMessage.Render("slack"))

	builderWithMessageForTransport := builderWithDefaultMessage.WithMessageForTransport("email", "Hello, email!")

	withMessageForTransport := builderWithMessageForTransport.Build()
	assert.Equal(t, event.ID("a1c11a53-c4be-488f-89b6-f83bf2d48dab"), withMessageForTransport.Context().ID)
	assert.Equal(t, event.Type("com.example.test"), withMessageForTransport.Context().Type)
	assert.Empty(t, withMessageForTransport.Recipient().ToList()) // This should be empty since we didn't add recipient to this builder
	assert.Equal(t, "Hello, email!", withMessageForTransport.Render("email"))
	assert.Equal(t, "Hello, world!", withMessageForTransport.Render("slack"))

	builderWithSlackOpts := builderWithMessageForTransport.WithSlackOptions(
		slack.MsgOptionAttachments(
			slack.Attachment{
				Title: "Hello",
				Text:  "world!",
			},
		),
	)

	withSlackOpts := builderWithSlackOpts.Build()
	assert.Equal(t, event.ID("a1c11a53-c4be-488f-89b6-f83bf2d48dab"), withSlackOpts.Context().ID)
	assert.Equal(t, event.Type("com.example.test"), withSlackOpts.Context().Type)
	assert.Empty(t, withSlackOpts.Recipient().ToList()) // This should be empty since we didn't add recipient to this builder
	assert.Equal(t, "Hello, email!", withSlackOpts.Render("email"))
	assert.Equal(t, "Hello, world!", withSlackOpts.Render("slack"))
	assert.Len(t, withSlackOpts.GetSlackOptions(), 1)
}

func TestBuilderImmutability(t *testing.T) {
	t.Parallel()

	originalBuilder := notification.NewBuilder(event.Context{
		ID:   "test-id",
		Type: "test-type",
	})

	// Test that WithRecipientIdentifiers creates a new builder instance
	builderWithRecipient := originalBuilder.WithRecipientIdentifiers(identifier.New(identifier.GenericUsername, "user1"))

	// Original builder should still be empty
	originalNotification := originalBuilder.Build()
	assert.Empty(t, originalNotification.Recipient().ToList())

	// New builder should have the recipient
	newNotification := builderWithRecipient.Build()
	assert.Len(t, newNotification.Recipient().ToList(), 1)
	assert.Equal(t, "user1", newNotification.Recipient().MustGet(identifier.GenericUsername))

	// Test that WithDefaultMessage creates a new builder instance
	builderWithMessage := originalBuilder.WithDefaultMessage("Hello")

	// Original builder should still have empty message
	originalNotification = originalBuilder.Build()
	assert.Empty(t, originalNotification.Render("email"))

	// New builder should have the message
	newNotification = builderWithMessage.Build()
	assert.Equal(t, "Hello", newNotification.Render("email"))

	// Test that WithMessageForTransport creates a new builder instance
	builderWithTransportMessage := originalBuilder.WithMessageForTransport("email", "Email message")

	// Original builder should still have empty message
	originalNotification = originalBuilder.Build()
	assert.Empty(t, originalNotification.Render("email"))

	// New builder should have the transport-specific message
	newNotification = builderWithTransportMessage.Build()
	assert.Equal(t, "Email message", newNotification.Render("email"))
}

func TestNotificationWithRecipient(t *testing.T) {
	t.Parallel()

	// Create an original notification
	originalNotification := notification.NewBuilder(event.Context{
		ID:   "test-id",
		Type: "test-type",
	}).
		WithRecipientIdentifiers(identifier.New(identifier.GenericUsername, "original-user")).
		WithDefaultMessage("Hello message").
		Build()

	// Test the new WithRecipient method on the notification interface
	newRecipient := identifier.NewSet(identifier.New(identifier.GenericUsername, "new-user"))
	modifiedNotification := originalNotification.WithRecipient(newRecipient)

	// Verify the original notification is unchanged
	assert.Equal(t, "original-user", originalNotification.Recipient().MustGet(identifier.GenericUsername))
	assert.Equal(t, "Hello message", originalNotification.Render("email"))

	// Verify the modified notification has the new recipient
	assert.Equal(t, "new-user", modifiedNotification.Recipient().MustGet(identifier.GenericUsername))
	assert.Equal(t, "Hello message", modifiedNotification.Render("email"))
	assert.Equal(t, originalNotification.Context().ID, modifiedNotification.Context().ID)
	assert.Equal(t, originalNotification.Context().Type, modifiedNotification.Context().Type)

	// Test that both notifications maintain their independence
	anotherRecipient := identifier.NewSet(identifier.New(identifier.GenericUsername, "another-user"))
	anotherNotification := modifiedNotification.WithRecipient(anotherRecipient)

	assert.Equal(t, "original-user", originalNotification.Recipient().MustGet(identifier.GenericUsername))
	assert.Equal(t, "new-user", modifiedNotification.Recipient().MustGet(identifier.GenericUsername))
	assert.Equal(t, "another-user", anotherNotification.Recipient().MustGet(identifier.GenericUsername))
}

func TestNotificationWithRecipientImmutability(t *testing.T) {
	t.Parallel()

	// Create a notification with maps and slices that could be shared
	originalNotification := notification.NewBuilder(event.Context{
		ID:   "test-id",
		Type: "test-type",
	}).
		WithRecipientIdentifiers(identifier.New(identifier.GenericUsername, "original-user")).
		WithMessageForTransport("email", "Email message").
		WithMessageForTransport("slack", "Slack message").
		WithSlackOptions(slack.MsgOptionText("original text", false)).
		Build()

	// Create a new notification with WithRecipient
	newRecipient := identifier.NewSet(identifier.New(identifier.GenericUsername, "new-user"))
	modifiedNotification := originalNotification.WithRecipient(newRecipient)

	// Verify they have different recipients
	assert.Equal(t, "original-user", originalNotification.Recipient().MustGet(identifier.GenericUsername))
	assert.Equal(t, "new-user", modifiedNotification.Recipient().MustGet(identifier.GenericUsername))

	// Verify they both have the expected messages initially
	assert.Equal(t, "Email message", originalNotification.Render("email"))
	assert.Equal(t, "Slack message", originalNotification.Render("slack"))
	assert.Equal(t, "Email message", modifiedNotification.Render("email"))
	assert.Equal(t, "Slack message", modifiedNotification.Render("slack"))

	// Verify they both have the expected slack options initially
	// Cast to RichNotification to access GetSlackOptions method
	originalRich := originalNotification
	modifiedRich := modifiedNotification.(slack2.RichNotification)
	assert.Len(t, originalRich.GetSlackOptions(), 1)
	assert.Len(t, modifiedRich.GetSlackOptions(), 1)

	// Test that the messagePerTransport maps are independent
	// (We can't directly modify them from the outside since they're private,
	// but we can verify they're independent by checking they don't share memory)

	// Create a new notification from the modified one with different transport message
	furtherModified := notification.NewBuilder(event.Context{
		ID:   "test-id",
		Type: "test-type",
	}).
		WithRecipientIdentifiers(identifier.New(identifier.GenericUsername, "further-user")).
		WithMessageForTransport("email", "Different email message").
		WithMessageForTransport("slack", "Slack message"). // Keep slack the same
		WithSlackOptions(slack.MsgOptionText("original text", false)).
		Build()

	// Use WithRecipient to create a new one
	finalNotification := furtherModified.WithRecipient(identifier.NewSet(identifier.New(identifier.GenericUsername, "final-user")))

	// Verify all notifications maintain their independence
	assert.Equal(t, "Email message", originalNotification.Render("email"))
	assert.Equal(t, "Email message", modifiedNotification.Render("email"))
	assert.Equal(t, "Different email message", furtherModified.Render("email"))
	assert.Equal(t, "Different email message", finalNotification.Render("email"))

	// Verify recipients are all different
	assert.Equal(t, "original-user", originalNotification.Recipient().MustGet(identifier.GenericUsername))
	assert.Equal(t, "new-user", modifiedNotification.Recipient().MustGet(identifier.GenericUsername))
	assert.Equal(t, "further-user", furtherModified.Recipient().MustGet(identifier.GenericUsername))
	assert.Equal(t, "final-user", finalNotification.Recipient().MustGet(identifier.GenericUsername))
}
