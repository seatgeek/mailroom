// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package webhooks

import (
	"errors"
	"net/http"

	"github.com/go-playground/webhooks/v6/azuredevops"
	"github.com/go-playground/webhooks/v6/bitbucket"
	bitbucketserver "github.com/go-playground/webhooks/v6/bitbucket-server"
	"github.com/go-playground/webhooks/v6/docker"
	"github.com/go-playground/webhooks/v6/gitea"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/go-playground/webhooks/v6/gitlab"
	"github.com/go-playground/webhooks/v6/gogs"
	"github.com/seatgeek/mailroom/mailroom/source"
)

var ErrInvalidPayload = errors.New("invalid payload")

type EventType interface{ ~string }

// hook is an interface that all webhooks implement upstream
type hook[Event EventType] interface {
	Parse(r *http.Request, events ...Event) (interface{}, error)
}

// Make sure those webhooks indeed implement the hook interface
var _ hook[azuredevops.Event] = &azuredevops.Webhook{}
var _ hook[bitbucket.Event] = &bitbucket.Webhook{}
var _ hook[bitbucketserver.Event] = &bitbucketserver.Webhook{}
var _ hook[docker.Event] = &docker.Webhook{}
var _ hook[gitea.Event] = &gitea.Webhook{}
var _ hook[github.Event] = &github.Webhook{}
var _ hook[gitlab.Event] = &gitlab.Webhook{}
var _ hook[gogs.Event] = &gogs.Webhook{}

// Adapter allows the use of webhooks as a source.PayloadParser
type Adapter[Event EventType] struct {
	hook   hook[Event]
	events []Event
}

func NewAdapter[Event EventType](hook hook[Event], events ...Event) *Adapter[Event] {
	adapter := &Adapter[Event]{
		hook:   hook,
		events: events,
	}

	return adapter
}

func (a Adapter[Event]) Parse(req *http.Request) (any, error) {
	payload, err := a.hook.Parse(req, a.events...)
	if err != nil {
		if isErrEventNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return payload, nil
}

var _ source.PayloadParser = &Adapter[string]{}

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

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
