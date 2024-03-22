// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package identifier

import (
	"fmt"
	"strings"
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

// Collection is a map of NamespaceAndKind to a value.
// Each entry is basically an Identifier.
type Collection map[NamespaceAndKind]string

// Email returns any email Identifier in the Collection, or false if none exists.
func (c *Collection) Email() (Identifier, bool) {
	if c == nil {
		return Identifier{}, false
	}

	// Prefer the generic (non-namespaced) email if it exists
	if val, ok := (*c)[GenericEmail]; ok {
		return Identifier{
			NamespaceAndKind: GenericEmail,
			Value:            val,
		}, true
	}

	// Otherwise any email will do
	for nsAndKind, val := range *c {
		_, kind := nsAndKind.Split()
		if kind == string(KindEmail) {
			return Identifier{
				NamespaceAndKind: nsAndKind,
				Value:            val,
			}, true
		}
	}

	return Identifier{}, false
}

// ToList returns the Collection as a slice of Identifier objects.
func (c *Collection) ToList() []Identifier {
	if c == nil {
		return nil
	}

	res := make([]Identifier, 0, len(*c))
	for key, val := range *c {
		res = append(res, Identifier{
			NamespaceAndKind: key,
			Value:            val,
		})
	}

	return res
}

// NewCollection creates a new Collection from a slice of Identifier objects
func NewCollection(ids ...Identifier) Collection {
	res := Collection{}
	for _, id := range ids {
		res[id.NamespaceAndKind] = id.Value
	}

	return res
}
