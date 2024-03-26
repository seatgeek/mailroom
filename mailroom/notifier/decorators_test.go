// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/notification"
	"github.com/seatgeek/mailroom/mailroom/notifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWithTimeout(t *testing.T) {
	t.Parallel()

	fakeNotification := notification.NewBuilder("test").Build()

	timeout := 30 * time.Second
	expectedDeadline := time.Now().Add(timeout)

	transport := notifier.NewMockTransport(t)
	transport.EXPECT().ID().Return("test")
	transport.EXPECT().Push(mock.AnythingOfType("*context.timerCtx"), mock.Anything).Run(
		func(ctx context.Context, notification common.Notification) {
			deadline, ok := ctx.Deadline()
			assert.True(t, ok, "Expected context to have a deadline")
			assert.WithinDuration(t, expectedDeadline, deadline, time.Second, "Deadline should be within a second of expected")
			assert.Same(t, fakeNotification, notification, "Notification should be the same")
		}).Return(nil)

	wrapped := notifier.WithTimeout(transport, timeout)

	assert.Equal(t, transport.ID(), wrapped.ID(), "ID should be the same")

	_ = wrapped.Push(context.Background(), fakeNotification)
}

func TestWithRetry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		maxRetries   uint64
		givenErrs    []error
		wantAttempts int
		wantErr      error
	}{
		{
			name:       "no errors",
			maxRetries: 2,
			givenErrs: []error{
				nil,
			},
			wantAttempts: 1,
			wantErr:      nil,
		},
		{
			name:       "one error",
			maxRetries: 2,
			givenErrs: []error{
				errors.New("test"),
				nil,
			},
			wantAttempts: 2,
			wantErr:      nil,
		},
		{
			name:       "one permanent error",
			maxRetries: 2,
			givenErrs: []error{
				notifier.Permanent(errors.New("test")),
			},
			wantAttempts: 1,
			wantErr:      errors.New("test"),
		},
		{
			name:       "max attempts",
			maxRetries: 2,
			givenErrs: []error{
				errors.New("err 1"),
				errors.New("err 2"),
				errors.New("err 3"),
			},
			wantAttempts: 3,
			wantErr:      errors.New("err 3"),
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			transport := notifier.NewMockTransport(t)
			transport.EXPECT().ID().Return("test")
			for _, givenErr := range tc.givenErrs {
				transport.EXPECT().Push(mock.Anything, mock.Anything).Return(givenErr).Once()
			}

			wrapped := notifier.WithRetry(transport, tc.maxRetries, func(b *backoff.ExponentialBackOff) {
				b.InitialInterval = 1 * time.Millisecond
				b.MaxInterval = 10 * time.Millisecond
				b.MaxElapsedTime = 20 * time.Millisecond
			})

			assert.Equal(t, transport.ID(), wrapped.ID(), "ID should be the same")

			err := wrapped.Push(context.Background(), notification.NewBuilder("test").Build())

			assert.Equal(t, tc.wantErr, err, "Error should match")
		})
	}
}
