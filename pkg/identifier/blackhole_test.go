// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package identifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlackholeIdentifier(t *testing.T) {
	t.Parallel()

	t.Run("KindBlackhole constant", func(t *testing.T) {
		assert.Equal(t, Kind("blackhole"), KindBlackhole)
	})

	t.Run("GenericBlackhole constant", func(t *testing.T) {
		assert.Equal(t, NamespaceAndKind("blackhole"), GenericBlackhole)
		namespace, kind := GenericBlackhole.Split()
		assert.Equal(t, "", namespace)
		assert.Equal(t, "blackhole", kind)
		assert.Equal(t, KindBlackhole, GenericBlackhole.Kind())
	})

	t.Run("create blackhole identifier", func(t *testing.T) {
		id := New("blackhole", "discard")
		assert.Equal(t, NamespaceAndKind("blackhole"), id.NamespaceAndKind)
		assert.Equal(t, "discard", id.Value)
		assert.Equal(t, KindBlackhole, id.NamespaceAndKind.Kind())
	})

	t.Run("create namespaced blackhole identifier", func(t *testing.T) {
		id := New("example.com/blackhole", "discard")
		assert.Equal(t, NamespaceAndKind("example.com/blackhole"), id.NamespaceAndKind)
		assert.Equal(t, "discard", id.Value)
		assert.Equal(t, KindBlackhole, id.NamespaceAndKind.Kind())
		
		namespace, kind := id.NamespaceAndKind.Split()
		assert.Equal(t, "example.com", namespace)
		assert.Equal(t, "blackhole", kind)
	})
}