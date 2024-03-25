// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package slack

import (
	"context"
	"fmt"

	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/identifier"
	"github.com/seatgeek/mailroom/mailroom/notifier"
	"github.com/slack-go/slack"
)

var ID = identifier.NewNamespaceAndKind("slack.com", identifier.KindID)

// Transport supports sending messages to Slack
type Transport struct {
	id     common.TransportID
	client *slack.Client
}

// RichNotification is an optional interface that can be implemented by any notification supporting Slack formatting
type RichNotification interface {
	common.Notification
	RenderSlack() []slack.MsgOption
}

// Push sends a notification to a Slack user
// In addition to supporting common.Notification, it also supports RichNotification for more complex messages
// that might include attachments, blocks, etc.
func (s *Transport) Push(ctx context.Context, notification common.Notification) error {
	id, ok := notification.Recipient().Get(ID)
	if !ok {
		return fmt.Errorf("%w: recipient does not have a Slack ID", notifier.ErrPermanentFailure)
	}

	options := s.getMessageOptions(notification)

	_, _, err := s.client.PostMessageContext(ctx, id, options...)

	return err
}

func (s *Transport) getMessageOptions(notification common.Notification) []slack.MsgOption {
	if n, ok := notification.(RichNotification); ok {
		opts := n.RenderSlack()
		if len(opts) > 0 {
			return opts
		}
	}

	return []slack.MsgOption{
		slack.MsgOptionText(notification.Render(s.id), false),
	}
}

func (s *Transport) ID() common.TransportID {
	return s.id
}

var _ notifier.Transport = &Transport{}

// NewTransport creates a new Slack Transport
// It requires a TransportID, a Slack API token, and optionally some slack.Options
func NewTransport(id common.TransportID, token string, opts ...slack.Option) *Transport {
	return &Transport{
		id:     id,
		client: slack.New(token, opts...),
	}
}
