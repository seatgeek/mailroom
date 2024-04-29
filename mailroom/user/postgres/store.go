// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package postgres

import (
	"fmt"
	"time"

	"github.com/seatgeek/mailroom/mailroom/identifier"
	"github.com/seatgeek/mailroom/mailroom/user"
	"gorm.io/gorm"
)

type UserModel struct {
	Key         string           `gorm:"primarykey"`
	Preferences user.Preferences `gorm:"serializer:json"`

	// Identifiers is a map of all identifiers for the user
	Identifiers map[identifier.NamespaceAndKind]string `gorm:"serializer:json"`
	// Emails contains the subset of Identifiers that have Kind=="email" (for easier fallback lookup)
	Emails []string `gorm:"serializer:json"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
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

// Add upserts a user to the postgres store
func (s *Store) Add(u *user.User) error {
	var emails []string
	for _, id := range u.Identifiers.ToList() {
		if id.Kind() == identifier.KindEmail {
			emails = append(emails, id.Value)
		}
	}

	result := s.db.Save(&UserModel{
		Key:         u.Key,
		Preferences: u.Preferences,
		Identifiers: u.Identifiers.ToMap(),
		Emails:      emails,
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

	if len(users) > 1 {
		return nil, fmt.Errorf("found multiple users with identifiers: %v", possibleIdentifiers)
	}

	if len(users) == 1 {
		return users[0].ToUser(), nil
	}

	// No users found; fall back to email identifiers if possible
	possibleEmails := make(map[string]struct{})
	for _, id := range possibleIdentifiers.ToList() {
		if id.Kind() == identifier.KindEmail {
			possibleEmails[id.Value] = struct{}{}
		}
	}

	query = s.db.Model(&UserModel{})
	for email := range possibleEmails {
		query = query.Or("emails @> ?", fmt.Sprintf(`"%s"`, email))
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}

	if len(users) > 1 {
		return nil, fmt.Errorf("found multiple users with email identifiers: %v", possibleIdentifiers)
	}

	if len(users) == 1 {
		return users[0].ToUser(), nil
	}

	return nil, user.ErrUserNotFound
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
func (s *Store) GetByIdentifier(id identifier.Identifier) (*user.User, error) {
	var u UserModel
	if err := s.db.Where("identifiers @> ?", fmt.Sprintf(`{"%s": "%s"}`, id.NamespaceAndKind, id.Value)).First(&u).Error; err == nil {
		return u.ToUser(), nil
	}

	// Fall back to any email identifier
	if id.Kind() == identifier.KindEmail {
		if err := s.db.Where("emails @> ?", fmt.Sprintf(`"%s"`, id.Value)).First(&u).Error; err == nil {
			return u.ToUser(), nil
		}
	}

	return nil, user.ErrUserNotFound
}

// SetPreferences implements user.Store.
func (s *Store) SetPreferences(key string, prefs user.Preferences) error {
	return s.db.Model(&UserModel{}).Where("key = ?", key).Update("preferences", prefs).Error
}

var _ user.Store = &Store{}
