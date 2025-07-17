// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notification

import (
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	slack2 "github.com/seatgeek/mailroom/pkg/notifier/slack"
	"github.com/slack-go/slack"
)

type builderOpts struct {
	context             event.Context
	recipients          identifier.Set
	fallbackMessage     string
	messagePerTransport map[event.TransportKey]string
	slackOpts           []slack.MsgOption
}

// Builder provides a fluent interface for constructing rich notification objects
type Builder struct {
	opts builderOpts
}

// NewBuilder creates a new fluent Builder instance
func NewBuilder(context event.Context) Builder {
	return Builder{
		opts: builderOpts{
			context:             context,
			recipients:          identifier.NewSet(),
			messagePerTransport: make(map[event.TransportKey]string),
		},
	}
}

// WithRecipient sets the recipient of the notification
// It's like WithRecipientIdentifiers, but it accepts a single identifier set
func (b Builder) WithRecipient(identifiers identifier.Set) Builder {
	newBuilder := b
	newBuilder.opts.recipients = identifiers
	return newBuilder
}

// WithRecipientIdentifiers sets the recipient of the notification
// It's like WithRecipient but it accepts multiple identifiers as variadic arguments
func (b Builder) WithRecipientIdentifiers(identifiers ...identifier.Identifier) Builder {
	newBuilder := b
	newBuilder.opts.recipients = identifier.NewSet(identifiers...)
	return newBuilder
}

// WithDefaultMessage sets the default message to be used if no message is provided for a specific transport
func (b Builder) WithDefaultMessage(message string) Builder {
	newBuilder := b
	newBuilder.opts.fallbackMessage = message
	return newBuilder
}

// WithMessageForTransport sets a specific message to be used for a specific transport
func (b Builder) WithMessageForTransport(transportKey event.TransportKey, message string) Builder {
	newBuilder := b
	// Copy the map to avoid shared references
	newMap := make(map[event.TransportKey]string, len(b.opts.messagePerTransport)+1)
	for k, v := range b.opts.messagePerTransport {
		newMap[k] = v
	}
	newMap[transportKey] = message
	newBuilder.opts.messagePerTransport = newMap
	return newBuilder
}

// WithSlackOptions sets the Slack options (like attachments, blocks, etc.) to be used when sending the notification
func (b Builder) WithSlackOptions(opts ...slack.MsgOption) Builder {
	newBuilder := b
	// Copy the slice to avoid shared references
	newSlackOpts := make([]slack.MsgOption, len(b.opts.slackOpts), len(b.opts.slackOpts)+len(opts))
	copy(newSlackOpts, b.opts.slackOpts)
	newSlackOpts = append(newSlackOpts, opts...)
	newBuilder.opts.slackOpts = newSlackOpts
	return newBuilder
}

// Build constructs the rich notification object from the previously set options
func (b Builder) Build() slack2.RichNotification {
	// Return a copy with mutable fields copied to maintain immutability
	return &builderOpts{
		context:         b.opts.context,
		recipients:      b.opts.recipients,
		fallbackMessage: b.opts.fallbackMessage,
		messagePerTransport: func() map[event.TransportKey]string {
			newMap := make(map[event.TransportKey]string, len(b.opts.messagePerTransport))
			for k, v := range b.opts.messagePerTransport {
				newMap[k] = v
			}
			return newMap
		}(),
		slackOpts: func() []slack.MsgOption {
			newSlackOpts := make([]slack.MsgOption, len(b.opts.slackOpts))
			copy(newSlackOpts, b.opts.slackOpts)
			return newSlackOpts
		}(),
	}
}

var _ slack2.RichNotification = &builderOpts{}

func (b *builderOpts) Context() event.Context {
	return b.context
}

func (b *builderOpts) Recipient() identifier.Set {
	return b.recipients
}

func (b *builderOpts) Render(key event.TransportKey) string {
	if message, ok := b.messagePerTransport[key]; ok {
		return message
	}

	return b.fallbackMessage
}

func (b *builderOpts) GetSlackOptions() []slack.MsgOption {
	return b.slackOpts
}
