// Copyright 2024 SeatGeek, Inc.
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
	"github.com/seatgeek/mailroom/pkg/common"
	"github.com/seatgeek/mailroom/pkg/handler"
	"github.com/seatgeek/mailroom/pkg/notifier"
	"github.com/seatgeek/mailroom/pkg/server"
	"github.com/seatgeek/mailroom/pkg/user"
)

// interface for gorilla/mux router
type MuxRouter interface {
	http.Handler
	HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *mux.Route
}

// Server is the heart of the mailroom application
// It listens for incoming webhooks, parses them, generates notifications, and dispatches them to users.
type Server struct {
	listenAddr string
	handlers   []handler.Handler
	notifier   notifier.Notifier
	transports []notifier.Transport
	userStore  user.Store
	router     MuxRouter
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

// WithListenAddr sets the IP and port the server listens on, in the form "host:port"
func WithListenAddr(addr string) Opt {
	return func(s *Server) {
		s.listenAddr = addr
	}
}

// WithHandlers adds handler.Handler instances to the server
func WithHandlers(handlers ...handler.Handler) Opt {
	return func(s *Server) {
		s.handlers = append(s.handlers, handlers...)
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

// WithRouter sets the mux.Router used for the server
func WithRouter(router MuxRouter) Opt {
	return func(s *Server) {
		s.router = router
	}
}

func (s *Server) validate(ctx context.Context) error {
	for _, src := range s.handlers {
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

	// Mount all handlers
	for _, src := range s.handlers {
		endpoint := "/event/" + src.Key()
		slog.Debug("mounting handler", "endpoint", endpoint)
		hsm.HandleFunc(endpoint, server.CreateEventHandler(src, s.notifier))
	}

	// Expose routes for managing user preferences
	prefs := user.NewPreferencesHandler(s.userStore, s.handlers, transportKeys(s.transports))
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

func transportKeys(transports []notifier.Transport) []common.TransportKey {
	keys := make([]common.TransportKey, len(transports))
	for i, t := range transports {
		keys[i] = t.Key()
	}
	return keys
}
