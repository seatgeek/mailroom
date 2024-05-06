// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier

import (
	"context"
	"fmt"
	"io"

	"github.com/seatgeek/mailroom/mailroom/common"
)

// WriterNotifier is a notifier that simply writes notifications somewhere, like a file or stdout
// It is primarily used for testing and debugging
type WriterNotifier struct {
	key    common.TransportKey
	writer io.Writer
}

var _ Transport = &WriterNotifier{}

func (c *WriterNotifier) Key() common.TransportKey {
	return c.key
}

func (c *WriterNotifier) Push(_ context.Context, n common.Notification) error {
	_, err := fmt.Fprintf(
		c.writer,
		"notification: id=%s type=%s, to=%s, message=%s\n",
		n.Context().ID,
		n.Context().Type,
		n.Recipient(),
		n.Render("writer"),
	)

	return err
}

// NewWriterNotifier creates a Notifier that writes notifications to places like files or stdout
func NewWriterNotifier(key common.TransportKey, writer io.Writer) *WriterNotifier {
	return &WriterNotifier{key: key, writer: writer}
}
