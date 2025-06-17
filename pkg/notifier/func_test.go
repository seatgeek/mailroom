// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier_test

import (
	"context"
	"errors"
	"testing"

	"github.com/seatgeek/mailroom/pkg/event"
	notifier2 "github.com/seatgeek/mailroom/pkg/notifier"
	"github.com/stretchr/testify/assert"
)

func TestNewNotifier(t *testing.T) {
	t.Parallel()

	someError := errors.New("some error")

	notifier := notifier2.NewNotifier(func(ctx context.Context, notification event.Notification) error {
		return someError
	})

	err := notifier.Push(t.Context(), nil)

	assert.Same(t, someError, err)
}

func TestNewTransport(t *testing.T) {
	t.Parallel()

	someError := errors.New("some error")

	transport := notifier2.NewTransport("key", func(ctx context.Context, notification event.Notification) error {
		return someError
	})

	assert.Equal(t, event.TransportKey("key"), transport.Key())

	err := transport.Push(t.Context(), nil)
	assert.Same(t, someError, err)
}
