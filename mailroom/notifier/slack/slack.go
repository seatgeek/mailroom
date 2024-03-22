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

type Transport struct {
	id     common.TransportID
	client *slack.Client
}

func (s *Transport) Push(ctx context.Context, notification common.Notification) error {
	id, ok := notification.Recipient[ID]
	if !ok {
		return fmt.Errorf("%w: recipient does not have a Slack ID", notifier.ErrPermanentFailure)
	}

	body := notification.Message.Render(s.id)

	_, _, err := s.client.PostMessageContext(ctx, id, slack.MsgOptionText(body, false))

	return err
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
