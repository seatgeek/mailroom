// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"github.com/seatgeek/mailroom/mailroom/common"
)

type User struct {
	identifiers []common.Identifier
}

func New(identifiers ...common.Identifier) *User {
	return &User{
		identifiers: identifiers,
	}
}

func (r *User) AddIdentifier(identifier common.Identifier) {
	r.identifiers = append(r.identifiers, identifier)
}

// GetIdentifier returns the first Identifier that matches the given query
// For example, to find any email address, pass an NamespaceAndKind with Kind="email".
// Any query field that is empty will act as a wildcard.
func (r *User) GetIdentifier(query common.NamespaceAndKind) (common.Identifier, bool) {
	for _, id := range r.identifiers {
		if query.Namespace != "" && query.Namespace != id.Namespace {
			continue
		}

		if query.Kind != "" && query.Kind != id.Kind {
			continue
		}

		return id, true
	}

	return common.Identifier{}, false
}
