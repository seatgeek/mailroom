// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"github.com/seatgeek/mailroom/mailroom/identifier"
)

type User struct {
	Identifiers identifier.Collection
}

func New(options ...Option) *User {
	u := &User{
		Identifiers: make(identifier.Collection),
	}

	for _, opt := range options {
		opt(u)
	}

	return u
}

type Option func(*User)

func WithIdentifier(id identifier.Identifier) Option {
	return func(u *User) {
		u.Identifiers[id.NamespaceAndKind] = id.Value
	}
}

func WithIdentifiers(ids identifier.Collection) Option {
	return func(u *User) {
		for k, v := range ids {
			u.Identifiers[k] = v
		}
	}
}
