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

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/notifier/preference"
	"github.com/seatgeek/mailroom/pkg/user"
	"github.com/seatgeek/mailroom/pkg/validation"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	s := New()

	assert.NotNil(t, s)
	assert.Equal(t, "0.0.0.0:8000", s.listenAddr)
}

func TestServer_Options(t *testing.T) {
	t.Parallel()

	parser := event.NewMockParser(t)
	processor := event.NewMockProcessor(t)

	s := New(WithParser("foo", parser), WithProcessors(processor))

	assert.NotNil(t, s)
	assert.Equal(t, parser, s.parsers["foo"])
	assert.Contains(t, s.processors, processor)
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
			name: "returns error if a parser fails to validate",
			opts: []Opt{
				WithListenAddr(":0"),
				WithParser("foo", &parserThatFailsToValidate{err: errValidationFailed}),
			},
			wantErr: errValidationFailed,
		},
		{
			name: "returns error if a processors fails to validate",
			opts: []Opt{
				WithListenAddr(":0"),
				WithProcessors(&processorThatFailsToValidate{err: errValidationFailed}),
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
			ctx, cancel := context.WithCancel(t.Context())
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

type parserThatFailsToValidate struct {
	err error
}

var (
	_ event.Parser         = parserThatFailsToValidate{}
	_ validation.Validator = parserThatFailsToValidate{}
)

func (s parserThatFailsToValidate) Key() string {
	return "test"
}

func (s parserThatFailsToValidate) Parse(req *http.Request) (*event.Event, error) {
	panic("not called in our tests")
}

func (s parserThatFailsToValidate) EventTypes() []event.TypeDescriptor {
	return nil
}

func (s parserThatFailsToValidate) Validate(_ context.Context) error {
	return s.err
}

type processorThatFailsToValidate struct {
	err error
}

var (
	_ event.Processor      = processorThatFailsToValidate{}
	_ validation.Validator = processorThatFailsToValidate{}
)

func (p processorThatFailsToValidate) Process(ctx context.Context, evt event.Event, notifications []event.Notification) ([]event.Notification, error) {
	panic("not called in our tests")
}

func (p processorThatFailsToValidate) Validate(_ context.Context) error {
	return p.err
}

type transportThatFailsToValidate struct {
	err error
}

func (t transportThatFailsToValidate) Push(_ context.Context, _ event.Notification) error {
	panic("not called in our tests")
}

func (t transportThatFailsToValidate) Key() event.TransportKey {
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

func (s userStoreThatFailsToValidate) SetPreferences(_ context.Context, _ string, _ preference.Map) error {
	panic("not called in our tests")
}

func (s userStoreThatFailsToValidate) Validate(_ context.Context) error {
	return s.err
}
