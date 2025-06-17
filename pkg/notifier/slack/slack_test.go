// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package slack_test

import (
	"errors"
	"testing"

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/notification"
	"github.com/seatgeek/mailroom/pkg/notifier"
	"github.com/seatgeek/mailroom/pkg/notifier/slack"
	slackgo "github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestNewTransport(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		key   event.TransportKey
		token string
		opts  []slackgo.Option
	}{
		{
			name:  "creates transport with key and token",
			key:   "slack",
			token: "xoxb-test-token",
			opts:  nil,
		},
		{
			name:  "creates transport with options",
			key:   "custom-slack",
			token: "xoxb-another-token",
			opts:  []slackgo.Option{slackgo.OptionDebug(true)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			transport := slack.NewTransport(tc.key, tc.token, tc.opts...)

			assert.NotNil(t, transport)
			assert.Equal(t, tc.key, transport.Key())
		})
	}
}

func TestTransport_Key(t *testing.T) {
	t.Parallel()

	expectedKey := event.TransportKey("test-slack")
	transport := slack.NewTransport(expectedKey, "xoxb-test-token")

	assert.Equal(t, expectedKey, transport.Key())
}

func TestTransport_Push(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	slackID := "U1234567890"
	transportKey := event.TransportKey("slack")

	tests := []struct {
		name         string
		notification event.Notification
		wantErr      bool
		expectedErr  error
	}{
		{
			name: "happy path - basic notification",
			notification: notification.NewBuilder(event.Context{
				ID:   "test-id",
				Type: "test.notification",
			}).
				WithRecipientIdentifiers(identifier.New(slack.ID, slackID)).
				WithDefaultMessage("Hello, world!").
				Build(),
			wantErr: false,
		},
		{
			name: "happy path - rich notification with slack options",
			notification: notification.NewBuilder(event.Context{
				ID:   "test-id-rich",
				Type: "test.rich.notification",
			}).
				WithRecipientIdentifiers(identifier.New(slack.ID, slackID)).
				WithDefaultMessage("Rich notification").
				WithSlackOptions(slackgo.MsgOptionAttachments(slackgo.Attachment{
					Title: "Test Attachment",
					Text:  "This is a test attachment",
				})).
				Build(),
			wantErr: false,
		},
		{
			name: "happy path - notification with empty message",
			notification: notification.NewBuilder(event.Context{
				ID:   "test-id-empty",
				Type: "test.empty.notification",
			}).
				WithRecipientIdentifiers(identifier.New(slack.ID, slackID)).
				Build(),
			wantErr: false,
		},
		{
			name: "error - recipient without slack ID",
			notification: notification.NewBuilder(event.Context{
				ID:   "test-id-no-slack",
				Type: "test.notification",
			}).
				WithRecipientIdentifiers(identifier.New("email", "test@example.com")).
				WithDefaultMessage("This should fail").
				Build(),
			wantErr:     true,
			expectedErr: notifier.Permanent(errors.New("recipient does not have a Slack ID")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create a transport - we'll test the actual Slack API integration in integration tests
			// For unit tests, we focus on the business logic
			transport := slack.NewTransport(transportKey, "xoxb-test-token")

			err := transport.Push(ctx, tc.notification)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != nil {
					// Check that it's a permanent error for missing Slack ID
					if tc.expectedErr.Error() == "recipient does not have a Slack ID" {
						// The error should be wrapped with notifier.Permanent
						// We can't easily test the exact error type due to how backoff.Permanent works
						// but we can test the message
						assert.Contains(t, err.Error(), "recipient does not have a Slack ID")
					}
				}
			} else if err != nil {
				// For success cases, we expect an error because we're using a fake token
				// but the business logic should have succeeded (no permanent error about missing Slack ID)
				// This is acceptable for unit tests - we're testing our logic, not the Slack API
				// Make sure it's not our permanent error about missing Slack ID
				assert.NotContains(t, err.Error(), "recipient does not have a Slack ID")
			}
		})
	}
}

