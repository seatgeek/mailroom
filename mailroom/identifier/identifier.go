// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package identifier

import (
	"fmt"
	"strings"
)

// NamespaceAndKind is a combination of a namespace and a kind.
// The Namespace is optional, and if it is not present, it is represented as an empty string.
// The Kind is required.
type NamespaceAndKind struct {
	Namespace string
	Kind      Kind
}

var (
	GenericEmail    = NamespaceAndKind{Kind: KindEmail}
	GenericUsername = NamespaceAndKind{Kind: KindUsername}
	GenericID       = NamespaceAndKind{Kind: ID}
)

func (n NamespaceAndKind) String() string {
	if n.Namespace == "" {
		return string(n.Kind)
	}

	return fmt.Sprintf("%s/%s", n.Namespace, n.Kind)
}

func For(s string) NamespaceAndKind {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) == 1 {
		return NamespaceAndKind{Kind: Kind(parts[0])}
	}

	return NamespaceAndKind{
		Namespace: parts[0],
		Kind:      Kind(parts[1]),
	}
}

type Kind string

const (
	KindEmail    Kind = "email"
	KindUsername Kind = "username"
	ID           Kind = "id"
)

// An Identifier is a unique reference to some user or group.
type Identifier struct {
	NamespaceAndKind
	Value string
}

type nsKindType interface {
	~string | NamespaceAndKind
}

type valueType interface {
	~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func New[T1 nsKindType, T2 valueType](namespaceAndKind T1, value T2) Identifier {
	if nsAndKind, ok := any(namespaceAndKind).(NamespaceAndKind); ok {
		return Identifier{
			NamespaceAndKind: nsAndKind,
			Value:            fmt.Sprint(value),
		}
	}

	return Identifier{
		NamespaceAndKind: For(fmt.Sprint(namespaceAndKind)),
		Value:            fmt.Sprint(value),
	}
}

type Collection map[NamespaceAndKind]string

func (i *Collection) Get(query NamespaceAndKind) (Identifier, bool) {
	if *i == nil {
		return Identifier{}, false
	}

	for key, val := range *i {
		if query.Namespace != "" && query.Namespace != key.Namespace {
			continue
		}

		if query.Kind != "" && query.Kind != key.Kind {
			continue
		}

		return Identifier{
			NamespaceAndKind: key,
			Value:            val,
		}, true
	}

	return Identifier{}, false
}

func (i *Collection) ToList() []Identifier {
	if *i == nil {
		return nil
	}

	res := make([]Identifier, 0, len(*i))
	for key, val := range *i {
		res = append(res, Identifier{
			NamespaceAndKind: key,
			Value:            val,
		})
	}

	return res
}

func NewCollection(ids ...Identifier) Collection {
	res := Collection{}
	for _, id := range ids {
		res[id.NamespaceAndKind] = id.Value
	}

	return res
}
