// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package mailroom

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/notifier"
	"github.com/seatgeek/mailroom/pkg/notifier/preference"
	"github.com/seatgeek/mailroom/pkg/server"
	"github.com/seatgeek/mailroom/pkg/user"
	"github.com/seatgeek/mailroom/pkg/validation"
)

// Server is the heart of the mailroom application
// It listens for incoming webhooks, parses them, generates notifications, and dispatches them to users.
type Server struct {
	listenAddr         string
	parsers            map[string]event.Parser
	processors         []event.Processor
	notifier           notifier.Notifier
	transports         []notifier.Transport
	defaultPreferences preference.Provider
	userStore          user.Store
	router             *mux.Router
}

type Opt func(s *Server)

// New returns a new server
func New(opts ...Opt) *Server {
	s := &Server{
		listenAddr:         "0.0.0.0:8000",
		router:             mux.NewRouter(),
		parsers:            make(map[string]event.Parser),
		defaultPreferences: preference.Default(true),
	}

	for _, opt := range opts {
		opt(s)
	}

	s.notifier = notifier.New(s.transports, preference.Chain{
		user.NewPreferenceProvider(s.userStore),
		s.defaultPreferences,
	})

	return s
}

// WithListenAddr sets the IP and port the server listens on, in the form "host:port"
func WithListenAddr(addr string) Opt {
	return func(s *Server) {
		s.listenAddr = addr
	}
}

// WithParser adds an event.Parser to the server with the given key.
// The key is used as the API endpoint for the server.
func WithParser(key string, parser event.Parser) Opt {
	return func(s *Server) {
		s.parsers[key] = parser
	}
}

// WithParserAndGenerator is a convenience function that adds an event.Parser and its corresponding processor (which generates notifications) in a single call.
func WithParserAndGenerator(key string, parser event.Parser, generator event.Processor) Opt {
	return func(s *Server) {
		s.parsers[key] = parser
		s.processors = append(s.processors, generator)
	}
}

// WithProcessors adds event.Processor instances to the server in the order given.
func WithProcessors(processors ...event.Processor) Opt {
	return func(s *Server) {
		s.processors = append(s.processors, processors...)
	}
}

// WithTransports adds notifier.Transport instances to the server
func WithTransports(transports ...notifier.Transport) Opt {
	return func(s *Server) {
		s.transports = append(s.transports, transports...)
	}
}

// WithUserStore sets the user.Store for the server
func WithUserStore(us user.Store) Opt {
	return func(s *Server) {
		s.userStore = us
	}
}

// WithDefaultPreferences sets the default preferences for the server
func WithDefaultPreferences(prefs preference.Provider) Opt {
	return func(s *Server) {
		s.defaultPreferences = prefs
	}
}

// WithRouter sets the mux.Router used for the server
func WithRouter(router *mux.Router) Opt {
	return func(s *Server) {
		s.router = router
	}
}

func (s *Server) validate(ctx context.Context) error { //nolint:revive // high cognitive complexity okay here
	for key, parser := range s.parsers {
		if v, ok := parser.(validation.Validator); ok {
			if err := v.Validate(ctx); err != nil {
				return fmt.Errorf("parser %s (%T) failed to validate: %w", key, parser, err)
			}
		}
	}

	for _, processor := range s.processors {
		if v, ok := processor.(validation.Validator); ok {
			if err := v.Validate(ctx); err != nil {
				return fmt.Errorf("processor %T failed to validate: %w", processor, err)
			}
		}
	}

	for _, t := range s.transports {
		if v, ok := t.(validation.Validator); ok {
			if err := v.Validate(ctx); err != nil {
				return fmt.Errorf("transport %s failed to validate: %w", t.Key(), err)
			}
		}
	}

	if v, ok := s.userStore.(validation.Validator); ok {
		if err := v.Validate(ctx); err != nil {
			return fmt.Errorf("user store failed to validate: %w", err)
		}
	}

	if v, ok := s.defaultPreferences.(validation.Validator); ok {
		if err := v.Validate(ctx); err != nil {
			return fmt.Errorf("default preferences failed to validate: %w", err)
		}
	}

	return nil
}

// Run starts the server in a Goroutine and blocks until the server is shut down.
// If the given context is canceled, the server will attempt to shut down gracefully.
func (s *Server) Run(ctx context.Context) error {
	if err := s.validate(ctx); err != nil {
		return fmt.Errorf("server validation failed: %w", err)
	}

	return s.serveHttp(ctx)
}

func (s *Server) serveHttp(ctx context.Context) error {
	hsm := s.router

	hsm.HandleFunc("/healthz", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(200)
		_, _ = writer.Write([]byte("^_^\n"))
	})

	// Mount all parsers
	for key, parser := range s.parsers {
		endpoint := "/event/" + key
		slog.Debug("mounting parser", "endpoint", endpoint)
		hsm.HandleFunc(endpoint, server.CreateEventProcessingHandler(key, parser, s.processors, s.notifier))
	}

	// Expose routes for managing user preferences
	prefs := user.NewPreferencesHandler(s.userStore, s.parsers, transportKeys(s.transports), s.defaultPreferences)
	hsm.HandleFunc("/users/{key}/preferences", prefs.GetPreferences).Methods("GET")
	hsm.HandleFunc("/users/{key}/preferences", prefs.UpdatePreferences).Methods("PUT")
	hsm.HandleFunc("/configuration", prefs.ListOptions).Methods("GET")

	hs := &http.Server{
		Addr:              s.listenAddr,
		Handler:           hsm,
		ReadHeaderTimeout: 2 * time.Second,
	}

	// Run the server in a Goroutine
	httpExited := make(chan error)
	go (func() {
		defer close(httpExited)

		slog.Info("http server listening on " + s.listenAddr)

		httpExited <- hs.ListenAndServe()
	})()

	select {
	// Wait for the context to be canceled
	case <-ctx.Done():
		slog.Info("shutting down http server gracefully")
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()

		if err := hs.Shutdown(shutdownCtx); err != nil { //nolint:contextcheck
			return fmt.Errorf("failed to gracefully shutdown http server: %w", err)
		}

		return nil
	// Or wait for the server to exit on its own (with some error)
	case err := <-httpExited:
		return err
	}
}

func transportKeys(transports []notifier.Transport) []event.TransportKey {
	keys := make([]event.TransportKey, len(transports))
	for i, t := range transports {
		keys[i] = t.Key()
	}
	return keys
}
