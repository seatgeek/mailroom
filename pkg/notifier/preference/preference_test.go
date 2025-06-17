// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package preference_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/notifier/preference"
	"github.com/seatgeek/mailroom/pkg/validation"
	"github.com/stretchr/testify/assert"
)

func TestChain_Wants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		chain preference.Chain
		want  *bool
	}{
		{
			name:  "empty chain",
			chain: preference.Chain{},
			want:  nil,
		},
		{
			name: "return Deliver if first preference returns true",
			chain: preference.Chain{
				preference.Default(true),
				preference.Default(false),
			},
			want: ptrTo(true),
		},
		{
			name: "return DoNotDeliver if first preference returns false",
			chain: preference.Chain{
				preference.Default(false),
				preference.Default(true),
			},
			want: ptrTo(false),
		},
		{
			name: "fallback to next preference if first returns nil",
			chain: preference.Chain{
				preference.Func(func(context.Context, event.Notification, event.TransportKey) *bool {
					return nil
				}),
				preference.Default(true),
			},
			want: ptrTo(true),
		},
		{
			name: "return nil if no preferences match",
			chain: preference.Chain{
				preference.Func(func(context.Context, event.Notification, event.TransportKey) *bool {
					return nil
				}),
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.chain.Wants(t.Context(), event.NewMockNotification(t), "some-transport-key")
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestChain_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		chain   preference.Chain
		wantErr string
	}{
		{
			name:    "empty chain is valid",
			chain:   preference.Chain{},
			wantErr: "",
		},
		{
			name: "all preferences valid",
			chain: preference.Chain{
				preferenceThatValidates{},
				preferenceThatValidates{},
			},
			wantErr: "",
		},
		{
			name: "some preferences invalid",
			chain: preference.Chain{
				preferenceThatValidates{err: errors.New("first preference invalid")},
				preferenceThatValidates{},
				preferenceThatValidates{err: errors.New("third preference invalid")},
			},
			wantErr: "first preference invalid\nthird preference invalid",
		},
		{
			name: "all preferences invalid",
			chain: preference.Chain{
				preferenceThatValidates{err: errors.New("first preference invalid")},
				preferenceThatValidates{err: errors.New("second preference invalid")},
			},
			wantErr: "first preference invalid\nsecond preference invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.chain.Validate(t.Context())
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

type preferenceThatValidates struct {
	err error
}

var (
	_ preference.Provider  = preferenceThatValidates{}
	_ validation.Validator = preferenceThatValidates{}
)

func (p preferenceThatValidates) Wants(_ context.Context, _ event.Notification, _ event.TransportKey) *bool {
	return nil
}

func (p preferenceThatValidates) Validate(_ context.Context) error {
	return p.err
}

func TestMap(t *testing.T) {
	t.Parallel()

	prefMap := preference.Map{
		"some-event-type": {
			"slack": true,
		},
	}

	tests := []struct {
		name      string
		eventType event.Type
		transport event.TransportKey
		want      *bool
	}{
		{
			name:      "event type and transport are defined",
			eventType: "some-event-type",
			transport: "slack",
			want:      ptrTo(true),
		},
		{
			name:      "event type defined, transport not defined",
			eventType: "some-event-type",
			transport: "email",
			want:      nil,
		},
		{
			name:      "event type not defined",
			eventType: "another-event-type",
			transport: "slack",
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			eventContext := event.Context{
				Type: tt.eventType,
			}

			notification := event.NewMockNotification(t)
			notification.EXPECT().Context().Return(eventContext).Maybe()

			got := prefMap.Wants(t.Context(), notification, tt.transport)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value bool
		want  *bool
	}{
		{
			value: true,
			want:  ptrTo(true),
		},
		{
			value: false,
			want:  ptrTo(false),
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("returns %v", tt.value), func(t *testing.T) {
			t.Parallel()

			pref := preference.Default(tt.value)
			got := pref.Wants(t.Context(), event.NewMockNotification(t), "some-transport-key")
			assert.Equal(t, tt.want, got)
		})
	}
}

func ptrTo[T any](v T) *T {
	return &v
}
