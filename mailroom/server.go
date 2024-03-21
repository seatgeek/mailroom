// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package mailroom

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

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
	sources    []*source.Source
	notifier   notifier.Notifier
	transports []notifier.Transport
	userStore  user.Store
}

type Opt func(s *Server)

// New returns a new server
func New(opts ...Opt) *Server {
	s := &Server{
		listenAddr: "0.0.0.0:8000",
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
func WithSources(sources ...*source.Source) Opt {
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

// Run starts the server
func (s *Server) Run(ctx context.Context) error {
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

func (s *Server) serveHttp(ctx context.Context) error {
	hsm := http.NewServeMux()

	hsm.HandleFunc("/healthz", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(200)
		_, _ = writer.Write([]byte("^_^\n"))
	})

	// Mount all sources wrapped with our error handler
	for _, src := range s.sources {
		endpoint := "/event/" + src.ID
		slog.Debug("mounting source", "endpoint", endpoint)
		hsm.HandleFunc(endpoint, server.HandleErr(server.CreateHandler(ctx, src, s.notifier)))
	}

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
