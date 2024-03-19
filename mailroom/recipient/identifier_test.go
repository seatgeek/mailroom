// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package recipient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentifier_NamespaceAndKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		identifier Identifier
		want       string
	}{
		{
			name: "no namespace",
			identifier: Identifier{
				Kind: Email,
			},
			want: "email",
		},
		{
			name: "with namespace",
			identifier: Identifier{
				Namespace: "gitlab.com",
				Kind:      Username,
			},
			want: "gitlab.com/username",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.identifier.NamespaceAndKind()

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNewIdentifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		nsAndKind string
		value     interface{}
		want      Identifier
	}{
		{
			name:      "no namespace",
			nsAndKind: "email",
			value:     "codell@seatgeek.com",
			want: Identifier{
				Kind:  Email,
				Value: "codell@seatgeek.com",
			},
		},
		{
			name:      "with namespace",
			nsAndKind: "gitlab.com/username",
			value:     "codell",
			want: Identifier{
				Namespace: "gitlab.com",
				Kind:      Username,
				Value:     "codell",
			},
		},
		{
			name:      "int64 value",
			nsAndKind: "gitlab.com/id",
			value:     int64(123456),
			want: Identifier{
				Namespace: "gitlab.com",
				Kind:      ID,
				Value:     "123456",
			},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got Identifier

			switch v := tc.value.(type) {
			case string:
				got = NewIdentifier(tc.nsAndKind, v)
			case int64:
				got = NewIdentifier(tc.nsAndKind, v)
			}

			assert.Equal(t, tc.want, got)
		})
	}
}
