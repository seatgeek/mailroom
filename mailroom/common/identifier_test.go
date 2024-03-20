// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNamespaceAndKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  NamespaceAndKind
	}{
		{
			input: "email",
			want: NamespaceAndKind{
				Namespace: "",
				Kind:      Email,
			},
		},
		{
			input: "gitlab.com/username",
			want: NamespaceAndKind{
				Namespace: "gitlab.com",
				Kind:      Username,
			},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got := parseNamespaceAndKind(tc.input)

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNamespaceAndKind_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		nsAndKind NamespaceAndKind
		want      string
	}{
		{
			name: "no namespace",
			nsAndKind: NamespaceAndKind{
				Kind: Email,
			},
			want: "email",
		},
		{
			name: "with namespace",
			nsAndKind: NamespaceAndKind{
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

			got := tc.nsAndKind.String()

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNewIdentifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		nsAndKind     string
		value         interface{}
		wantNamespace string
		wantKind      IdentifierKind
		wantValue     string
	}{
		{
			name:          "no namespace",
			nsAndKind:     "email",
			value:         "codell@seatgeek.com",
			wantNamespace: "",
			wantKind:      Email,
			wantValue:     "codell@seatgeek.com",
		},
		{
			name:          "with namespace",
			nsAndKind:     "gitlab.com/username",
			value:         "codell",
			wantNamespace: "gitlab.com",
			wantKind:      Username,
			wantValue:     "codell",
		},
		{
			name:          "int64 value",
			nsAndKind:     "gitlab.com/id",
			value:         int64(123456),
			wantNamespace: "gitlab.com",
			wantKind:      ID,
			wantValue:     "123456",
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

			assert.Equal(t, tc.wantNamespace, got.Namespace)
			assert.Equal(t, tc.wantKind, got.Kind)
			assert.Equal(t, tc.wantValue, got.Value)
		})
	}
}
