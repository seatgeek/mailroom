// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package validation

import "context"

// Validator can be implemented by any parser, generator, transport, etc. to validate its configuration at runtime
type Validator interface {
	// Validate should return an error if the configuration is invalid
	// Errors returned by Validate are considered fatal
	Validate(ctx context.Context) error
}
