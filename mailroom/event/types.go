// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package event

import (
	"net/url"
	"time"
)

// Event is some action that occurred in an external system that we may want to send notifications for
type Event[T Payload] struct {
	Context
	Data T
}

type Payload = any

// Context contains the metadata for an event
// The fields are based on the CloudEvent spec: https://github.com/cloudevents/spec/blob/main/cloudevents/spec.md
type Context struct {
	ID      ID        // required
	Source  Source    // required
	Type    Type      // required
	Subject string    // optional
	Time    time.Time // optional
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
