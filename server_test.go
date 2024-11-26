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
	"github.com/seatgeek/mailroom/pkg/common"
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/handler"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/seatgeek/mailroom/pkg/notifier"
	"github.com/seatgeek/mailroom/pkg/server"
	"github.com/seatgeek/mailroom/pkg/user"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	s := New()

	assert.NotNil(t, s)
	assert.Equal(t, "0.0.0.0:8000", s.listenAddr)
}

func TestWithHandlers(t *testing.T) {
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

var _ handler.Handler = handlerThatFailsToValidate{}
var _ common.Validator = handlerThatFailsToValidate{}

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

func (s userStoreThatFailsToValidate) Get(_ string) (*user.User, error) {
	panic("not called in our tests")
}

func (s userStoreThatFailsToValidate) GetByIdentifier(identifier identifier.Identifier) (*user.User, error) {
	panic("not called in our tests")
}

func (s userStoreThatFailsToValidate) Find(_ identifier.Set) (*user.User, error) {
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

	srcGitlab := handler.NewMockHandler(t)
	srcGitlab.EXPECT().Key().Return("gitlab").Maybe()
	srcGitlab.EXPECT().EventTypes().Return([]event.TypeDescriptor{
		{
			Key:         "com.gitlab.push",
			Title:       "Push",
			Description: "Emitted when a user pushes code to a GitLab repository",
		},
	})

	srcArgo := handler.NewMockHandler(t)
	srcArgo.EXPECT().Key().Return("argo").Maybe()
	srcArgo.EXPECT().EventTypes().Return([]event.TypeDescriptor{
		{
			Key:         "com.argocd.sync-succeeded",
			Title:       "Sync Succeeded",
			Description: "Emitted when an Argo CD sync operation completes successfully",
		},
	})

	tpSlack := notifier.NewMockTransport(t)
	tpSlack.EXPECT().Key().Return("slack")
	tpEmail := notifier.NewMockTransport(t)
	tpEmail.EXPECT().Key().Return("email")

	u := user.New(
		"rufus",
		user.WithPreference("com.gitlab.push", "slack", false),
	)
	userStore := user.NewInMemoryStore(u)

	return New(WithHandlers(srcGitlab, srcArgo), WithTransports(tpSlack, tpEmail), WithUserStore(userStore))
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
			assert.NoError(t, err)

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
