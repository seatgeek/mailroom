// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package recipient

// Recipient is somebody that should receive a notification
//
// A recipient may be known by different IDs in different systems. The ID we receive in the webhook payload
// (e.g. a GitLab ID or email address) probably won't contain the ID we need to send them a notification
// (e.g. a Slack ID). So instead of just passing a scalar ID around, we need some way to relate that original ID
// with the system it came from, so we can later translate it to the correct ID for the transport we're using.
//
// Additionally, we want to avoid performing that lookup when we create the Notification object as we don't yet know
// the user's preferences for receiving notifications; it would be wasteful to look up their Slack ID if they've turned
// those notifications off.
type Recipient struct {
	identifiers []Identifier
}

func New(identifiers ...Identifier) *Recipient {
	return &Recipient{
		identifiers: identifiers,
	}
}

func (r *Recipient) Add(identifier Identifier) {
	r.identifiers = append(r.identifiers, identifier)
}

// Get returns the first Identifier that matches the given query
// For example, to find any email address, pass an Identifier with Kind="email".
// Any query field that is empty will act as a wildcard.
func (r *Recipient) Get(query Identifier) (Identifier, bool) {
	for _, id := range r.identifiers {
		if query.Namespace != "" && query.Namespace != id.Namespace {
			continue
		}

		if query.Kind != "" && query.Kind != id.Kind {
			continue
		}

		if query.Value != "" && query.Value != id.Value {
			continue
		}

		return id, true
	}

	return Identifier{}, false
}
