// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notification

import (
	"github.com/seatgeek/mailroom/pkg/common"
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	slack2 "github.com/seatgeek/mailroom/pkg/notifier/slack"
	"github.com/slack-go/slack"
)

type builderOpts struct {
	context             event.Context
	recipients          identifier.Set
	fallbackMessage     string
	messagePerTransport map[common.TransportKey]string
	slackOpts           []slack.MsgOption
}

// Builder provides a fluent interface for constructing rich notification objects
type Builder struct {
	opts builderOpts
}

// NewBuilder creates a new fluent Builder instance
func NewBuilder(context event.Context) *Builder {
	return &Builder{
		opts: builderOpts{
			context:             context,
			recipients:          identifier.NewSet(),
			messagePerTransport: make(map[common.TransportKey]string),
		},
	}
}

// WithRecipient sets the recipient of the notification
// It's like WithRecipientIdentifiers, but it accepts a single identifier set
func (b *Builder) WithRecipient(identifiers identifier.Set) *Builder {
	b.opts.recipients = identifiers
	return b
}

// WithRecipientIdentifiers sets the recipient of the notification
// It's like WithRecipient but it accepts multiple identifiers as variadic arguments
func (b *Builder) WithRecipientIdentifiers(identifiers ...identifier.Identifier) *Builder {
	b.opts.recipients = identifier.NewSet(identifiers...)
	return b
}

// WithDefaultMessage sets the default message to be used if no message is provided for a specific transport
func (b *Builder) WithDefaultMessage(message string) *Builder {
	b.opts.fallbackMessage = message
	return b
}

// WithMessageForTransport sets a specific message to be used for a specific transport
func (b *Builder) WithMessageForTransport(transportKey common.TransportKey, message string) *Builder {
	b.opts.messagePerTransport[transportKey] = message
	return b
}

// WithSlackOptions sets the Slack options (like attachments, blocks, etc.) to be used when sending the notification
func (b *Builder) WithSlackOptions(opts ...slack.MsgOption) *Builder {
	b.opts.slackOpts = opts
	return b
}

// Build constructs the rich notification object from the previously set options
func (b *Builder) Build() slack2.RichNotification {
	return &b.opts
}

var _ slack2.RichNotification = &builderOpts{}

func (b *builderOpts) Context() event.Context {
	return b.context
}

func (b *builderOpts) Recipient() identifier.Set {
	return b.recipients
}

func (b *builderOpts) Render(key common.TransportKey) string {
	if message, ok := b.messagePerTransport[key]; ok {
		return message
	}

	return b.fallbackMessage
}

func (b *builderOpts) GetSlackOptions() []slack.MsgOption {
	return b.slackOpts
}
