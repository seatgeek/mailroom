// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/identifier"
)

// User represents somebody who may receive notifications
type User struct {
	Identifiers identifier.Collection
	preferences map[common.EventType]map[common.TransportID]bool
}

// New creates a new User with the given options
func New(options ...Option) *User {
	u := &User{
		Identifiers: identifier.NewCollection(),
		preferences: make(map[common.EventType]map[common.TransportID]bool),
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
func WithPreference(event common.EventType, transport common.TransportID, wants bool) Option {
	return func(u *User) {
		if u.preferences[event] == nil {
			u.preferences[event] = make(map[common.TransportID]bool)
		}

		u.preferences[event][transport] = wants
	}
}

// Wants returns true if the user wants to receive the given event via the given transport
// We assume that all preferences are opt-out by default; in other words, if a user has no preference
// for a given event, we assume they DO want it. We only return false if they have explicitly
// said they do not want it (and have false set in the map).
func (r *User) Wants(event common.EventType, transport common.TransportID) bool {
	if _, exists := r.preferences[event]; !exists {
		// No preference set for this event, so assume they want it.
		return true
	}

	if want, exists := r.preferences[event][transport]; exists {
		return want
	}

	// No preference set for this transport, so assume they want it.
	return true
}

func (r *User) String() string {
	return r.Identifiers.String()
}
