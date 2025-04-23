// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package mailroom

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/seatgeek/mailroom/pkg/common"
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/handler"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/user"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	s := New()

	assert.NotNil(t, s)
	assert.Equal(t, "0.0.0.0:8000", s.listenAddr)
}

func TestWithHandlers(t *testing.T) {
	t.Parallel()

	src1 := handler.NewMockHandler(t)
	src1.EXPECT().Key().Return("foo").Maybe()
	src2 := handler.NewMockHandler(t)
	src2.EXPECT().Key().Return("bar").Maybe()

	s := New(WithHandlers(src1, src2))

	assert.NotNil(t, s)
	assert.Contains(t, s.handlers, src1)
	assert.Contains(t, s.handlers, src2)
}

func TestRun(t *testing.T) {
	t.Parallel()

	errValidationFailed := errors.New("some validation failed error")

	tests := []struct {
		name    string
		opts    []Opt
		wantErr error
	}{
		{
			name: "starts and shuts down",
			opts: []Opt{
				WithListenAddr(":0"),
			},
			wantErr: nil,
		},
		{
			name: "returns error if a handler fails to validate",
			opts: []Opt{
				WithListenAddr(":0"),
				WithHandlers(&handlerThatFailsToValidate{err: errValidationFailed}),
			},
			wantErr: errValidationFailed,
		},
		{
			name: "returns error if a transport fails to validate",
			opts: []Opt{
				WithListenAddr(":0"),
				WithTransports(&transportThatFailsToValidate{
					err: errValidationFailed,
				}),
			},
			wantErr: errValidationFailed,
		},
		{
			name: "returns error if a user store fails to validate",
			opts: []Opt{
				WithListenAddr(":0"),
				WithUserStore(&userStoreThatFailsToValidate{
					err: errValidationFailed,
				}),
			},
			wantErr: errValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := New(tt.opts...)
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				time.Sleep(500 * time.Millisecond)
				cancel()
			}()

			err := s.Run(ctx)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

type handlerThatFailsToValidate struct {
	err error
}

var (
	_ handler.Handler  = handlerThatFailsToValidate{}
	_ common.Validator = handlerThatFailsToValidate{}
)

func (s handlerThatFailsToValidate) Validate(_ context.Context) error {
	return s.err
}

func (s handlerThatFailsToValidate) Key() string {
	return "some-handler"
}

func (s handlerThatFailsToValidate) Process(_ *http.Request) ([]common.Notification, error) {
	panic("not implemented")
}

func (s handlerThatFailsToValidate) EventTypes() []event.TypeDescriptor {
	panic("not implemented")
}

type transportThatFailsToValidate struct {
	err error
}

func (t transportThatFailsToValidate) Push(_ context.Context, _ common.Notification) error {
	panic("not called in our tests")
}

func (t transportThatFailsToValidate) Key() common.TransportKey {
	return "test"
}

func (t transportThatFailsToValidate) Validate(_ context.Context) error {
	return t.err
}

type userStoreThatFailsToValidate struct {
	err error
}

func (s userStoreThatFailsToValidate) Get(_ context.Context, _ string) (*user.User, error) {
	panic("not called in our tests")
}

func (s userStoreThatFailsToValidate) GetByIdentifier(_ context.Context, identifier identifier.Identifier) (*user.User, error) {
	panic("not called in our tests")
}

func (s userStoreThatFailsToValidate) Find(_ context.Context, _ identifier.Set) (*user.User, error) {
	panic("not called in our tests")
}

func (s userStoreThatFailsToValidate) SetPreferences(_ context.Context, _ string, _ user.Preferences) error {
	panic("not called in our tests")
}

func (s userStoreThatFailsToValidate) Validate(_ context.Context) error {
	return s.err
}
