// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/seatgeek/mailroom/pkg/common"
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/handler"
	"github.com/stretchr/testify/assert"
)

func TestPreferencesHandler_HydrateUserPreferences(t *testing.T) {
	t.Parallel()

	h := createHandler(t)

	testCases := []struct {
		name              string
		storedPreferences Preferences
		expected          Preferences
	}{
		{
			"no stored preferences",
			nil,
			Preferences{
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
			Preferences{
				"com.gitlab.push": {
					"slack": true,
				},
				"com.argocd.sync-succeeded": {
					"slack": false,
				},
			},
			Preferences{
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
			Preferences{
				"com.gitlab.push": {
					"slack": true,
					"email": false,
				},
			},
			Preferences{
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
			Preferences{
				"com.mongodb.scale": {
					"slack": true,
					"email": false,
				},
			},
			Preferences{
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
			Preferences{
				"com.gitlab.push": {
					"slack":         true,
					"email":         false,
					"carrierpidgon": false,
					"telegraph":     true,
				},
			},
			Preferences{
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

			hydrated := h.buildCurrentUserPreferences(tc.storedPreferences)

			assert.Equal(t, tc.expected, hydrated)
		})
	}
}

func TestPreferencesHandler_GetPreferences(t *testing.T) {
	t.Parallel()

	handler := createHandler(t)
	router := mux.NewRouter()
	router.HandleFunc("/users/{key}/preferences", handler.GetPreferences).Methods("GET")

	t.Run("Happy path", func(t *testing.T) {
		t.Parallel()

		writer := httptest.NewRecorder()
		router.ServeHTTP(writer, httptest.NewRequest("GET", "/users/rufus/preferences", nil))

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
		t.Parallel()

		writer := httptest.NewRecorder()
		router.ServeHTTP(writer, httptest.NewRequest("GET", "/users/taylor/preferences", nil))

		assert.Equal(t, 404, writer.Code)
	})
}

func TestPreferencesHandler_UpdatePreferences(t *testing.T) {
	t.Parallel()

	handler := createHandler(t)

	router := mux.NewRouter()
	router.HandleFunc("/users/{key}/preferences", handler.UpdatePreferences).Methods("PUT")

	t.Run("Happy path", func(t *testing.T) {
		t.Parallel()

		writer := httptest.NewRecorder()
		router.ServeHTTP(writer, httptest.NewRequest("PUT", "/users/rufus/preferences", bytes.NewBufferString(`{
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

		user, err := handler.userStore.Get("rufus")
		assert.NoError(t, err)

		assert.False(t, user.Wants("com.gitlab.push", "slack"))
		assert.False(t, user.Wants("com.gitlab.push", "email"))
		assert.False(t, user.Wants("com.argocd.sync-succeeded", "slack"))
		assert.True(t, user.Wants("com.argocd.sync-succeeded", "email"))
	})

	t.Run("User doesn't exist", func(t *testing.T) {
		t.Parallel()

		writer := httptest.NewRecorder()
		router.ServeHTTP(writer, httptest.NewRequest("PUT", "/users/taylor/preferences", bytes.NewBufferString(`{
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

	t.Run("Bad request body", func(t *testing.T) {
		t.Parallel()

		writer := httptest.NewRecorder()
		router.ServeHTTP(writer, httptest.NewRequest("PUT", "/users/rufus/preferences", bytes.NewBufferString(`{{{{ lol this isn't json!`)))

		assert.Equal(t, 400, writer.Code)
	})
}

func TestPreferencesHandler_ListOptions(t *testing.T) {
	t.Parallel()

	handler := createHandler(t)

	router := mux.NewRouter()
	router.HandleFunc("/configuration", handler.ListOptions).Methods("GET")

	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, httptest.NewRequest("GET", "/configuration", nil))

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

func createHandler(t *testing.T) *PreferencesHandler {
	t.Helper()

	srcGitlab := handler.NewMockHandler(t)
	srcGitlab.EXPECT().Key().Return("gitlab").Maybe()
	srcGitlab.EXPECT().EventTypes().Maybe().Return([]event.TypeDescriptor{
		{
			Key:         "com.gitlab.push",
			Title:       "Push",
			Description: "Emitted when a user pushes code to a GitLab repository",
		},
	})

	srcArgo := handler.NewMockHandler(t)
	srcArgo.EXPECT().Key().Return("argo").Maybe()
	srcArgo.EXPECT().EventTypes().Maybe().Return([]event.TypeDescriptor{
		{
			Key:         "com.argocd.sync-succeeded",
			Title:       "Sync Succeeded",
			Description: "Emitted when an Argo CD sync operation completes successfully",
		},
	})

	handlers := []handler.Handler{
		srcGitlab,
		srcArgo,
	}

	transports := []common.TransportKey{
		"slack",
		"email",
	}

	u := New("rufus", WithPreference("com.gitlab.push", "slack", false))
	userStore := NewInMemoryStore(u)

	return NewPreferencesHandler(userStore, handlers, transports)
}
