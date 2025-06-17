// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

// Package postgres provides a postgresql-backed implementation of the user.Store interface
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/notifier/preference"
	"github.com/seatgeek/mailroom/pkg/user"
	"gorm.io/gorm"
)

// UserModel is the gorm model for a user
type UserModel struct {
	Key         string         `gorm:"primarykey"`
	Preferences preference.Map `gorm:"serializer:json"`

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

// ToUser converts a UserModel to a user.User
func (u *UserModel) ToUser() *user.User {
	return &user.User{
		Key:         u.Key,
		Preferences: u.Preferences,
		Identifiers: identifier.NewSetFromMap(u.Identifiers),
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
func (s *Store) Add(ctx context.Context, u *user.User) error {
	var emails []string
	for _, id := range u.Identifiers.ToList() {
		if id.Kind() == identifier.KindEmail {
			emails = append(emails, id.Value)
		}
	}

	result := s.db.WithContext(ctx).Save(&UserModel{
		Key:         u.Key,
		Preferences: u.Preferences,
		Identifiers: u.Identifiers.ToMap(),
		Emails:      emails,
	})
	return result.Error
}

// Find implements user.Store.
func (s *Store) Find(ctx context.Context, possibleIdentifiers identifier.Set) (*user.User, error) {
	if possibleIdentifiers.Len() == 0 {
		return nil, fmt.Errorf("%w: no identifiers provided", user.ErrUserNotFound)
	}

	query := s.db.WithContext(ctx).Model(&UserModel{})
	for _, id := range possibleIdentifiers.ToList() {
		query = query.Or("identifiers @> ?", fmt.Sprintf(`{"%s": "%s"}`, id.NamespaceAndKind, id.Value))
	}

	var users []UserModel
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}

	if len(users) > 1 {
		return nil, fmt.Errorf("%w: found multiple users with identifiers %v", user.ErrUserNotFound, possibleIdentifiers)
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

	if len(possibleEmails) == 0 {
		return nil, fmt.Errorf("%w: no identifiers matched and no fallback emails were available", user.ErrUserNotFound)
	}

	query = s.db.WithContext(ctx).Model(&UserModel{})
	for email := range possibleEmails {
		query = query.Or("emails @> ?", fmt.Sprintf(`"%s"`, email))
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}

	if len(users) > 1 {
		return nil, fmt.Errorf("%w: found multiple users with email identifiers %v", user.ErrUserNotFound, possibleEmails)
	}

	if len(users) == 1 {
		return users[0].ToUser(), nil
	}

	return nil, user.ErrUserNotFound
}

// Get implements user.Store.
func (s *Store) Get(ctx context.Context, key string) (*user.User, error) {
	var u UserModel
	if err := s.db.WithContext(ctx).Where("key = ?", key).First(&u).Error; err != nil {
		return nil, err
	}

	return u.ToUser(), nil
}

// GetByIdentifier implements user.Store.
func (s *Store) GetByIdentifier(ctx context.Context, id identifier.Identifier) (*user.User, error) {
	var u UserModel
	if err := s.db.WithContext(ctx).Where("identifiers @> ?", fmt.Sprintf(`{"%s": "%s"}`, id.NamespaceAndKind, id.Value)).First(&u).Error; err == nil {
		return u.ToUser(), nil
	}

	// Fall back to any email identifier
	if id.Kind() == identifier.KindEmail {
		if err := s.db.WithContext(ctx).Where("emails @> ?", fmt.Sprintf(`"%s"`, id.Value)).First(&u).Error; err == nil {
			return u.ToUser(), nil
		}
	}

	return nil, user.ErrUserNotFound
}

// SetPreferences implements user.Store.
func (s *Store) SetPreferences(ctx context.Context, key string, prefs preference.Map) error {
	return s.db.WithContext(ctx).Model(&UserModel{}).Where("key = ?", key).Update("preferences", prefs).Error
}

var _ user.Store = &Store{}
