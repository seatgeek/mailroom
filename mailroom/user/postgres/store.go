// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package postgres

import (
	"fmt"

	"github.com/seatgeek/mailroom/mailroom/identifier"
	"github.com/seatgeek/mailroom/mailroom/user"
	"gorm.io/gorm"
)

type UserModel struct {
	gorm.Model

	Key         string                                 `gorm:"unique"`
	Preferences user.Preferences                       `gorm:"serializer:json"`
	Identifiers map[identifier.NamespaceAndKind]string `gorm:"serializer:json"`
}

func (u *UserModel) TableName() string {
	return "users"
}

func (u *UserModel) ToUser() *user.User {
	return &user.User{
		Key:         u.Key,
		Preferences: u.Preferences,
		Identifiers: identifier.NewCollectionFromMap(u.Identifiers),
	}
}

type Store struct {
	db *gorm.DB
}

// NewPostgresStore creates a new postgres store
func NewPostgresStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

// Add adds a user to the postgres store
func (s *Store) Add(u *user.User) error {
	result := s.db.Create(&UserModel{
		Key:         u.Key,
		Preferences: u.Preferences,
		Identifiers: u.Identifiers.ToMap(),
	})
	return result.Error
}

// Find implements user.Store.
func (s *Store) Find(possibleIdentifiers identifier.Collection) (*user.User, error) {
	query := s.db.Model(&UserModel{})
	for _, id := range possibleIdentifiers.ToList() {
		query = query.Or("identifiers @> ?", fmt.Sprintf(`{"%s": "%s"}`, id.NamespaceAndKind, id.Value))
	}

	var users []UserModel
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, user.ErrUserNotFound
	}

	if len(users) > 1 {
		return nil, fmt.Errorf("found multiple users with identifiers: %v", possibleIdentifiers)
	}

	return users[0].ToUser(), nil
}

// Get implements user.Store.
func (s *Store) Get(key string) (*user.User, error) {
	var u UserModel
	if err := s.db.Where("key = ?", key).First(&u).Error; err != nil {
		return nil, err
	}

	return u.ToUser(), nil
}

// GetByIdentifier implements user.Store.
func (s *Store) GetByIdentifier(identifier identifier.Identifier) (*user.User, error) {
	var u UserModel
	if err := s.db.Where("identifiers @> ?", fmt.Sprintf(`{"%s": "%s"}`, identifier.NamespaceAndKind, identifier.Value)).First(&u).Error; err != nil {
		return nil, err
	}

	return u.ToUser(), nil
}

// SetPreferences implements user.Store.
func (s *Store) SetPreferences(key string, prefs user.Preferences) error {
	return s.db.Model(&UserModel{}).Where("key = ?", key).Update("preferences", prefs).Error
}

var _ user.Store = &Store{}
