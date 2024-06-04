// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"errors"

	"github.com/seatgeek/mailroom/mailroom/identifier"
)

// ErrUserNotFound is returned when a user is not found in the store
// It should only be returned when the lookup process itself succeeded (e.g. no db connection errors), but
// we failed to locate a single known user in the store. Think of it like a 404 error.
var ErrUserNotFound = errors.New("user not found")

// Store is the interface that all user stores must implement
// For all methods that search by identifier, the store MUST return the user that matches the identifier exactly.
// If no user matches the exact identifier, and the identifier is an "email" kind, the store SHOULD attempt to find
// a user where any email identifier matches the given email, regardless of namespace. This will allow for onboarding
// new integrations that utilize email identifiers without having to update all existing user information in the store.
type Store interface {
	// Get returns a user by its key
	Get(key string) (*User, error)
	// Get returns a user by a given identifier
	GetByIdentifier(identifier identifier.Identifier) (*User, error)

	// Find searches for a user matching any of the given identifiers
	// (The user is not required to match all of them, just one is enough)
	Find(possibleIdentifiers identifier.Collection) (*User, error)

	SetPreferences(key string, prefs Preferences) error
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
func (s *InMemoryStore) Add(u *User) error {
	s.users = append(s.users, u)
	return nil
}

func (s *InMemoryStore) Get(key string) (*User, error) {
	for _, u := range s.users {
		if u.Key == key {
			return u, nil
		}
	}

	return nil, ErrUserNotFound
}

func (s *InMemoryStore) GetByIdentifier(identifier identifier.Identifier) (*User, error) {
	isEmail := identifier.Kind() == "email"

	for _, u := range s.users {
		for _, existing := range u.Identifiers.ToList() {
			// Look for an exact match
			if existing.NamespaceAndKind == identifier.NamespaceAndKind && existing.Value == identifier.Value {
				return u, nil
			}

			// Or if the identifier is an email, look for any matching email
			if isEmail && existing.Kind() == "email" && existing.Value == identifier.Value {
				return u, nil
			}
		}
	}

	return nil, ErrUserNotFound
}

func (s *InMemoryStore) Find(possibleIdentifiers identifier.Collection) (*User, error) {
	for _, i := range possibleIdentifiers.ToList() {
		u, err := s.GetByIdentifier(i)
		if err == nil {
			return u, nil
		}
	}

	return nil, ErrUserNotFound
}

func (s *InMemoryStore) SetPreferences(key string, prefs Preferences) error {
	u, err := s.Get(key)
	if err != nil {
		return err
	}
	u.Preferences = prefs
	return nil
}
