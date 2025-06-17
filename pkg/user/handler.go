// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"cmp"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"slices"

	"github.com/gorilla/mux"
	"github.com/seatgeek/mailroom/pkg/event"
)

// PreferencesHandler exposes an HTTP API for managing user preferences
type PreferencesHandler struct {
	userStore  Store
	parsers    map[string]event.Parser
	transports []event.TransportKey
}

// NewPreferencesHandler creates a new PreferencesHandler for managing user preferences
func NewPreferencesHandler(userStore Store, parsers map[string]event.Parser, transports []event.TransportKey) *PreferencesHandler {
	return &PreferencesHandler{
		userStore:  userStore,
		parsers:    parsers,
		transports: transports,
	}
}

type preferencesBody struct {
	Preferences Preferences `json:"preferences"`
}

// GetPreferences returns the preferences for a given user
func (ph *PreferencesHandler) GetPreferences(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]

	u, err := ph.userStore.Get(request.Context(), key)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			slog.Info("user not found", "key", key)
			http.Error(writer, "user not found", http.StatusNotFound)
			return
		}

		slog.Error("failed to get user", "key", key, "error", err)
		http.Error(writer, "failed to get user", http.StatusInternalServerError)
		return
	}

	hydratedUserPreferences := ph.buildCurrentUserPreferences(u.Preferences)
	resp := preferencesBody{Preferences: hydratedUserPreferences}

	writeJson(writer, resp)
}

// UpdatePreferences updates the preferences for a given user
func (ph *PreferencesHandler) UpdatePreferences(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]

	var req preferencesBody
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		slog.Error("failed to decode request", "error", err)
		http.Error(writer, "failed to decode request", http.StatusBadRequest)
		return
	}

	err := ph.userStore.SetPreferences(request.Context(), key, req.Preferences)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			slog.Info("user not found", "key", key)
			http.Error(writer, "user not found", http.StatusNotFound)
			return
		}

		slog.Error("failed to save preferences", "key", key, "error", err)
		http.Error(writer, "failed to save preferences", http.StatusInternalServerError)
		return
	}

	writeJson(writer, preferencesBody{
		Preferences: ph.buildCurrentUserPreferences(req.Preferences),
	})
}

// Builds a current mapping of user preferences based on what is stored in the
// user store and the parsers and transports that are registered with the server.
//
// Only event types and transports that are currently active in the server will
// be included in the preference map. User is opted in to any preference that is
// not stored.
func (ph *PreferencesHandler) buildCurrentUserPreferences(p Preferences) Preferences {
	hydratedPreferences := make(Preferences)

	for _, src := range ph.parsers {
		for _, eventType := range src.EventTypes() {
			for _, transportKey := range ph.transports {
				if hydratedPreferences[eventType.Key] == nil {
					hydratedPreferences[eventType.Key] = make(map[event.TransportKey]bool)
				}
				hydratedPreferences[eventType.Key][transportKey] = p.Wants(eventType.Key, transportKey)
			}
		}
	}

	return hydratedPreferences
}

type transport struct {
	Key event.TransportKey `json:"key"`
}

type source struct {
	Key        string                 `json:"key"`
	EventTypes []event.TypeDescriptor `json:"event_types"`
}

type availableOptions struct {
	Sources    []source    `json:"sources"`
	Transports []transport `json:"transports"`
}

// ListOptions returns the available sources and transports for setting preferences
func (ph *PreferencesHandler) ListOptions(writer http.ResponseWriter, _ *http.Request) {
	sources := make([]source, 0, len(ph.parsers))
	for key, src := range ph.parsers {
		sources = append(sources, source{
			Key:        key,
			EventTypes: src.EventTypes(),
		})
	}

	// Sort sources by key (asc)
	slices.SortFunc(sources, func(a, b source) int {
		return cmp.Compare(a.Key, b.Key)
	})

	transports := make([]transport, len(ph.transports))
	for i, tp := range ph.transports {
		transports[i] = transport{
			Key: tp,
		}
	}

	resp := availableOptions{
		Sources:    sources,
		Transports: transports,
	}

	writeJson(writer, resp)
}

func writeJson(writer http.ResponseWriter, value any) {
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		slog.Error("failed to encode response", "error", err)
		writer.WriteHeader(500)
	}
}
