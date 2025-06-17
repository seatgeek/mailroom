// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

// Package user provides types and functions for managing users who may want to receive notifications
package user

import (
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/notifier/preference"
)

// User is somebody who may receive notifications and their preferences for receiving those.
type User struct {
	// Key is only used for indexing the user in the user store, e.g. for REST operations.
	Key string
	// Identifiers are unique attributes attached to a user that represent that user in the
	// scope of external systems, e.g. a gitlab.com/id or a slack.com/id.
	Identifiers identifier.Set
	Preferences preference.Map
}

// New creates a new User with the given options
func New(key string, options ...Option) *User {
	u := &User{
		Key:         key,
		Identifiers: identifier.NewSet(),
		Preferences: make(preference.Map),
	}

	for _, opt := range options {
		opt(u)
	}

	return u
}

type Option func(*User)

// WithIdentifier adds an identifier to a User
func WithIdentifier(id identifier.Identifier) Option {
	return func(u *User) {
		u.Identifiers.Add(id)
	}
}

// WithIdentifiers adds multiple identifiers to a User
func WithIdentifiers(ids identifier.Set) Option {
	return func(u *User) {
		u.Identifiers.Merge(ids)
	}
}

// WithPreference adds a notification preference to a User
func WithPreference(evt event.Type, transport event.TransportKey, wants bool) Option {
	return func(u *User) {
		if u.Preferences[evt] == nil {
			u.Preferences[evt] = make(map[event.TransportKey]bool)
		}

		u.Preferences[evt][transport] = wants
	}
}

// WithPreferences sets all the Provider for a User
func WithPreferences(p preference.Map) Option {
	return func(u *User) {
		u.Preferences = p
	}
}

// String returns a simple string representation of a User's identify (useful for logging)
func (r *User) String() string {
	if (r == nil) || (r.Identifiers == nil) {
		return "<unknown>"
	}

	return r.Identifiers.String()
}
