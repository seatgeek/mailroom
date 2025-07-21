// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

// Package event provides types and functions for working with incoming events
package event

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/seatgeek/mailroom/pkg/identifier"
)

// Event is some action that occurred in an external system that we may want to send a Notification for
type Event struct {
	Context
	Data Payload
}

type Payload = any

// Parser is anything capable of parsing an incoming HTTP request into a canonical Event object.
type Parser interface {
	// Parse handles incoming webhooks, verifying them and returning a parsed Event (or an error)
	Parse(req *http.Request) (*Event, error)
	// EventTypes returns descriptors for all EventTypes that the parser may produce
	EventTypes() []TypeDescriptor
}

// Context contains the metadata for an event
// The fields are based on the CloudEvent spec: https://github.com/cloudevents/spec/blob/main/cloudevents/spec.md
type Context struct {
	ID      ID                // required
	Source  Source            // required
	Type    Type              // required
	Subject string            // optional
	Time    time.Time         // optional
	Labels  map[string]string // optional
}

// WithID returns a copy of the Context with the ID field set to the provided value
func (c Context) WithID(newID ID) Context {
	c.ID = newID
	return c
}

// WithSource returns a copy of the Context with the Source field set to the provided value
func (c Context) WithSource(newSource Source) Context {
	c.Source = newSource
	return c
}

// WithType returns a copy of the Context with the Type field set to the provided value
func (c Context) WithType(newType Type) Context {
	c.Type = newType
	return c
}

// WithSubject returns a copy of the Context with the Subject field set to the provided value
func (c Context) WithSubject(newSubject string) Context {
	c.Subject = newSubject
	return c
}

// WithTime returns a copy of the Context with the Time field set to the provided value
func (c Context) WithTime(newTime time.Time) Context {
	c.Time = newTime
	return c
}

// WithLabels returns a copy of the Context with the Labels field set to the provided value
func (c Context) WithLabels(newLabels map[string]string) Context {
	c.Labels = newLabels
	return c
}

// ID is a unique identifier for an event occurrence
// It should be a non-empty string that is unique within the context of the EventSource
// See https://github.com/cloudevents/spec/blob/main/cloudevents/spec.md#id
type ID string

// Source identifies the context in which an event happened
// It should be a non-empty URI reference (preferably an absolute URI)
// See https://github.com/cloudevents/spec/blob/main/cloudevents/spec.md#source-1
type Source struct {
	uri url.URL
}

// NewSource creates a new EventSource from a URI string
// If the URI is invalid, it will return nil
func NewSource(uri string) *Source {
	if uri == "" {
		return nil
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		return nil
	}

	return &Source{
		uri: *parsed,
	}
}

// MustSource creates a new EventSource from a URI string, panicking if the URI is invalid
func MustSource(uri string) Source {
	s := NewSource(uri)
	if s == nil {
		panic("invalid source URI")
	}

	return *s
}

func (s *Source) String() string {
	if s == nil {
		return ""
	}

	return s.uri.String()
}

// Type describes the type of event related to the originating occurrence.
// It may be used for routing, observability, etc. It must comply with CloudEvent `type` spec:
// https://github.com/cloudevents/spec/blob/main/cloudevents/spec.md#type
// Basically, it should be a non-empty string containing a reverse-DNS name.
// For example: "com.gitlab.push"
type Type string

// TypeDescriptor describes an event type in user-friendly terms
type TypeDescriptor struct {
	Key Type `json:"key"`
	// Title should be a human-readable title that describes the event, independent of the source.
	// So the title for "com.gitlab.merge_request.approved" could be "Merge Request Approved".
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// Notification is a notification that should be sent
type Notification interface {
	// Context provides the metadata for the notification
	Context() Context
	// Recipient returns the intended recipient of the notification
	Recipient() identifier.Set
	// Render returns the message to be sent via the given transport
	Render(TransportKey) string
	// WithRecipient returns a new notification with the specified recipient
	WithRecipient(identifier.Set) Notification
	// Clone returns a deep copy of the notification
	Clone() Notification
}

// TransportKey is a type that identifies a specific type of transport for sending notifications
type TransportKey string // eg. "slack"; "email"

// Processor creates or modifies a list of notifications for a given event.
type Processor interface {
	// Process takes an event and a slice of notifications (from previous processors or an empty slice
	// for the first processor) and returns a potentially modified slice of notifications.
	Process(ctx context.Context, evt Event, notifications []Notification) ([]Notification, error)
}
