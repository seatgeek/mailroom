// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"errors"

	"github.com/seatgeek/mailroom/mailroom/identifier"
)

var ErrUserNotFound = errors.New("user not found")

// Store is the interface that all user stores must implement
type Store interface {
	// Get returns a user by a specific identifier
	Get(identifier identifier.Identifier) (*User, error)

	// Find searches for a user matching any of the given identifiers
	// (The user is not required to match all of them, just one is enough)
	Find(possibleIdentifiers identifier.Collection) (*User, error)
}

// InMemoryStore is a simple in-memory implementation of the Store interface
// This is especially useful for testing
type InMemoryStore struct {
	users []*User
}

var _ Store = &InMemoryStore{}

// NewInMemoryStore creates a new in-memory store with the given users
func NewInMemoryStore(users ...*User) *InMemoryStore {
	return &InMemoryStore{users: users}
}

// Add adds a user to the in-memory store
func (s *InMemoryStore) Add(u *User) {
	s.users = append(s.users, u)
}

func (s *InMemoryStore) Get(identifier identifier.Identifier) (*User, error) {
	for _, u := range s.users {
		for k, v := range u.Identifiers {
			if k == identifier.NamespaceAndKind && v == identifier.Value {
				return u, nil
			}
		}
	}

	return nil, ErrUserNotFound
}

func (s *InMemoryStore) Find(possibleIdentifiers identifier.Collection) (*User, error) {
	for _, i := range possibleIdentifiers.ToList() {
		u, err := s.Get(i)
		if err == nil {
			return u, nil
		}
	}

	return nil, ErrUserNotFound
}
