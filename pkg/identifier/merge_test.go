// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package identifier

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeAndDeduplicate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []Set
		want  []Set
	}{
		{
			name:  "no sets",
			input: []Set{},
			want:  nil,
		},
		{
			name: "one set",
			input: []Set{
				NewSet(New(GenericUsername, "rufus")),
			},
			want: []Set{
				NewSet(New(GenericUsername, "rufus")),
			},
		},
		{
			name: "multiple sets, no overlap",
			input: []Set{
				NewSet(New(GenericID, "111"), New(GenericUsername, "user1"), New(GenericEmail, "user1@example.com")),
				NewSet(New(GenericID, "222"), New(GenericUsername, "user2"), New(GenericEmail, "user2@example.com")),
				NewSet(New(GenericID, "333"), New(GenericUsername, "user3"), New(GenericEmail, "user3@example.com")),
			},
			want: []Set{
				NewSet(New(GenericID, "111"), New(GenericUsername, "user1"), New(GenericEmail, "user1@example.com")),
				NewSet(New(GenericID, "222"), New(GenericUsername, "user2"), New(GenericEmail, "user2@example.com")),
				NewSet(New(GenericID, "333"), New(GenericUsername, "user3"), New(GenericEmail, "user3@example.com")),
			},
		},
		{
			name: "multiple sets, some overlap",
			input: []Set{
				NewSet(New(GenericID, "111"), New(GenericUsername, "rufus")),
				NewSet(New(GenericUsername, "rufus"), New(GenericEmail, "rufus@example.com")),
				NewSet(New(GenericUsername, "somebodyelse")),
			},
			want: []Set{
				NewSet(New(GenericID, "111"), New(GenericUsername, "rufus"), New(GenericEmail, "rufus@example.com")),
				NewSet(New(GenericUsername, "somebodyelse")),
			},
		},
		{
			name: "multiple sets, all overlap",
			input: []Set{
				NewSet(New(GenericID, "111"), New(GenericUsername, "rufus")),
				NewSet(New(GenericEmail, "rufus@example.com")),
				NewSet(New(GenericUsername, "rufus"), New(GenericEmail, "rufus@example.com")),
			},
			want: []Set{
				NewSet(New(GenericID, "111"), New(GenericUsername, "rufus"), New(GenericEmail, "rufus@example.com")),
			},
		},
		{
			name: "some overlaps have different values for the same NamespaceAndKind",
			input: []Set{
				NewSet(New(GenericID, "111"), New(GenericUsername, "rufus")),
				NewSet(New(GenericUsername, "rufus"), New(GenericEmail, "rufus@example.com")),
				NewSet(New(GenericUsername, "rufus"), New(GenericEmail, "rufus@example.net")),
			},
			want: []Set{
				NewSet(New(GenericID, "111"), New(GenericUsername, "rufus"), New(GenericEmail, "rufus@example.net")),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := MergeAndDeduplicate(tc.input...)

			// We can't directly compare the slices since the order is not guaranteed, so we'll compare sorted string representations instead
			wantStr := sortThenStringify(t, tc.want)
			gotStr := sortThenStringify(t, got)
			assert.Equal(t, wantStr, gotStr)
		})
	}
}

// sortThenStringify is a test helper
func sortThenStringify(t *testing.T, sets []Set) string {
	t.Helper()

	result := make([]string, len(sets))
	for i, set := range sets {
		result[i] = set.String()
	}

	// Sort the strings so that the order doesn't matter
	sort.Strings(result)

	return strings.Join(result, ", ")
}
