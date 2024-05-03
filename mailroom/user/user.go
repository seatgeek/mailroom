// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/event"
	"github.com/seatgeek/mailroom/mailroom/identifier"
)

// Wants returns true if the user wants to receive the given event via the given transport
// We assume that all preferences are opt-out by default; in other words, if a user has no preference
// for a given event, we assume they DO want it. We only return false if they have explicitly
// said they do not want it (and have false set in the map).
type Preferences map[event.Type]map[common.TransportKey]bool

func (p Preferences) Wants(event event.Type, transport common.TransportKey) bool {
	if _, exists := p[event]; !exists {
		// No preference set for this event, so assume they want it.
		return true
	}

	if want, exists := p[event][transport]; exists {
		return want
	}

	// No preference set for this transport, so assume they want it.
	return true
}

// User represents somebody who may receive notifications
type User struct {
	// Key is only used for indexing the user in the user store, e.g. for REST operations.
	Key string
	// Identifiers are unique attributes attached to a user that represent that user in the
	// scope of external systems, e.g. a gitlab.com/id or a slack.com/id.
	Identifiers identifier.Collection
	Preferences
}

// New creates a new User with the given options
func New(key string, options ...Option) *User {
	u := &User{
		Key:         key,
		Identifiers: identifier.NewCollection(),
		Preferences: make(Preferences),
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
func WithIdentifiers(ids identifier.Collection) Option {
	return func(u *User) {
		u.Identifiers.Merge(ids)
	}
}

// WithPreference adds a notification preference to a User
func WithPreference(event event.Type, transport common.TransportKey, wants bool) Option {
	return func(u *User) {
		if u.Preferences[event] == nil {
			u.Preferences[event] = make(map[common.TransportKey]bool)
		}

		u.Preferences[event][transport] = wants
	}
}

func WithPreferences(p Preferences) Option {
	return func(u *User) {
		u.Preferences = p
	}
}

func (r *User) String() string {
	return r.Identifiers.String()
}
