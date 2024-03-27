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

	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/identifier"
	"github.com/seatgeek/mailroom/mailroom/source"
	"github.com/seatgeek/mailroom/mailroom/user"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	s := New()

	assert.NotNil(t, s)
	assert.Equal(t, "0.0.0.0:8000", s.listenAddr)
}

func TestWithSources(t *testing.T) {
	src1 := &source.Source{ID: "foo"}
	src2 := &source.Source{ID: "bar"}
	s := New(WithSources(src1, src2))

	assert.NotNil(t, s)
	assert.Contains(t, s.sources, src1)
	assert.Contains(t, s.sources, src2)
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
				WithSources(&source.Source{
					ID: "foo",
					Parser: &parserThatFailsToValidate{
						err: errValidationFailed,
					},
				}),
			},
			wantErr: errValidationFailed,
		},
		{
			name: "returns error if a generator fails to validate",
			opts: []Opt{
				WithListenAddr(":0"),
				WithSources(&source.Source{
					ID: "foo",
					Generator: &generatorThatFailsToValidate{
						err: errValidationFailed,
					},
				}),
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
		tt := tt

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

type parserThatFailsToValidate struct {
	err error
}

func (p parserThatFailsToValidate) Parse(_ *http.Request) (payload any, err error) {
	panic("not called in our tests")
}

func (p parserThatFailsToValidate) Validate(_ context.Context) error {
	return p.err
}

type generatorThatFailsToValidate struct {
	err error
}

func (g generatorThatFailsToValidate) Generate(_ any) ([]common.Notification, error) {
	panic("not called in our tests")
}

func (g generatorThatFailsToValidate) Validate(_ context.Context) error {
	return g.err
}

type transportThatFailsToValidate struct {
	err error
}

func (t transportThatFailsToValidate) Push(_ context.Context, _ common.Notification) error {
	panic("not called in our tests")
}

func (t transportThatFailsToValidate) ID() common.TransportID {
	return "test"
}

func (t transportThatFailsToValidate) Validate(_ context.Context) error {
	return t.err
}

type userStoreThatFailsToValidate struct {
	err error
}

func (s userStoreThatFailsToValidate) Get(_ identifier.Identifier) (*user.User, error) {
	panic("not called in our tests")
}

func (s userStoreThatFailsToValidate) Find(_ identifier.Collection) (*user.User, error) {
	panic("not called in our tests")
}

func (s userStoreThatFailsToValidate) Validate(_ context.Context) error {
	return s.err
}
