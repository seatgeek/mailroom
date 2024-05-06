// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package mailroom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/event"
	"github.com/seatgeek/mailroom/mailroom/notifier"
	"github.com/seatgeek/mailroom/mailroom/server"
	"github.com/seatgeek/mailroom/mailroom/source"
	"github.com/seatgeek/mailroom/mailroom/user"
	"gopkg.in/tomb.v2"
)

var ErrShutdown = errors.New("shutting down")

// Server is the heart of the mailroom application
// It listens for incoming webhooks, parses them, generates notifications, and dispatches them to users.
type Server struct {
	listenAddr string
	sources    []source.Source
	notifier   notifier.Notifier
	transports []notifier.Transport
	userStore  user.Store
	router     *mux.Router
}

type Opt func(s *Server)

// New returns a new server
func New(opts ...Opt) *Server {
	s := &Server{
		listenAddr: "0.0.0.0:8000",
		router:     mux.NewRouter(),
	}

	for _, opt := range opts {
		opt(s)
	}

	s.notifier = notifier.New(s.userStore, s.transports...)

	return s
}

// WithListenAddr sets the IP and port the server listens on
func WithListenAddr(addr string) Opt {
	return func(s *Server) {
		s.listenAddr = addr
	}
}

// WithSources adds sources to the server
func WithSources(sources ...source.Source) Opt {
	return func(s *Server) {
		s.sources = append(s.sources, sources...)
	}
}

// WithTransports adds named transports to the server
func WithTransports(transports ...notifier.Transport) Opt {
	return func(s *Server) {
		s.transports = append(s.transports, transports...)
	}
}

// WithUserStore sets the user store for the server
func WithUserStore(us user.Store) Opt {
	return func(s *Server) {
		s.userStore = us
	}
}

// WithRouter sets the function used to create a router for the server
func WithRouter(router *mux.Router) Opt {
	return func(s *Server) {
		s.router = router
	}
}

func (s *Server) validate(ctx context.Context) error {
	for _, src := range s.sources {
		if v, ok := src.(common.Validator); ok {
			if err := v.Validate(ctx); err != nil {
				return fmt.Errorf("parser %s failed to validate: %w", src.Key(), err)
			}
		}
	}

	for _, t := range s.transports {
		if v, ok := t.(common.Validator); ok {
			if err := v.Validate(ctx); err != nil {
				return fmt.Errorf("transport %s failed to validate: %w", t.Key(), err)
			}
		}
	}

	if v, ok := s.userStore.(common.Validator); ok {
		if err := v.Validate(ctx); err != nil {
			return fmt.Errorf("user store failed to validate: %w", err)
		}
	}

	return nil
}

// Run starts the server
func (s *Server) Run(ctx context.Context) error {
	if err := s.validate(ctx); err != nil {
		return fmt.Errorf("server validation failed: %w", err)
	}

	httpTomb, httpTombCtx := tomb.WithContext(ctx)
	defer httpTomb.Kill(ErrShutdown)
	httpTomb.Go(func() error { return s.serveHttp(httpTombCtx) })

	for {
		select {
		case <-httpTomb.Dead():
			slog.Warn("httpTomb died")
			return httpTomb.Err()
		case <-ctx.Done():
			slog.Warn("shutting down due to user signal, byee")
			return nil
		}
	}
}

// Builds a current mapping of user preferences based on what is stored in the
// user store and the sources and transports that are registered with the server.
//
// Only event types and transports that are currently active in the server will
// be included in the preference map. User is opted in to any preference that is
// not stored.
func (s *Server) buildCurrentUserPreferences(p user.Preferences) user.Preferences {
	hydratedPreferences := make(user.Preferences)

	for _, src := range s.sources {
		for _, eventType := range src.EventTypes() {
			for _, transport := range s.transports {
				if hydratedPreferences[eventType.Key] == nil {
					hydratedPreferences[eventType.Key] = make(map[common.TransportKey]bool)
				}
				hydratedPreferences[eventType.Key][transport.Key()] = p.Wants(eventType.Key, transport.Key())
			}
		}
	}

	return hydratedPreferences
}

