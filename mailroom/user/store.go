// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"errors"

	"github.com/seatgeek/mailroom/mailroom/common"
)

var ErrUserNotFound = errors.New("user not found")

type Store interface {
	Get(identifier common.Identifier) (*User, error)
}

type InMemoryStore struct {
	users []*User
}

func NewInMemoryStore(users ...*User) *InMemoryStore {
	return &InMemoryStore{users: users}
}

func (s *InMemoryStore) Add(u *User) {
	s.users = append(s.users, u)
}

func (s *InMemoryStore) Get(identifier common.Identifier) (*User, error) {
	for _, u := range s.users {
		for _, id := range u.identifiers {
			if id == identifier {
				return u, nil
			}
		}
	}

	return nil, ErrUserNotFound
}
