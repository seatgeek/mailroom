// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"testing"

	"github.com/seatgeek/mailroom/mailroom/common"

	"github.com/stretchr/testify/assert"
)

func TestUser_GetIdentifier(t *testing.T) {
	t.Parallel()

	email := common.NewIdentifier("email", "codell@seatgeek.com")
	gitlabUsername := common.NewIdentifier("gitlab.com/username", "codell")
	slackId := common.NewIdentifier("slack.com/id", "U1234567")
	slackUsername := common.NewIdentifier("slack.com/username", "colin.odell")

	user := New(email, gitlabUsername, slackId, slackUsername)

	tests := []struct {
		name       string
		query      common.NamespaceAndKind
		wantId     common.Identifier
		wantExists bool
	}{
		{
			name:       "any email",
			query:      common.NamespaceAndKind{Kind: common.Email},
			wantId:     email,
			wantExists: true,
		},
		{
			name:       "any username",
			query:      common.NamespaceAndKind{Kind: common.Username},
			wantId:     gitlabUsername,
			wantExists: true,
		},
		{
			name:       "specific username",
			query:      common.NamespaceAndKind{Namespace: "slack.com", Kind: common.Username},
			wantId:     slackUsername,
			wantExists: true,
		},
		{
			name:       "non-existent",
			query:      common.NamespaceAndKind{Kind: "foo"},
			wantExists: false,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotId, gotExists := user.GetIdentifier(tc.query)

			assert.Equal(t, tc.wantId, gotId)
			assert.Equal(t, tc.wantExists, gotExists)
		})
	}
}
