// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package source_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/event"
	"github.com/seatgeek/mailroom/mailroom/source"
	"github.com/stretchr/testify/assert"
)

func TestCompositeSource_Parse(t *testing.T) {
	t.Parallel()

	somePayload := event.Event[any]{
		Data: struct{}{},
	}

	someNotifications := []common.Notification{
		common.NewMockNotification(t),
	}

	someError := errors.New("something failed")

	tests := []struct {
		name              string
		parser            source.PayloadParser[any]
		generator         source.NotificationGenerator[any]
		wantNotifications []common.Notification
		wantErr           error
	}{
		{
			name:              "happy path",
			parser:            fakeParser{Returns: &somePayload},
			generator:         fakeGenerator{Generates: someNotifications},
			wantNotifications: someNotifications,
			wantErr:           nil,
		},
		{
			name: "no interesting event",
			parser: fakeParser{
				Returns: nil,
			},
			wantNotifications: nil,
			wantErr:           nil,
		},
		{
			name:    "parse error",
			parser:  fakeParser{ReturnsError: someError},
			wantErr: someError,
		},
		{
			name:      "generator error",
			parser:    fakeParser{Returns: &somePayload},
			generator: fakeGenerator{GeneratesError: someError},
			wantErr:   someError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			src := source.New[any]("foo", tc.parser, tc.generator)

			got, gotErr := src.Parse(&http.Request{})

			assert.Equal(t, tc.wantNotifications, got)
			assert.Equal(t, tc.wantErr, gotErr)
		})
	}
}

func TestCompositeSource_EventTypes(t *testing.T) {
	t.Parallel()

	someEventTypes := []event.TypeDescriptor{
		{
			Key:         "foo",
			Title:       "Foo",
			Description: "Foo bar baz",
		},
	}

	src := source.New[any]("foo", fakeParser{}, fakeGenerator{ReturnsEventTypes: someEventTypes})

	assert.Equal(t, someEventTypes, src.EventTypes())
}

func TestCompositeSource_Validate(t *testing.T) {
	t.Parallel()

	someValidationError := errors.New("some error")

	tests := []struct {
		name      string
		parser    source.PayloadParser[any]
		generator source.NotificationGenerator[any]
		want      error
	}{
		{
			name:      "happy path",
			parser:    fakeParser{Validates: nil},
			generator: fakeGenerator{Validates: nil},
			want:      nil,
		},
		{
			name:      "parser validation error",
			parser:    fakeParser{Validates: someValidationError},
			generator: fakeGenerator{Validates: nil},
			want:      someValidationError,
		},
		{
			name:      "generator validation error",
			parser:    fakeParser{Validates: nil},
			generator: fakeGenerator{Validates: someValidationError},
			want:      someValidationError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			src := source.New[any]("foo", tc.parser, tc.generator)

			got := src.Validate(context.Background())

			assert.ErrorIs(t, got, tc.want)
		})
	}
}

type fakeParser struct {
	Returns      *event.Event[any]
	ReturnsError error
	Validates    error
}

func (f fakeParser) Parse(_ *http.Request) (*event.Event[any], error) {
	return f.Returns, f.ReturnsError
}

func (f fakeParser) Validate(_ context.Context) error {
	return f.Validates
}

var _ source.PayloadParser[any] = (*fakeParser)(nil)
var _ common.Validator = (*fakeParser)(nil)

type fakeGenerator struct {
	Generates         []common.Notification
	GeneratesError    error
	ReturnsEventTypes []event.TypeDescriptor
	Validates         error
}

func (f fakeGenerator) Generate(_ event.Event[any]) ([]common.Notification, error) {
	return f.Generates, f.GeneratesError
}

func (f fakeGenerator) EventTypes() []event.TypeDescriptor {
	return f.ReturnsEventTypes
}

func (f fakeGenerator) Validate(_ context.Context) error {
	return f.Validates
}

var _ source.NotificationGenerator[any] = (*fakeGenerator)(nil)
var _ common.Validator = (*fakeGenerator)(nil)
