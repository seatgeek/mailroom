// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package mailroom

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/identifier"
	"github.com/seatgeek/mailroom/mailroom/notifier"
	"github.com/seatgeek/mailroom/mailroom/server"
	"github.com/seatgeek/mailroom/mailroom/source"
	"github.com/seatgeek/mailroom/mailroom/user"
	"github.com/stretchr/testify/assert"
)

type dummyGenerator struct {
	eventTypes []common.EventTypeDescriptor
}

func (d *dummyGenerator) Generate(payload any) ([]common.Notification, error) {
	return nil, nil
}

func (d *dummyGenerator) EventTypes() []common.EventTypeDescriptor {
	return d.eventTypes
}

func TestNew(t *testing.T) {
	s := New()

	assert.NotNil(t, s)
	assert.Equal(t, "0.0.0.0:8000", s.listenAddr)
}

func TestWithSources(t *testing.T) {
	src1 := &source.Source{Key: "foo"}
	src2 := &source.Source{Key: "bar"}
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
					Key: "foo",
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
					Key: "foo",
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

func (g generatorThatFailsToValidate) EventTypes() []common.EventTypeDescriptor {
	panic("not called in our tests")
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

func (s userStoreThatFailsToValidate) Get(_ string) (*user.User, error) {
	panic("not called in our tests")
}

func (s userStoreThatFailsToValidate) GetByIdentifier(identifier identifier.Identifier) (*user.User, error) {
	panic("not called in our tests")
}

func (s userStoreThatFailsToValidate) Find(_ identifier.Collection) (*user.User, error) {
	panic("not called in our tests")
}

func (s userStoreThatFailsToValidate) SetPreferences(_ string, _ user.Preferences) error {
	panic("not called in our tests")
}

func (s userStoreThatFailsToValidate) Validate(_ context.Context) error {
	return s.err
}

func mkServer(t *testing.T) *Server {
	t.Helper()

	srcGitlab := &source.Source{
		Key: "gitlab",
		Generator: &dummyGenerator{
			eventTypes: []common.EventTypeDescriptor{
				{
					Key:         "com.gitlab.push",
					Title:       "Push",
					Description: "Emitted when a user pushes code to a GitLab repository",
				},
			},
		},
	}
	srcArgo := &source.Source{
		Key: "argo",
		Generator: &dummyGenerator{
			eventTypes: []common.EventTypeDescriptor{
				{
					Key:         "com.argocd.sync-succeeded",
					Title:       "Sync Succeeded",
					Description: "Emitted when an Argo CD sync operation completes successfully",
				},
			},
		},
	}

	tpSlack := notifier.NewMockTransport(t)
	tpSlack.On("Key").Return(common.TransportKey("slack"))
	tpEmail := notifier.NewMockTransport(t)
	tpEmail.On("Key").Return(common.TransportKey("email"))

	u := user.New(
		"rufus",
		user.WithPreference("com.gitlab.push", "slack", false),
	)
	userStore := user.NewInMemoryStore(u)

	return New(WithSources(srcGitlab, srcArgo), WithTransports(tpSlack, tpEmail), WithUserStore(userStore))
}

func TestHydrateUserPreferences(t *testing.T) {
	t.Parallel()

	s := mkServer(t)

	testCases := []struct {
		name              string
		storedPreferences user.Preferences
		expected          user.Preferences
	}{
		{
			"no stored preferences",
			nil,
			user.Preferences{
				"com.gitlab.push": {
					"slack": true,
					"email": true,
				},
				"com.argocd.sync-succeeded": {
					"slack": true,
					"email": true,
				},
			},
		},
		{
			"only subset of transports stored",
			user.Preferences{
				"com.gitlab.push": {
					"slack": true,
				},
				"com.argocd.sync-succeeded": {
					"slack": false,
				},
			},
			user.Preferences{
				"com.gitlab.push": {
					"slack": true,
					"email": true,
				},
				"com.argocd.sync-succeeded": {
					"slack": false,
					"email": true,
				},
			},
		},
		{
			"only subset of event types stored",
			user.Preferences{
				"com.gitlab.push": {
					"slack": true,
					"email": false,
				},
			},
			user.Preferences{
				"com.gitlab.push": {
					"slack": true,
					"email": false,
				},
				"com.argocd.sync-succeeded": {
					"slack": true,
					"email": true,
				},
			},
		},
		{
			"unknown event type stored",
			user.Preferences{
				"com.mongodb.scale": {
					"slack": true,
					"email": false,
				},
			},
			user.Preferences{
				"com.gitlab.push": {
					"slack": true,
					"email": true,
				},
				"com.argocd.sync-succeeded": {
					"slack": true,
					"email": true,
				},
			},
		},
		{
			"unknown transport stored",
			user.Preferences{
				"com.gitlab.push": {
					"slack":         true,
					"email":         false,
					"carrierpidgon": false,
					"telegraph":     true,
				},
			},
			user.Preferences{
				"com.gitlab.push": {
					"slack": true,
					"email": false,
				},
				"com.argocd.sync-succeeded": {
					"slack": true,
					"email": true,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hydrated := s.buildCurrentUserPreferences(tc.storedPreferences)

			assert.Equal(t, tc.expected, hydrated)
		})
	}
}

func TestPreferences(t *testing.T) {
	t.Run("GET", func(t *testing.T) {
		s := mkServer(t)

		r := mux.NewRouter()
		r.HandleFunc("/users/{key}/preferences", server.HandleErr(s.handleGetPreferences)).Methods("GET")

		t.Run("Happy path", func(t *testing.T) {
			writer := httptest.NewRecorder()
			r.ServeHTTP(writer, httptest.NewRequest("GET", "/users/rufus/preferences", nil))

			assert.Equal(t, 200, writer.Code)
			assert.JSONEq(t, `{
				"preferences": {
					"com.gitlab.push": {
						"slack": false,
						"email": true
					},
					"com.argocd.sync-succeeded": {
						"slack": true,
						"email": true
					}
				}
			}`, writer.Body.String())
		})

		t.Run("User doesn't exist", func(t *testing.T) {
			writer := httptest.NewRecorder()
			r.ServeHTTP(writer, httptest.NewRequest("GET", "/users/taylor/preferences", nil))

			assert.Equal(t, 404, writer.Code)
		})
	})

	t.Run("PUT", func(t *testing.T) {
		s := mkServer(t)

		r := mux.NewRouter()
		r.HandleFunc("/users/{key}/preferences", server.HandleErr(s.handlePutPreferences)).Methods("PUT")

		t.Run("Happy path", func(t *testing.T) {
			writer := httptest.NewRecorder()

			r.ServeHTTP(writer, httptest.NewRequest("PUT", "/users/rufus/preferences", bytes.NewBufferString(`{
				"preferences": {
					"com.gitlab.push": {
						"slack": false,
						"email": false
					},
					"com.argocd.sync-succeeded": {
						"slack": false,
						"email": true
					}
				}
			}`)))

			assert.Equal(t, 200, writer.Code)
			assert.JSONEq(t, `{
				"preferences": {
					"com.gitlab.push": {
						"slack": false,
						"email": false
					},
					"com.argocd.sync-succeeded": {
						"slack": false,
						"email": true
					}
				}
			}`, writer.Body.String())

			user, err := s.userStore.Get("rufus")
			assert.Nil(t, err)

			assert.False(t, user.Wants("com.gitlab.push", "slack"))
			assert.False(t, user.Wants("com.gitlab.push", "email"))
			assert.False(t, user.Wants("com.argocd.sync-succeeded", "slack"))
			assert.True(t, user.Wants("com.argocd.sync-succeeded", "email"))
		})

		t.Run("User doesn't exist", func(t *testing.T) {
			writer := httptest.NewRecorder()
			r.ServeHTTP(writer, httptest.NewRequest("PUT", "/users/taylor/preferences", bytes.NewBufferString(`{
				"preferences": {
					"com.gitlab.push": {
						"slack": false,
						"email": false
					},
					"com.argocd.sync-succeeded": {
						"slack": false,
						"email": true
					}
				}
			}`)))

			assert.Equal(t, 404, writer.Code)
		})
	})
}

func TestConfiguration(t *testing.T) {
	s := mkServer(t)

	r := mux.NewRouter()
	r.HandleFunc("/configuration", server.HandleErr(s.handleGetConfiguration)).Methods("GET")

	writer := httptest.NewRecorder()
	r.ServeHTTP(writer, httptest.NewRequest("GET", "/configuration", nil))

	assert.Equal(t, 200, writer.Code)
	assert.JSONEq(t, `{
		"sources": [
			{
				"key": "gitlab",
				"event_types": [
					{
						"key": "com.gitlab.push",
						"title": "Push",
						"description": "Emitted when a user pushes code to a GitLab repository"
					}
				]
			},
			{
				"key": "argo",
				"event_types": [
					{
						"key": "com.argocd.sync-succeeded",
						"title": "Sync Succeeded",
						"description": "Emitted when an Argo CD sync operation completes successfully"
					}
				]
			}
		],
		"transports": [
			{
				"key": "slack"
			},
			{
				"key": "email"
			}
		]
	}`, writer.Body.String())
}
