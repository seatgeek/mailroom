// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/notification"
	"github.com/seatgeek/mailroom/pkg/notifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWithTimeout(t *testing.T) {
	t.Parallel()

	fakeNotification := event.NewMockNotification(t)

	timeout := 30 * time.Second
	expectedDeadline := time.Now().Add(timeout)

	transport := notifier.NewMockTransport(t)
	transport.EXPECT().Key().Return("test")
	transport.EXPECT().Push(mock.AnythingOfType("*context.timerCtx"), mock.Anything).Run(
		func(ctx context.Context, notification event.Notification) {
			deadline, ok := ctx.Deadline()
			assert.True(t, ok, "Expected context to have a deadline")
			assert.WithinDuration(t, expectedDeadline, deadline, time.Second, "Deadline should be within a second of expected")
			assert.Same(t, fakeNotification, notification, "Notification should be the same")
		}).Return(nil)

	wrapped := notifier.WithTimeout(transport, timeout)

	assert.Equal(t, transport.Key(), wrapped.Key(), "Key should be the same")

	_ = wrapped.Push(t.Context(), fakeNotification)
}

func TestWithRetry(t *testing.T) {
	t.Parallel()

	fakeNotification := notification.NewBuilder(event.Context{
		ID:   "a1c11a53-c4be-488f-89b6-f83bf2d48dab",
		Type: "test",
	}).Build()

	err1 := errors.New("err 1")
	err2 := errors.New("err 2")

	tests := []struct {
		name         string
		maxTries     uint
		givenErrs    []error
		wantAttempts int
		wantErr      error
	}{
		{
			name:     "no errors",
			maxTries: 2,
			givenErrs: []error{
				nil,
			},
			wantAttempts: 1,
			wantErr:      nil,
		},
		{
			name:     "one error",
			maxTries: 2,
			givenErrs: []error{
				err1,
				nil,
			},
			wantAttempts: 2,
			wantErr:      nil,
		},
		{
			name:     "one permanent error",
			maxTries: 2,
			givenErrs: []error{
				notifier.Permanent(err1),
			},
			wantAttempts: 1,
			wantErr:      err1,
		},
		{
			name:     "max attempts",
			maxTries: 2,
			givenErrs: []error{
				err1,
				err2,
			},
			wantAttempts: 2,
			wantErr:      err2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			transport := notifier.NewMockTransport(t)
			transport.EXPECT().Key().Return("test")
			for _, givenErr := range tc.givenErrs {
				transport.EXPECT().Push(mock.Anything, mock.Anything).Return(givenErr).Once()
			}

			wrapped := notifier.WithRetry(transport, tc.maxTries, func() backoff.BackOff {
				return backoff.NewConstantBackOff(10 * time.Millisecond)
			})

			assert.Equal(t, transport.Key(), wrapped.Key(), "Key should be the same")

			err := wrapped.Push(t.Context(), fakeNotification)

			assert.ErrorIs(t, err, tc.wantErr, "Error should match")
		})
	}
}

func TestWithLogging(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level slog.Level
	}{
		{
			name:  "logs at info level",
			level: slog.LevelInfo,
		},
		{
			name:  "logs at debug level",
			level: slog.LevelDebug,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			notification := notification.NewBuilder(event.Context{
				ID:   "a1c11a53-c4be-488f-89b6-f83bf2d48dab",
				Type: "test",
			}).
				WithRecipientIdentifiers(
					identifier.New("username", "rufus"),
					identifier.New("email", "rufus@seatgeek.com"),
				).
				WithDefaultMessage("hello world").
				Build()

			transport := notifier.NewMockTransport(t)
			transport.EXPECT().Key().Return("test")
			transport.EXPECT().Push(mock.Anything, mock.Anything).Return(nil)

			buffer := new(bytes.Buffer)

			logger := slog.New(
				slog.NewJSONHandler(buffer, &slog.HandlerOptions{
					Level: tc.level,
				}),
			)

			wrapped := notifier.WithLogging(transport, logger, tc.level)

			assert.Equal(t, transport.Key(), wrapped.Key(), "Key should be the same")

			_ = wrapped.Push(t.Context(), notification)

			var logEntry struct {
				Level   string            `json:"level"`
				Msg     string            `json:"msg"`
				ID      string            `json:"id"`
				Type    string            `json:"type"`
				To      map[string]string `json:"to"`
				Message string            `json:"message"`
			}
			if err := json.Unmarshal(buffer.Bytes(), &logEntry); err != nil {
				t.Fatalf("failed to unmarshal log entry: %s", err)
			}

			assert.Equal(t, tc.level.String(), logEntry.Level)
			assert.Equal(t, "sent notification", logEntry.Msg)
			assert.Equal(t, "a1c11a53-c4be-488f-89b6-f83bf2d48dab", logEntry.ID)
			assert.Equal(t, "test", logEntry.Type)
			assert.Equal(t, "rufus@seatgeek.com", logEntry.To["email"])
			assert.Equal(t, "rufus", logEntry.To["username"])
			assert.Equal(t, "hello world", logEntry.Message)
		})
	}
}
