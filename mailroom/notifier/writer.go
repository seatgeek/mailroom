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
	id     common.TransportID
	writer io.Writer
}

var _ Transport = &WriterNotifier{}

func (c *WriterNotifier) ID() common.TransportID {
	return c.id
}

func (c *WriterNotifier) Push(_ context.Context, n common.Notification) error {
	_, err := fmt.Fprintf(
		c.writer,
		"notification: type=%s, to=%s, message=%s\n",
		n.Type(),
		n.Recipients(),
		n.Render("writer"),
	)

	return err
}

// NewWriterNotifier creates a Notifier that writes notifications to places like files or stdout
func NewWriterNotifier(id common.TransportID, writer io.Writer) *WriterNotifier {
	return &WriterNotifier{id: id, writer: writer}
}
