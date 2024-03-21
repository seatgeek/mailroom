// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/seatgeek/mailroom/mailroom/identifier"

	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/stretchr/testify/require"
)

func TestWriterNotifier_Push(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	notifier := NewWriterNotifier("buffer", buffer)

	err := notifier.Push(context.Background(), common.Notification{
		Type: "com.example.test",
		Message: common.RendererFunc(func(transport common.TransportID) string {
			return fmt.Sprintf("Hello, %s!", transport)
		}),
		Initiator: identifier.Collection{
			identifier.GenericUsername: "rufus",
		},
		Recipient: identifier.Collection{
			identifier.GenericUsername: "codell",
		},
	})

	require.NoError(t, err)
	require.Equal(t, "notification: type=com.example.test, from=map[username:rufus], to=map[username:codell], message=Hello, writer!\n", buffer.String())
}
