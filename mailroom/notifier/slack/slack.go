// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package slack

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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
		return notifier.Permanent(errors.New("recipient does not have a Slack ID"))
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

func (s *Transport) Validate() error {
	resp, err := s.client.AuthTest()
	if err != nil {
		return notifier.Permanent(fmt.Errorf("authentication failed: %w", err))
	}

	slog.Info("Slack transport connected", "transport", s.id, "slack_team", resp.Team, "slack_user", resp.User)
	return nil
}

var _ notifier.Transport = &Transport{}
var _ common.Validator = &Transport{}

// NewTransport creates a new Slack Transport
// It requires a TransportID, a Slack API token, and optionally some slack.Options
func NewTransport(id common.TransportID, token string, opts ...slack.Option) *Transport {
	return &Transport{
		id:     id,
		client: slack.New(token, opts...),
	}
}
