// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package identifier

// MergeAndDeduplicate merges any sets that share overlapping identifiers, returning a slice of distinct sets,
// each representing a unique collection of identifiers without duplicates.
// Note that this function does not modify the input sets, nor does it guarantee the order of the output sets.
func MergeAndDeduplicate(sets ...Set) []Set {
	// parent maps each set index to its parent index
	parent := make(map[int]int)

	// Initialize each set's parent to itself
	for i := range sets {
		parent[i] = i
	}

	// idToSetIndices maps an Identifier to the indices of sets that contain it
	idToSetIndices := make(map[Identifier][]int)

	// Build the idToSetIndices map
	for i, set := range sets {
		ids := set.ToList()
		for _, id := range ids {
			idToSetIndices[id] = append(idToSetIndices[id], i)
		}
	}

	// Union sets that share identifiers
	for _, indices := range idToSetIndices {
		if len(indices) > 1 {
			for i := 1; i < len(indices); i++ {
				union(parent, indices[0], indices[i])
			}
		}
	}

	// Group sets by their root parent
	parentToSetIndices := make(map[int][]int)
	for i := range sets {
		root := find(parent, i)
		parentToSetIndices[root] = append(parentToSetIndices[root], i)
	}

	// Merge sets within the same group
	result := make([]Set, 0, len(parentToSetIndices))
	for _, indices := range parentToSetIndices {
		mergedSet := NewSet()
		for _, idx := range indices {
			mergedSet.Merge(sets[idx])
		}
		result = append(result, mergedSet)
	}

	return result
}

// find returns the root parent of the set index i
func find(parent map[int]int, i int) int {
	if parent[i] != i {
		parent[i] = find(parent, parent[i]) // Path compression
	}
	return parent[i]
}

// union merges the sets containing indices i and j
func union(parent map[int]int, i, j int) {
	rootI := find(parent, i)
	rootJ := find(parent, j)
	if rootI != rootJ {
		parent[rootJ] = rootI // Merge the two sets
	}
}
