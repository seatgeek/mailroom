// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/notifier/preference"
	"github.com/stretchr/testify/assert"
)

func TestPreferencesHandler_HydrateUserPreferences(t *testing.T) {
	t.Parallel()

	h := createHandler(t)

	testCases := []struct {
		name              string
		storedPreferences preference.Map
		expected          preference.Map
	}{
		{
			"no stored preferences",
			nil,
			preference.Map{
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
			preference.Map{
				"com.gitlab.push": {
					"slack": true,
				},
				"com.argocd.sync-succeeded": {
					"slack": false,
				},
			},
			preference.Map{
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
			preference.Map{
				"com.gitlab.push": {
					"slack": true,
					"email": false,
				},
			},
			preference.Map{
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
			preference.Map{
				"com.mongodb.scale": {
					"slack": true,
					"email": false,
				},
			},
			preference.Map{
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
			preference.Map{
				"com.gitlab.push": {
					"slack":         true,
					"email":         false,
					"carrierpidgon": false,
					"telegraph":     true,
				},
			},
			preference.Map{
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

			hydrated := h.buildCurrentUserPreferences(t.Context(), tc.storedPreferences)

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

		user, err := handler.userStore.Get(t.Context(), "rufus")
		assert.NoError(t, err)

		expectedMap := preference.Map{
			"com.gitlab.push": {
				"slack": false,
				"email": false,
			},
			"com.argocd.sync-succeeded": {
				"slack": false,
				"email": true,
			},
		}
		assert.Equal(t, expectedMap, user.Preferences)
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
				"key": "argo",
				"event_types": [
					{
						"key": "com.argocd.sync-succeeded",
						"title": "Sync Succeeded",
						"description": "Emitted when an Argo CD sync operation completes successfully"
					}
				]
			},
			{
				"key": "gitlab",
				"event_types": [
					{
						"key": "com.gitlab.push",
						"title": "Push",
						"description": "Emitted when a user pushes code to a GitLab repository"
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

	srcGitlab := event.NewMockParser(t)
	srcGitlab.EXPECT().EventTypes().Maybe().Return([]event.TypeDescriptor{
		{
			Key:         "com.gitlab.push",
			Title:       "Push",
			Description: "Emitted when a user pushes code to a GitLab repository",
		},
	})

	srcArgo := event.NewMockParser(t)
	srcArgo.EXPECT().EventTypes().Maybe().Return([]event.TypeDescriptor{
		{
			Key:         "com.argocd.sync-succeeded",
			Title:       "Sync Succeeded",
			Description: "Emitted when an Argo CD sync operation completes successfully",
		},
	})

	parsers := map[string]event.Parser{
		"gitlab": srcGitlab,
		"argo":   srcArgo,
	}

	transports := []event.TransportKey{
		"slack",
		"email",
	}

	u := New("rufus", WithPreference("com.gitlab.push", "slack", false))
	userStore := NewInMemoryStore(u)

	return NewPreferencesHandler(userStore, parsers, transports, preference.Default(true))
}

func TestListOptions(t *testing.T) {
	t.Parallel()

	store := NewMockStore(t)
	// Use mock Parser
	parser1 := event.NewMockParser(t)
	parser1.EXPECT().EventTypes().Return([]event.TypeDescriptor{{
		Key:   "event1",
		Title: "Event 1",
	}})
	parser2 := event.NewMockParser(t)
	parser2.EXPECT().EventTypes().Return([]event.TypeDescriptor{{
		Key:   "event2",
		Title: "Event 2",
	}})

	parsers := map[string]event.Parser{
		"src1": parser1,
		"src2": parser2,
	}

	transports := []event.TransportKey{"t1", "t2"}

	// Update to use Parser
	ph := NewPreferencesHandler(store, parsers, transports, preference.Default(true))
	req := httptest.NewRequest("GET", "/configuration", nil)
	w := httptest.NewRecorder()

	ph.ListOptions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{
		"sources": [
			{
				"key": "src1",
				"event_types": [
					{"key": "event1", "title": "Event 1"}
				]
			},
			{
				"key": "src2",
				"event_types": [
					{"key": "event2", "title": "Event 2"}
				]
			}
		],
		"transports": [
			{"key": "t1"},
			{"key": "t2"}
		]
	}`, w.Body.String())
}
