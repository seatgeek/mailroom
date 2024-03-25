// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notification

import (
	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/identifier"
	slack2 "github.com/seatgeek/mailroom/mailroom/notifier/slack"
	"github.com/slack-go/slack"
)

type builderOpts struct {
	eventType           common.EventType
	recipients          identifier.Collection
	fallbackMessage     string
	messagePerTransport map[common.TransportID]string
	slackOpts           []slack.MsgOption
}

type Builder struct {
	opts builderOpts
}

func NewBuilder(eventType common.EventType) *Builder {
	return &Builder{
		opts: builderOpts{
			eventType:           eventType,
			recipients:          identifier.NewCollection(),
			messagePerTransport: make(map[common.TransportID]string),
		},
	}
}

func (b *Builder) WithRecipients(recipients identifier.Collection) *Builder {
	b.opts.recipients = recipients
	return b
}

func (b *Builder) WithRecipient(recipient identifier.Identifier) *Builder {
	b.opts.recipients = identifier.NewCollection(recipient)
	return b
}

func (b *Builder) WithDefaultMessage(message string) *Builder {
	b.opts.fallbackMessage = message
	return b
}

func (b *Builder) WithMessageForTransport(transportID common.TransportID, message string) *Builder {
	b.opts.messagePerTransport[transportID] = message
	return b
}

func (b *Builder) WithSlackMessage(opts ...slack.MsgOption) *Builder {
	b.opts.slackOpts = opts
	return b
}

func (b *Builder) Build() slack2.RichNotification {
	return &b.opts
}

var _ slack2.RichNotification = &builderOpts{}

func (b *builderOpts) Type() common.EventType {
	return b.eventType
}

func (b *builderOpts) Recipients() identifier.Collection {
	return b.recipients
}

func (b *builderOpts) Render(id common.TransportID) string {
	if message, ok := b.messagePerTransport[id]; ok {
		return message
	}

	return b.fallbackMessage
}

func (b *builderOpts) AddRecipients(collection identifier.Collection) {
	b.recipients.Merge(collection)
}

func (b *builderOpts) RenderSlack() []slack.MsgOption {
	return b.slackOpts
}
