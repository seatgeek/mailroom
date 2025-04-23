// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

// Package webhooks provides a way to use github.com/go-playground/webhooks/v6 as a handler.PayloadParser
package webhooks

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/go-playground/webhooks/v6/azuredevops"
	"github.com/go-playground/webhooks/v6/bitbucket"
	bitbucketserver "github.com/go-playground/webhooks/v6/bitbucket-server"
	"github.com/go-playground/webhooks/v6/docker"
	"github.com/go-playground/webhooks/v6/gitea"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/go-playground/webhooks/v6/gitlab"
	"github.com/go-playground/webhooks/v6/gogs"
	"github.com/google/uuid"
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/handler"
)

type EventType interface{ ~string }

// hook is an interface that all webhooks implement upstream
type hook[Event EventType] interface {
	Parse(r *http.Request, events ...Event) (any, error)
}

// Make sure those webhooks indeed implement the hook interface
var (
	_ hook[azuredevops.Event]     = &azuredevops.Webhook{}
	_ hook[bitbucket.Event]       = &bitbucket.Webhook{}
	_ hook[bitbucketserver.Event] = &bitbucketserver.Webhook{}
	_ hook[docker.Event]          = &docker.Webhook{}
	_ hook[gitea.Event]           = &gitea.Webhook{}
	_ hook[github.Event]          = &github.Webhook{}
	_ hook[gitlab.Event]          = &gitlab.Webhook{}
	_ hook[gogs.Event]            = &gogs.Webhook{}
)

// Adapter allows the use of webhooks as a handler.PayloadParser
type Adapter[Event EventType] struct {
	hook   hook[Event]
	events []Event
}

// NewAdapter returns a new Adapter allowing the webhook library's hooks to be used as a handler.PayloadParser
func NewAdapter[Event EventType](hook hook[Event], events ...Event) *Adapter[Event] {
	adapter := &Adapter[Event]{
		hook:   hook,
		events: events,
	}

	return adapter
}

func (a Adapter[Event]) Parse(req *http.Request) (*event.Event[any], error) {
	payload, err := a.hook.Parse(req, a.events...)
	if err != nil {
		if isErrEventNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	if payload == nil {
		return nil, nil
	}

	hookType := reflect.TypeOf(a.hook).Name()
	payloadType := reflect.TypeOf(payload).Name()

	return &event.Event[any]{
		Context: event.Context{
			ID:     event.ID(uuid.New().String()),
			Source: must(event.NewSource("/webhooks/" + hookType)),
			Type:   event.Type(payloadType),
		},
		Data: payload,
	}, nil
}

var _ handler.PayloadParser[any] = &Adapter[string]{}

// isErrEventNotFound checks if the error returned by webhooks means that the event was not on the allowlist
// It's kinda hacky, but it's the least-worst way I could think to do this.
func isErrEventNotFound(err error) bool {
	return errors.Is(err, bitbucket.ErrEventNotFound) ||
		errors.Is(err, bitbucketserver.ErrEventNotFound) ||
		errors.Is(err, gitea.ErrEventNotFound) ||
		errors.Is(err, github.ErrEventNotFound) ||
		errors.Is(err, gitlab.ErrEventNotFound) ||
		errors.Is(err, gogs.ErrEventNotFound)
}

func must[T any](t *T) T {
	if t == nil {
		panic("expected non-nil pointer")
	}
	return *t
}
