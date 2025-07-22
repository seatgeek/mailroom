// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

// Package identifier provides a way to identify users across different systems.
package identifier

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// NamespaceAndKind is a combination of a namespace and a kind.
// For example, in "slack.com/email", "slack.com" is the namespace and "email" is the kind.
// The namespace part is considered optional, and if it is not present, it is represented as an empty string.
// This is useful when a user is known by different emails, usernames, or IDs across different systems.
type NamespaceAndKind string

var (
	// GenericEmail is any email address not associated with a specific namespace or system.
	GenericEmail = NamespaceAndKind(KindEmail)

	// GenericUsername is any username not associated with a specific namespace or system.
	GenericUsername = NamespaceAndKind(KindUsername)

	// GenericID is any ID not associated with a specific namespace or system.
	GenericID = NamespaceAndKind(KindID)
)

// Split returns the namespace and kind parts of the NamespaceAndKind.
func (n NamespaceAndKind) Split() (string, string) {
	parts := strings.SplitN(string(n), "/", 2)
	if len(parts) == 1 {
		return "", parts[0]
	}

	return parts[0], parts[1]
}

func (n NamespaceAndKind) Namespace() string {
	namespace, _ := n.Split()
	return namespace
}

func (n NamespaceAndKind) Kind() Kind {
	_, kind := n.Split()
	return Kind(kind)
}

// NewNamespaceAndKind creates a new NamespaceAndKind from a namespace and a Kind.
func NewNamespaceAndKind[T ~string](namespace string, kind T) NamespaceAndKind {
	if namespace == "" {
		return NamespaceAndKind(kind)
	}

	return NamespaceAndKind(fmt.Sprintf("%s/%s", namespace, kind))
}

// Kind represents the type of identifier, such as an "email" or "username".
// This is used in conjunction with a namespace to uniquely identify a user in some system.
type Kind string

const (
	KindEmail    Kind = "email"
	KindUsername Kind = "username"
	KindID       Kind = "id"
)

// An Identifier is a unique reference to some user or group.
type Identifier struct {
	NamespaceAndKind
	Value string
}

// valueType is used by the generic New function to allow any string or integer type to be passed as the value argument.
type valueType interface {
	~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// New creates a new Identifier for a given namespaceAndKind and a value.
func New[T1 ~string, T2 valueType](namespaceAndKind T1, value T2) Identifier {
	return Identifier{
		NamespaceAndKind: NamespaceAndKind(namespaceAndKind),
		Value:            fmt.Sprint(value),
	}
}

// Set holds a thread-safe map of NamespaceAndKind to a value.
// Each entry is basically an Identifier.
type Set interface {
	Get(NamespaceAndKind) (string, bool)
	MustGet(NamespaceAndKind) string
	Add(Identifier)
	Merge(Set)
	Intersect(Set) Set
	ToList() []Identifier
	String() string
	ToMap() map[NamespaceAndKind]string
	Len() int
	Copy() Set
}

type set struct {
	ids   map[NamespaceAndKind]string
	mutex sync.RWMutex
}

func (c *set) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.ids)
}

func (c *set) Get(namespaceAndKind NamespaceAndKind) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	val, ok := c.ids[namespaceAndKind]
	return val, ok
}

func (c *set) MustGet(namespaceAndKind NamespaceAndKind) string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	val, ok := c.ids[namespaceAndKind]
	if !ok {
		panic(fmt.Sprintf("no value found for %s", namespaceAndKind))
	}

	return val
}

func (c *set) Add(id Identifier) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.ids[id.NamespaceAndKind] = id.Value
}

// Merge adds all the identifiers from another Set to this Set.
func (c *set) Merge(otherIdentifiers Set) {
	for _, id := range otherIdentifiers.ToList() {
		c.Add(id)
	}
}

// Intersect returns a new Set that contains only the identifiers that are present in both this Set and another Set.
func (c *set) Intersect(other Set) Set {
	if c.Len() == 0 || other.Len() == 0 {
		return NewSet()
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var common []Identifier
	for _, id := range other.ToList() {
		if val, ok := c.ids[id.NamespaceAndKind]; ok && val == id.Value {
			common = append(common, id)
		}
	}

	return NewSet(common...)
}

// ToList returns the Set as a slice of Identifier objects.
func (c *set) ToList() []Identifier {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	res := make([]Identifier, 0, len(c.ids))
	for key, val := range c.ids {
		res = append(res, Identifier{
			NamespaceAndKind: key,
			Value:            val,
		})
	}

	return res
}

func (c *set) String() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return strings.TrimPrefix(fmt.Sprintf("%v", c.ids), "map")
}

func (c *set) MarshalJSON() ([]byte, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return json.Marshal(c.ids)
}

// Copy creates a deep copy of the Set.
func (c *set) Copy() Set {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	res := &set{
		ids: make(map[NamespaceAndKind]string, len(c.ids)),
	}

	for key, value := range c.ids {
		res.ids[key] = value
	}

	return res
}

// NewSet creates a new Set from a slice of Identifier objects
func NewSet(ids ...Identifier) Set {
	res := &set{
		ids: make(map[NamespaceAndKind]string, len(ids)),
	}

	for _, id := range ids {
		res.ids[id.NamespaceAndKind] = id.Value
	}

	return res
}

// NewSetFromMap creates a new Set from a map of NamespaceAndKind to value.
func NewSetFromMap(ids map[NamespaceAndKind]string) Set {
	res := &set{
		ids: make(map[NamespaceAndKind]string, len(ids)),
	}

	for key, value := range ids {
		res.ids[key] = value
	}

	return res
}

// ToMap returns the Set as a map of NamespaceAndKind to value from a Set.
func (c *set) ToMap() map[NamespaceAndKind]string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	res := make(map[NamespaceAndKind]string, len(c.ids))
	for key, value := range c.ids {
		res[key] = value
	}

	return res
}
