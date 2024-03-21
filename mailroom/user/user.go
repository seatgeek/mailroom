// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"fmt"

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
		Identifiers: make(identifier.Collection),
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
		u.Identifiers[id.NamespaceAndKind] = id.Value
	}
}

// WithIdentifiers adds multiple identifiers to a User
func WithIdentifiers(ids identifier.Collection) Option {
	return func(u *User) {
		for k, v := range ids {
			u.Identifiers[k] = v
		}
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
func (r *User) Wants(event common.EventType, transport common.TransportID) bool {
	if r.preferences[event] == nil {
		return false
	}

	return r.preferences[event][transport]
}

func (r *User) String() string {
	return fmt.Sprint(r.Identifiers)
}
