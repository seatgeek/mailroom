// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package common

import (
	"fmt"
	"strings"
)

// NamespaceAndKind is a combination of a namespace and a kind.
// The Namespace is optional, and if it is not present, it is represented as an empty string.
// The Kind is required.
type NamespaceAndKind struct {
	Namespace string
	Kind      IdentifierKind
}

func (n NamespaceAndKind) String() string {
	if n.Namespace == "" {
		return string(n.Kind)
	}

	return fmt.Sprintf("%s/%s", n.Namespace, n.Kind)
}

func parseNamespaceAndKind(s string) NamespaceAndKind {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) == 1 {
		return NamespaceAndKind{Kind: IdentifierKind(parts[0])}
	}

	return NamespaceAndKind{
		Namespace: parts[0],
		Kind:      IdentifierKind(parts[1]),
	}
}

type IdentifierKind string

const (
	Email    IdentifierKind = "email"
	Username IdentifierKind = "username"
	ID       IdentifierKind = "id"
)

// An Identifier is a unique reference to some user or group.
type Identifier struct {
	NamespaceAndKind
	Value string
}

type identifierValue interface {
	~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func NewIdentifier[T identifierValue](namespaceAndKind string, value T) Identifier {
	return Identifier{
		NamespaceAndKind: parseNamespaceAndKind(namespaceAndKind),
		Value:            fmt.Sprint(value),
	}
}
