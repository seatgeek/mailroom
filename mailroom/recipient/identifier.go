// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package recipient

import (
	"fmt"
	"strings"
)

type IdentifierKind string

const (
	Email    IdentifierKind = "email"
	Username IdentifierKind = "username"
	ID       IdentifierKind = "id"
)

// An Identifier is a unique reference to some user or group.
// It is composed of an optional namespace, a required kind, and a required value.
type Identifier struct {
	Namespace string
	Kind      IdentifierKind
	Value     string
}

// NamespaceAndKind returns the namespace and kind of the identifier - suitable for use as an annotation/label.
func (i Identifier) NamespaceAndKind() string {
	namespaceAndKind := i.Namespace
	if namespaceAndKind != "" {
		namespaceAndKind += "/"
	}
	namespaceAndKind += string(i.Kind)

	return namespaceAndKind
}

type identifierValue interface {
	~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func NewIdentifier[T identifierValue](namespaceAndKind string, value T) Identifier {
	var namespace string
	var kind IdentifierKind

	parts := strings.SplitN(namespaceAndKind, "/", 2)
	if len(parts) == 2 {
		namespace = parts[0]
		kind = IdentifierKind(parts[1])
	} else {
		kind = IdentifierKind(parts[0])
	}

	return Identifier{
		Namespace: namespace,
		Kind:      kind,
		Value:     fmt.Sprint(value),
	}
}
