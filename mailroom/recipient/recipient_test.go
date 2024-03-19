// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package recipient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecipient_Get(t *testing.T) {
	t.Parallel()

	email := NewIdentifier("email", "codell@seatgeek.com")
	gitlabUsername := NewIdentifier("gitlab.com/username", "codell")
	slackId := NewIdentifier("slack.com/id", "U1234567")
	slackUsername := NewIdentifier("slack.com/username", "colin.odell")

	recipient := New(email, gitlabUsername, slackId, slackUsername)

	tests := []struct {
		name       string
		query      Identifier
		wantId     Identifier
		wantExists bool
	}{
		{
			name:       "any email",
			query:      Identifier{Kind: Email},
			wantId:     email,
			wantExists: true,
		},
		{
			name:       "any username",
			query:      Identifier{Kind: Username},
			wantId:     gitlabUsername,
			wantExists: true,
		},
		{
			name:       "specific username",
			query:      Identifier{Namespace: "slack.com", Kind: Username},
			wantId:     slackUsername,
			wantExists: true,
		},
		{
			name:       "specific value",
			query:      Identifier{Value: "U1234567"},
			wantId:     slackId,
			wantExists: true,
		},
		{
			name:       "non-existent",
			query:      Identifier{Kind: "foo"},
			wantExists: false,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotId, gotExists := recipient.Get(tc.query)

			assert.Equal(t, tc.wantId, gotId)
			assert.Equal(t, tc.wantExists, gotExists)
		})
	}
}
