// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier_test

import (
	"context"
	"errors"
	"testing"

	"github.com/seatgeek/mailroom/mailroom/common"
	notifier2 "github.com/seatgeek/mailroom/mailroom/notifier"
	"github.com/stretchr/testify/assert"
)

func TestNewNotifier(t *testing.T) {
	t.Parallel()

	someError := errors.New("some error")

	notifier := notifier2.NewNotifier(func(ctx context.Context, notification common.Notification) error {
		return someError
	})

	err := notifier.Push(context.Background(), nil)

	assert.Same(t, someError, err)
}

func TestNewTransport(t *testing.T) {
	t.Parallel()

	someError := errors.New("some error")

	transport := notifier2.NewTransport("key", func(ctx context.Context, notification common.Notification) error {
		return someError
	})

	assert.Equal(t, common.TransportKey("key"), transport.Key())

	err := transport.Push(context.Background(), nil)
	assert.Same(t, someError, err)
}
