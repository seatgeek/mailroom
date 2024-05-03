// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package event

// ID is a unique identifier for an event occurrence
// It should be a non-empty string that is unique within the context of the EventSource
// See https://github.com/cloudevents/spec/blob/main/cloudevents/spec.md#id
type ID string

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