type PreferencesBody struct {
	Preferences user.Preferences `json:"preferences"`
}

func (s *Server) handleGetPreferences(writer http.ResponseWriter, request *http.Request) error {
	vars := mux.Vars(request)
	key := vars["key"]

	u, err := s.userStore.Get(key)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			slog.Info("user not found", "key", key)
			return &server.Error{Code: http.StatusNotFound, Reason: err}
		}

		return err
	}

	hydratedUserPreferences := s.buildCurrentUserPreferences(u.Preferences)
	resp := PreferencesBody{Preferences: hydratedUserPreferences}

	if err := json.NewEncoder(writer).Encode(resp); err != nil {
		return err
	}

	return nil
}

func (s *Server) handlePutPreferences(writer http.ResponseWriter, request *http.Request) error {
	vars := mux.Vars(request)
	key := vars["key"]

	var req PreferencesBody
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		return err
	}

	err := s.userStore.SetPreferences(key, req.Preferences)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			slog.Info("user not found", "key", key)
			return &server.Error{Code: http.StatusNotFound, Reason: err}
		}

		return err
	}

	resp := PreferencesBody{Preferences: s.buildCurrentUserPreferences(req.Preferences)}
	if err := json.NewEncoder(writer).Encode(resp); err != nil {
		return err
	}

	return nil
}

type APITransport struct {
	Key common.TransportKey `json:"key"`
}

type APISource struct {
	Key        string                 `json:"key"`
	EventTypes []event.TypeDescriptor `json:"event_types"`
}

type APIConfiguration struct {
	Sources    []APISource    `json:"sources"`
	Transports []APITransport `json:"transports"`
}

func (s *Server) handleGetConfiguration(writer http.ResponseWriter, _ *http.Request) error {
	sources := make([]APISource, len(s.sources))
	for i, src := range s.sources {
		src := APISource{
			Key:        src.Key(),
			EventTypes: src.EventTypes(),
		}
		sources[i] = src
	}

	transports := make([]APITransport, len(s.transports))
	for i, transport := range s.transports {
		tp := APITransport{
			Key: transport.Key(),
		}
		transports[i] = tp
	}

	resp := APIConfiguration{
		Sources:    sources,
		Transports: transports,
	}

	if err := json.NewEncoder(writer).Encode(resp); err != nil {
		return err
	}

	return nil
}

func (s *Server) serveHttp(ctx context.Context) error {
	hsm := s.router

	hsm.HandleFunc("/healthz", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(200)
		_, _ = writer.Write([]byte("^_^\n"))
	})

	// Mount all sources wrapped with our error handler
	for _, src := range s.sources {
		endpoint := "/event/" + src.Key()
		slog.Debug("mounting source", "endpoint", endpoint)
		hsm.HandleFunc(endpoint, server.HandleErr(server.CreateEventHandler(ctx, src, s.notifier)))
	}

	hsm.HandleFunc("/users/{key}/preferences", server.HandleErr(s.handleGetPreferences)).Methods("GET")
	hsm.HandleFunc("/users/{key}/preferences", server.HandleErr(s.handlePutPreferences)).Methods("PUT")
	hsm.HandleFunc("/configuration", server.HandleErr(s.handleGetConfiguration)).Methods("GET")

	hs := &http.Server{
		Addr:              s.listenAddr,
		Handler:           hsm,
		ReadHeaderTimeout: 2 * time.Second,
	}

	httpExited := make(chan error)
	go (func() {
		defer close(httpExited)

		slog.Info("http server listening on " + s.listenAddr)

		httpExited <- hs.ListenAndServe()
	})()

	select {
	case <-ctx.Done():
		return ErrShutdown
	case err := <-httpExited:
		return fmt.Errorf("internal server exited: %w", err)
	}
}
