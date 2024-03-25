// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package identifier

import (
	"fmt"
	"strings"
	"sync"
)

// NamespaceAndKind is a combination of a namespace and a kind.
// For example, in "slack.com/email", "slack.com" is the namespace and "email" is the kind.
// The namespace part is considered optional, and if it is not present, it is represented as an empty string.
type NamespaceAndKind string

var (
	GenericEmail    = NamespaceAndKind(KindEmail)
	GenericUsername = NamespaceAndKind(KindUsername)
	GenericID       = NamespaceAndKind(KindID)
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

func (n NamespaceAndKind) Kind() string {
	_, kind := n.Split()
	return kind
}

// NewNamespaceAndKind creates a new NamespaceAndKind from a namespace and a kind.
func NewNamespaceAndKind[T ~string](namespace string, kind T) NamespaceAndKind {
	if namespace == "" {
		return NamespaceAndKind(kind)
	}

	return NamespaceAndKind(fmt.Sprintf("%s/%s", namespace, kind))
}

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

// Collection holds a thread-safe map of NamespaceAndKind to a value.
// Each entry is basically an Identifier.
type Collection interface {
	Get(NamespaceAndKind) (string, bool)
	MustGet(NamespaceAndKind) string
	Add(Identifier)
	Merge(Collection)
	ToList() []Identifier
}

type collection struct {
	ids   map[NamespaceAndKind]string
	mutex sync.RWMutex
}

func (c *collection) Get(namespaceAndKind NamespaceAndKind) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	val, ok := c.ids[namespaceAndKind]
	return val, ok
}

func (c *collection) MustGet(namespaceAndKind NamespaceAndKind) string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	val, ok := c.ids[namespaceAndKind]
	if !ok {
		panic(fmt.Sprintf("no value found for %s", namespaceAndKind))
	}

	return val
}

func (c *collection) Add(id Identifier) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.ids[id.NamespaceAndKind] = id.Value
}

func (c *collection) Merge(otherIdentifiers Collection) {
	for _, id := range otherIdentifiers.ToList() {
		c.Add(id)
	}
}

// ToList returns the Collection as a slice of Identifier objects.
func (c *collection) ToList() []Identifier {
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

func (c *collection) String() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return fmt.Sprintf("%v", c.ids)
}

// NewCollection creates a new Collection from a slice of Identifier objects
func NewCollection(ids ...Identifier) Collection {
	res := &collection{
		ids: make(map[NamespaceAndKind]string, len(ids)),
	}

	for _, id := range ids {
		res.ids[id.NamespaceAndKind] = id.Value
	}

	return res
}
