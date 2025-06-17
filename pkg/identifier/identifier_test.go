// Copyright 2025 SeatGeek, Inc.
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

func TestSet_ToList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		set  Set
		want []Identifier
	}{
		{
			name: "empty",
			set:  NewSet(),
			want: nil,
		},
		{
			name: "non-empty",
			set: NewSet(
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

			got := tc.set.ToList()

			assert.ElementsMatch(t, tc.want, got)
		})
	}
}

func TestSet_Add(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		original Set
		add      Identifier
		want     Set
	}{
		{
			name: "adds",
			original: NewSet(
				New("username", "rufus"),
			),
			add: New("email", "rufus@seatgeek.com"),
			want: NewSet(
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			),
		},
		{
			name: "overwrites",
			original: NewSet(
				New("username", "rufus"),
				New("email", "rufus@example.com"),
			),
			add: New("email", "rufus@seatgeek.com"),
			want: NewSet(
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			),
		},
		{
			name:     "empty original",
			original: NewSet(),
			add:      New("username", "rufus"),
			want:     NewSet(New("username", "rufus")),
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

func TestSet_Merge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		original Set
		merge    Set
		want     Set
	}{
		{
			name: "adds and overwrites",
			original: NewSet(
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			),
			merge: NewSet(
				New("id", "123"),
				New("email", "rufus@example.com"),
			),
			want: NewSet(
				New("username", "rufus"),
				New("email", "rufus@example.com"),
				New("id", "123"),
			),
		},
		{
			name:     "empty original",
			original: NewSet(),
			merge:    NewSet(New("username", "rufus")),
			want:     NewSet(New("username", "rufus")),
		},
		{
			name:     "empty merge",
			original: NewSet(New("username", "rufus")),
			merge:    NewSet(),
			want:     NewSet(New("username", "rufus")),
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

func TestSet_Intersect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		set1     Set
		set2     Set
		expected Set
	}{
		{
			name:     "both empty",
			set1:     NewSet(),
			set2:     NewSet(),
			expected: NewSet(),
		},
		{
			name: "one empty",
			set1: NewSet(),
			set2: NewSet(
				New("username", "rufus"),
			),
			expected: NewSet(),
		},
		{
			name: "no intersection",
			set1: NewSet(
				New("username", "rufus"),
			),
			set2: NewSet(
				New("email", "rufus@seatgeek.com"),
			),
			expected: NewSet(),
		},
		{
			name: "partial intersection",
			set1: NewSet(
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			),
			set2: NewSet(
				New("email", "rufus@seatgeek.com"),
				New("id", "123"),
			),
			expected: NewSet(
				New("email", "rufus@seatgeek.com"),
			),
		},
		{
			name: "full intersection",
			set1: NewSet(
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			),
			set2: NewSet(
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			),
			expected: NewSet(
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Test both directions to ensure the method is commutative
			result1 := tc.set1.Intersect(tc.set2)
			result2 := tc.set2.Intersect(tc.set1)

			assert.Equal(t, tc.expected, result1)
			assert.Equal(t, tc.expected, result2)
		})
	}
}

func TestSet_Get(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		set    Set
		key    NamespaceAndKind
		want   string
		wantOK bool
	}{
		{
			name: "found",
			set: NewSet(
				New("username", "rufus"),
			),
			key:    "username",
			want:   "rufus",
			wantOK: true,
		},
		{
			name: "not found",
			set: NewSet(
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

			got, ok := tc.set.Get(tc.key)

			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantOK, ok)
		})
	}
}

func TestSet_MustGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		set       Set
		key       NamespaceAndKind
		want      string
		wantPanic bool
	}{
		{
			name: "found",
			set: NewSet(
				New("username", "rufus"),
			),
			key:       "username",
			want:      "rufus",
			wantPanic: false,
		},
		{
			name: "not found",
			set: NewSet(
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
					tc.set.MustGet(tc.key)
				})
			} else {
				assert.NotPanics(t, func() {
					got := tc.set.MustGet(tc.key)
					assert.Equal(t, tc.want, got)
				})
			}
		})
	}
}

func TestSet_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		set  Set
		want string
	}{
		{
			name: "empty",
			set:  NewSet(),
			want: "[]",
		},
		{
			name: "one item",
			set: NewSet(
				New("username", "rufus"),
			),
			want: "[username:rufus]",
		},
		{
			name: "multiple items",
			set: NewSet(
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			),
			want: "[email:rufus@seatgeek.com username:rufus]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.set.String()

			assert.Equal(t, tc.want, got)
		})
	}
}
