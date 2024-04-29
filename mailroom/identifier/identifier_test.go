// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package identifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamespaceAndKind_Split(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input         NamespaceAndKind
		wantNamespace string
		wantKind      string
	}{
		{
			input:         NamespaceAndKind("email"),
			wantNamespace: "",
			wantKind:      "email",
		},
		{
			input:         NamespaceAndKind("gitlab.com/username"),
			wantNamespace: "gitlab.com",
			wantKind:      "username",
		},
	}

	for _, tc := range tests {
		t.Run(string(tc.input), func(t *testing.T) {
			t.Parallel()

			namespace, kind := tc.input.Split()

			assert.Equal(t, tc.wantNamespace, namespace)
			assert.Equal(t, tc.wantKind, kind)

			// We'll also test the Namespace() and Kind() methods here since they are simple wrappers around Split()
			assert.Equal(t, tc.wantNamespace, tc.input.Namespace())
			assert.Equal(t, Kind(tc.wantKind), tc.input.Kind())
		})
	}
}

func TestNewNamespaceAndKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		namespace string
		kind      any
		want      NamespaceAndKind
	}{
		{
			kind: KindEmail,
			want: "email",
		},
		{
			namespace: "gitlab.com",
			kind:      "username",
			want:      "gitlab.com/username",
		},
	}

	for _, tc := range tests {
		t.Run(string(tc.want), func(t *testing.T) {
			t.Parallel()

			var got NamespaceAndKind

			switch kind := tc.kind.(type) {
			case Kind:
				got = NewNamespaceAndKind(tc.namespace, kind)
			case string:
				got = NewNamespaceAndKind(tc.namespace, kind)
			}

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("no namespace", func(t *testing.T) {
		t.Parallel()

		got := New("email", "codell@seatgeek.com")

		assert.Equal(t, NamespaceAndKind("email"), got.NamespaceAndKind)
		assert.Equal(t, "codell@seatgeek.com", got.Value)
	})

	t.Run("with namespace", func(t *testing.T) {
		t.Parallel()

		got := New("gitlab.com/username", "codell")

		assert.Equal(t, NamespaceAndKind("gitlab.com/username"), got.NamespaceAndKind)
		assert.Equal(t, "codell", got.Value)
	})

	t.Run("int64 value", func(t *testing.T) {
		t.Parallel()

		got := New("gitlab.com/id", int64(123456))

		assert.Equal(t, NamespaceAndKind("gitlab.com/id"), got.NamespaceAndKind)
		assert.Equal(t, "123456", got.Value)
	})

	t.Run("namespace and kind already a NamespaceAndKind type", func(t *testing.T) {
		t.Parallel()

		nsAndKind := NamespaceAndKind("slack.com/id")

		got := New(
			nsAndKind,
			"U1234567",
		)

		assert.Equal(t, nsAndKind, got.NamespaceAndKind)
		assert.Equal(t, "U1234567", got.Value)
	})
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
			collection: NewCollection(),
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
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.collection.ToList()

			assert.ElementsMatch(t, tc.want, got)
		})
	}
}

func TestCollection_Add(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		original Collection
		add      Identifier
		want     Collection
	}{
		{
			name: "adds",
			original: NewCollection(
				New("username", "rufus"),
			),
			add: New("email", "rufus@seatgeek.com"),
			want: NewCollection(
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			),
		},
		{
			name: "overwrites",
			original: NewCollection(
				New("username", "rufus"),
				New("email", "rufus@example.com"),
			),
			add: New("email", "rufus@seatgeek.com"),
			want: NewCollection(
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			),
		},
		{
			name:     "empty original",
			original: NewCollection(),
			add:      New("username", "rufus"),
			want:     NewCollection(New("username", "rufus")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.original.Add(tc.add)

			assert.Equal(t, tc.want, tc.original)
		})
	}
}

func TestCollection_Merge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		original Collection
		merge    Collection
		want     Collection
	}{
		{
			name: "adds and overwrites",
			original: NewCollection(
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			),
			merge: NewCollection(
				New("id", "123"),
				New("email", "rufus@example.com"),
			),
			want: NewCollection(
				New("username", "rufus"),
				New("email", "rufus@example.com"),
				New("id", "123"),
			),
		},
		{
			name:     "empty original",
			original: NewCollection(),
			merge:    NewCollection(New("username", "rufus")),
			want:     NewCollection(New("username", "rufus")),
		},
		{
			name:     "empty merge",
			original: NewCollection(New("username", "rufus")),
			merge:    NewCollection(),
			want:     NewCollection(New("username", "rufus")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.original.Merge(tc.merge)

			assert.Equal(t, tc.want, tc.original)
		})
	}
}

func TestCollection_Get(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		collection Collection
		key        NamespaceAndKind
		want       string
		wantOK     bool
	}{
		{
			name: "found",
			collection: NewCollection(
				New("username", "rufus"),
			),
			key:    "username",
			want:   "rufus",
			wantOK: true,
		},
		{
			name: "not found",
			collection: NewCollection(
				New("username", "rufus"),
			),
			key:    "email",
			want:   "",
			wantOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := tc.collection.Get(tc.key)

			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantOK, ok)
		})
	}
}

func TestCollection_MustGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		collection Collection
		key        NamespaceAndKind
		want       string
		wantPanic  bool
	}{
		{
			name: "found",
			collection: NewCollection(
				New("username", "rufus"),
			),
			key:       "username",
			want:      "rufus",
			wantPanic: false,
		},
		{
			name: "not found",
			collection: NewCollection(
				New("username", "rufus"),
			),
			key:       "email",
			wantPanic: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.wantPanic {
				assert.Panics(t, func() {
					tc.collection.MustGet(tc.key)
				})
			} else {
				assert.NotPanics(t, func() {
					got := tc.collection.MustGet(tc.key)
					assert.Equal(t, tc.want, got)
				})
			}
		})
	}
}

func TestCollection_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		collection Collection
		want       string
	}{
		{
			name:       "empty",
			collection: NewCollection(),
			want:       "[]",
		},
		{
			name: "one item",
			collection: NewCollection(
				New("username", "rufus"),
			),
			want: "[username:rufus]",
		},
		{
			name: "multiple items",
			collection: NewCollection(
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			),
			want: "[email:rufus@seatgeek.com username:rufus]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.collection.String()

			assert.Equal(t, tc.want, got)
		})
	}
}
