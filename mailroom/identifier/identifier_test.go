// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package identifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
				Kind: KindEmail,
			},
			want: "email",
		},
		{
			name: "with namespace",
			nsAndKind: NamespaceAndKind{
				Namespace: "gitlab.com",
				Kind:      KindUsername,
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

func TestFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  NamespaceAndKind
	}{
		{
			input: "email",
			want: NamespaceAndKind{
				Namespace: "",
				Kind:      KindEmail,
			},
		},
		{
			input: "gitlab.com/username",
			want: NamespaceAndKind{
				Namespace: "gitlab.com",
				Kind:      KindUsername,
			},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got := For(tc.input)

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("no namespace", func(t *testing.T) {
		t.Parallel()

		got := New("email", "codell@seatgeek.com")

		assert.Equal(t, "", got.Namespace)
		assert.Equal(t, KindEmail, got.Kind)
		assert.Equal(t, "codell@seatgeek.com", got.Value)
	})

	t.Run("with namespace", func(t *testing.T) {
		t.Parallel()

		got := New("gitlab.com/username", "codell")

		assert.Equal(t, "gitlab.com", got.Namespace)
		assert.Equal(t, KindUsername, got.Kind)
		assert.Equal(t, "codell", got.Value)
	})

	t.Run("int64 value", func(t *testing.T) {
		t.Parallel()

		got := New("gitlab.com/id", int64(123456))

		assert.Equal(t, "gitlab.com", got.Namespace)
		assert.Equal(t, ID, got.Kind)
		assert.Equal(t, "123456", got.Value)
	})

	t.Run("namespace and kind already a struct", func(t *testing.T) {
		t.Parallel()

		got := New(
			NamespaceAndKind{
				Namespace: "slack.com",
				Kind:      "id",
			},
			"U1234567",
		)

		assert.Equal(t, "slack.com", got.Namespace)
		assert.Equal(t, ID, got.Kind)
		assert.Equal(t, "U1234567", got.Value)
	})
}

func TestCollection_Get(t *testing.T) {
	t.Parallel()

	email := New("email", "codell@seatgeek.com")
	gitlabUsername := New("gitlab.com/username", "codell")
	slackId := New("slack.com/id", "U1234567")
	slackUsername := New("slack.com/username", "colin.odell")

	identifiers := NewCollection(email, gitlabUsername, slackId, slackUsername)

	tests := []struct {
		name       string
		query      NamespaceAndKind
		wantOneOf  []Identifier
		wantExists bool
	}{
		{
			name:       "any email",
			query:      NamespaceAndKind{Kind: "email"},
			wantOneOf:  []Identifier{email},
			wantExists: true,
		},
		{
			name:       "any username",
			query:      NamespaceAndKind{Kind: "username"},
			wantOneOf:  []Identifier{gitlabUsername, slackUsername},
			wantExists: true,
		},
		{
			name:       "specific username",
			query:      NamespaceAndKind{Namespace: "slack.com", Kind: KindUsername},
			wantOneOf:  []Identifier{slackUsername},
			wantExists: true,
		},
		{
			name:       "non-existent",
			query:      NamespaceAndKind{Kind: "foo"},
			wantOneOf:  nil,
			wantExists: false,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, exists := identifiers.Get(tc.query)

			if len(tc.wantOneOf) == 0 {
				assert.Empty(t, got)
			} else {
				assert.Contains(t, tc.wantOneOf, got)
			}

			assert.Equal(t, tc.wantExists, exists)
		})
	}
}

func TestCollection_ToList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		collection Collection
		want       []Identifier
	}{
		{
			name:       "empty",
			collection: Collection{},
			want:       nil,
		},
		{
			name: "non-empty",
			collection: NewCollection(
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			),
			want: []Identifier{
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.collection.ToList()

			assert.ElementsMatch(t, tc.want, got)
		})
	}
}

func TestNewCollection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []Identifier
		want Collection
	}{
		{
			name: "empty",
			args: []Identifier{},
			want: Collection{},
		},
		{
			name: "non-empty",
			args: []Identifier{
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			},
			want: Collection{
				NamespaceAndKind{Kind: KindUsername}: "rufus",
				NamespaceAndKind{Kind: KindEmail}:    "rufus@seatgeek.com",
			},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewCollection(tc.args...)

			assert.Equal(t, tc.want, got)
		})
	}
}
