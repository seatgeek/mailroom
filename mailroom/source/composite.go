// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package source

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/event"
)

// PayloadParser is an interface for anything that parses incoming webhooks
type PayloadParser[T event.Payload] interface {
	// Parse verifies and parses incoming webhooks and returns a well-defined payload object or an error
	//
	// The payload return value contains the parsed payload struct, which will be passed to the NotificationGenerator.
	// Returning nil, nil is valid and indicates that the payload was parsed and determined to not be allowlisted, and
	// thus should be ignored.
	Parse(req *http.Request) (*event.Event[T], error)
}

// NotificationGenerator is an interface for anything that generates notifications from a parsed payload
type NotificationGenerator[T event.Payload] interface {
	// Generate takes a payload and returns a list of Notifications to be sent
	//
	// Some payloads may result in multiple notifications, for example the creation of a new merge request in GitLab
	// might result in notifications to multiple reviewers.
	Generate(event.Event[T]) ([]common.Notification, error)
	// EventTypes returns descriptors for all EventTypes that the generator may emit
	EventTypes() []event.TypeDescriptor
}

// compositeSource is a combination of a PayloadParser and a NotificationGenerator
// Both are required to be able to generate notifications from incoming webhooks, but they are kept separate to allow
// users to easily override the default generator with a custom one if needed.
type compositeSource[T event.Payload] struct {
	// Key is both a unique identifier for the source, and the endpoint that it listens on
	key       string
	Parser    PayloadParser[T]
	Generator NotificationGenerator[T]
}

func (s compositeSource[T]) Key() string {
	return s.key
}

func (s compositeSource[T]) EventTypes() []event.TypeDescriptor {
	return s.Generator.EventTypes()
}

func (s compositeSource[T]) Parse(req *http.Request) ([]common.Notification, error) {
	payload, err := s.Parser.Parse(req)
	if err != nil {
		slog.Error("failed to parse payload", "source", s.key, "error", err)
		return nil, err
	}

	if payload == nil {
		slog.Debug("ignoring uninteresting event", "source", s.key)
		return nil, nil
	}

	return s.Generator.Generate(*payload)
}

func (s compositeSource[T]) Validate(ctx context.Context) error {
	if v, ok := s.Parser.(common.Validator); ok {
		if err := v.Validate(ctx); err != nil {
			return fmt.Errorf("generator %s failed to validate: %w", s.Key(), err)
		}
	}

	if v, ok := s.Generator.(common.Validator); ok {
		if err := v.Validate(ctx); err != nil {
			return fmt.Errorf("generator %s failed to validate: %w", s.Key(), err)
		}
	}

	return nil
}

// New returns a new SourceImpl, pairing a PayloadParser and a NotificationGenerator together with some key
func New[T event.Payload](key string, parser PayloadParser[T], generator NotificationGenerator[T]) *compositeSource[T] {
	return &compositeSource[T]{
		key:       key,
		Parser:    parser,
		Generator: generator,
	}
}