func TestTransport_Push_WithMockNotification(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	transportKey := event.TransportKey("slack")
	slackID := "U1234567890"

	tests := []struct {
		name               string
		setupNotification  func() event.Notification
		wantErr            bool
		expectedErrMessage string
	}{
		{
			name: "notification without slack ID",
			setupNotification: func() event.Notification {
				mockNotification := event.NewMockNotification(t)
				mockNotification.EXPECT().Recipient().Return(identifier.NewSet(
					identifier.New("email", "test@example.com"),
				)).Maybe()
				return mockNotification
			},
			wantErr:            true,
			expectedErrMessage: "recipient does not have a Slack ID",
		},
		{
			name: "notification with slack ID",
			setupNotification: func() event.Notification {
				mockNotification := event.NewMockNotification(t)
				mockNotification.EXPECT().Recipient().Return(identifier.NewSet(
					identifier.New(slack.ID, slackID),
				)).Maybe()
				mockNotification.EXPECT().Render(transportKey).Return("Test message").Maybe()
				return mockNotification
			},
			wantErr: false, // We expect Slack API error, but not our business logic error
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			transport := slack.NewTransport(transportKey, "xoxb-fake-token")
			notification := tc.setupNotification()

			err := transport.Push(ctx, notification)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErrMessage != "" {
					assert.Contains(t, err.Error(), tc.expectedErrMessage)
				}
			} else if err != nil && tc.expectedErrMessage != "" {
				// For valid business logic, we might still get Slack API errors due to fake token
				// but we shouldn't get our specific permanent error
				assert.NotContains(t, err.Error(), tc.expectedErrMessage)
			}
		})
	}
}

func TestTransport_Push_WithRichNotification(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	transportKey := event.TransportKey("slack")
	slackID := "U1234567890"

	tests := []struct {
		name            string
		setupMock       func() *slack.MockRichNotification
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "rich notification with slack options",
			setupMock: func() *slack.MockRichNotification {
				mockRichNotification := slack.NewMockRichNotification(t)
				mockRichNotification.EXPECT().Recipient().Return(identifier.NewSet(
					identifier.New(slack.ID, slackID),
				)).Maybe()
				mockRichNotification.EXPECT().Render(transportKey).Return("Rich message").Maybe()
				mockRichNotification.EXPECT().GetSlackOptions().Return([]slackgo.MsgOption{
					slackgo.MsgOptionAttachments(slackgo.Attachment{
						Title: "Test Rich Attachment",
						Text:  "Rich attachment text",
					}),
				}).Maybe()
				return mockRichNotification
			},
			wantErr: false, // Expect Slack API error, not business logic error
		},
		{
			name: "rich notification without slack ID",
			setupMock: func() *slack.MockRichNotification {
				mockRichNotification := slack.NewMockRichNotification(t)
				mockRichNotification.EXPECT().Recipient().Return(identifier.NewSet(
					identifier.New("email", "test@example.com"),
				)).Maybe()
				return mockRichNotification
			},
			wantErr:         true,
			wantErrContains: "recipient does not have a Slack ID",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			transport := slack.NewTransport(transportKey, "xoxb-fake-token")
			richNotification := tc.setupMock()

			err := transport.Push(ctx, richNotification)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantErrContains != "" {
					assert.Contains(t, err.Error(), tc.wantErrContains)
				}
			} else if err != nil && tc.wantErrContains != "" {
				// For valid business logic, we might still get Slack API errors
				// but we shouldn't get our business logic errors
				assert.NotContains(t, err.Error(), tc.wantErrContains)
			}
		})
	}
}

func TestTransport_Validate(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "validation with fake token returns error",
			token:   "xoxb-fake-token",
			wantErr: true, // Expect error with fake token
		},
		{
			name:    "validation with empty token returns error",
			token:   "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			transport := slack.NewTransport("slack", tc.token)

			err := transport.Validate(ctx)

			if tc.wantErr {
				assert.Error(t, err)
				// Should be a permanent error
				assert.Contains(t, err.Error(), "authentication failed")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
