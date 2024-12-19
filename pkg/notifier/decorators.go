// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier

import (
	"context"
	"log/slog"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/seatgeek/mailroom/pkg/common"
)

// WithTimeout decorates the given Transport with a timeout
func WithTimeout(transport Transport, timeout time.Duration) Transport {
	return &withTimeout{
		Transport: transport,
		timeout:   timeout,
	}
}

type withTimeout struct {
	Transport
	timeout time.Duration
}

func (w *withTimeout) Push(ctx context.Context, notification common.Notification) error {
	ctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	return w.Transport.Push(ctx, notification)
}

func (w *withTimeout) Validate(ctx context.Context) error {
	if v, ok := w.Transport.(common.Validator); ok {
		return v.Validate(ctx)
	}

	return nil
}

type BackOff = backoff.BackOff

// WithRetry decorates the given Transport with retry logic using the provided backoff
func WithRetry(transport Transport, maxRetries uint64, backoffProvider func() BackOff) Transport {
	return &withRetry{
		Transport:  transport,
		maxRetries: maxRetries,
		backoff:    backoffProvider,
	}
}

type withRetry struct {
	Transport
	maxRetries uint64
	backoff    func() BackOff
}

func (w *withRetry) Push(ctx context.Context, notification common.Notification) error {
	_, err := backoff.Retry(
		ctx,
		func() (bool, error) {
			return true, w.Transport.Push(ctx, notification)
		},
		backoff.WithMaxTries(uint(w.maxRetries+1)),
		backoff.WithBackOff(w.backoff()),
		backoff.WithNotify(func(err error, duration time.Duration) {
			slog.Error("failed to push notification", "id", notification.Context().ID, "error", err, "next_retry", duration.String())
		}),
	)

	return err
}

func (w *withRetry) Validate(ctx context.Context) error {
	if v, ok := w.Transport.(common.Validator); ok {
		return v.Validate(ctx)
	}

	return nil
}

// WithLogging decorates the given Transport and logs every successful push (including the message body)
func WithLogging(transport Transport, logger *slog.Logger, logLevel slog.Level) Transport {
	return &withLogging{
		Transport: transport,
		logger:    logger,
		level:     logLevel,
	}
}

type withLogging struct {
	Transport
	logger *slog.Logger
	level  slog.Level
}

func (w *withLogging) Push(ctx context.Context, n common.Notification) error {
	err := w.Transport.Push(ctx, n)
	if err == nil {
		w.logger.Log(ctx, w.level, "sent notification", "id", n.Context().ID, "type", n.Context().Type, "to", n.Recipient(), "message", n.Render("logger"))
	}

	return err
}

func (w *withLogging) Validate(ctx context.Context) error {
	if v, ok := w.Transport.(common.Validator); ok {
		return v.Validate(ctx)
	}

	return nil
}
