// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"slices"

	"github.com/gorilla/mux"
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/notification"
	"github.com/seatgeek/mailroom/pkg/notifier/preference"
)

// PreferencesHandler exposes an HTTP API for managing user preferences
type PreferencesHandler struct {
	userStore  Store
	parsers    map[string]event.Parser
	transports []event.TransportKey
	defaults   preference.Provider
}

// NewPreferencesHandler creates a new PreferencesHandler for managing user preferences
func NewPreferencesHandler(userStore Store, parsers map[string]event.Parser, transports []event.TransportKey, defaults preference.Provider) *PreferencesHandler {
	return &PreferencesHandler{
		userStore:  userStore,
		parsers:    parsers,
		transports: transports,
		defaults:   defaults,
	}
}

type preferencesBody struct {
	Preferences preference.Map `json:"preferences"`
}

// GetPreferences returns the preferences for a given user
func (ph *PreferencesHandler) GetPreferences(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]

	u, err := ph.userStore.Get(request.Context(), key)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			slog.InfoContext(request.Context(), "user not found", "key", key)
			http.Error(writer, "user not found", http.StatusNotFound)
			return
		}

		slog.ErrorContext(request.Context(), "failed to get user", "key", key, "error", err)
		http.Error(writer, "failed to get user", http.StatusInternalServerError)
		return
	}

	hydratedUserPreferences := ph.buildCurrentUserPreferences(request.Context(), preference.Chain{u.Preferences, ph.defaults})
	resp := preferencesBody{Preferences: hydratedUserPreferences}

	writeJson(request.Context(), writer, resp)
}

// UpdatePreferences updates the preferences for a given user
func (ph *PreferencesHandler) UpdatePreferences(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]

	var req preferencesBody
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		slog.ErrorContext(request.Context(), "failed to decode request", "error", err)
		http.Error(writer, "failed to decode request", http.StatusBadRequest)
		return
	}

	err := ph.userStore.SetPreferences(request.Context(), key, req.Preferences)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			slog.InfoContext(request.Context(), "user not found", "key", key)
			http.Error(writer, "user not found", http.StatusNotFound)
			return
		}

		slog.ErrorContext(request.Context(), "failed to save preferences", "key", key, "error", err)
		http.Error(writer, "failed to save preferences", http.StatusInternalServerError)
		return
	}

	writeJson(request.Context(), writer, preferencesBody{
		Preferences: ph.buildCurrentUserPreferences(request.Context(), preference.Chain{req.Preferences, ph.defaults}),
	})
}

// Builds a current mapping of user preferences based on what is stored in the
// user store and the parsers and transports that are registered with the server.
//
// Only event types and transports that are currently active in the server will
// be included in the preference map. User is opted in to any preference that is
// not stored.
func (ph *PreferencesHandler) buildCurrentUserPreferences(ctx context.Context, p preference.Provider) preference.Map {
	hydratedPreferences := make(preference.Map)

	for _, src := range ph.parsers {
		for _, eventType := range src.EventTypes() {
			n := fakeNotificationFor(eventType.Key)
			for _, transportKey := range ph.transports {
				if hydratedPreferences[eventType.Key] == nil {
					hydratedPreferences[eventType.Key] = make(map[event.TransportKey]bool)
				}
				if wants := p.Wants(ctx, n, transportKey); wants != nil {
					hydratedPreferences[eventType.Key][transportKey] = *wants
				} else {
					// No preference or fallback was available, so show it as enabled by default.
					hydratedPreferences[eventType.Key][transportKey] = true
				}
			}
		}
	}

	return hydratedPreferences
}

// fakeNotificationFor creates a fake notification for the given event type.
// This is needed because preferences are based on notifications and their context,
// so we need to simulate such a notification to check preferences against it.
// Most preferences are usually based on just the event type, and will usually return
// nil to fall back to some default later in the chain - this is why we can get away
// with using a fake notification here.
//
// But if you're reading this and thinking "this dirty hack doesn't work for my needs"
// then please open an issue to explain your use case so we can improve this!
func fakeNotificationFor(eventType event.Type) event.Notification {
	return notification.NewBuilder(event.Context{Type: eventType}).Build()
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
func (ph *PreferencesHandler) ListOptions(writer http.ResponseWriter, request *http.Request) {
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

	writeJson(request.Context(), writer, resp)
}

func writeJson(ctx context.Context, writer http.ResponseWriter, value any) {
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		slog.ErrorContext(ctx, "failed to encode response", "error", err)
		writer.WriteHeader(500)
	}
}
