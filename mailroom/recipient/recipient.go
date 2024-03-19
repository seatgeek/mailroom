// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package recipient

type IdentifierType string

const Email IdentifierType = "email"

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
	identifiers map[IdentifierType]string // eg. {"gitlab": "12345"; "email": "alice@seatgeek.com", "slack": "U123"}
}

func New(idType IdentifierType, identifier string) *Recipient {
	return &Recipient{
		identifiers: map[IdentifierType]string{
			idType: identifier,
		},
	}
}

func (r *Recipient) Add(idType IdentifierType, identifier string) {
	r.identifiers[idType] = identifier
}

func (r *Recipient) Get(idType IdentifierType) (string, bool) {
	id, ok := r.identifiers[idType]
	return id, ok
}
