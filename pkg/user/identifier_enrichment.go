// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package user

import (
	"context"
	"errors"
	"log/slog"

	"github.com/seatgeek/mailroom/pkg/event"
)

// IdentifierEnrichmentProcessor is a processor that enriches notifications
// with all known identifiers for the recipient from the user.Store.
type IdentifierEnrichmentProcessor struct {
	userStore Store
}

// NewIdentifierEnrichmentProcessor creates a new IdentifierEnrichmentProcessor.
func NewIdentifierEnrichmentProcessor(us Store) *IdentifierEnrichmentProcessor {
	if us == nil {
		panic("user.Store cannot be nil for IdentifierEnrichmentProcessor")
	}

	return &IdentifierEnrichmentProcessor{userStore: us}
}

// Process enriches each notification's recipient with additional identifiers.
func (p *IdentifierEnrichmentProcessor) Process(ctx context.Context, evt event.Event, notifications []event.Notification) ([]event.Notification, error) {
	for _, n := range notifications {
		recipient := n.Recipient()
		if recipient == nil {
			continue
		}

		// Attempt to find the user based on the existing recipient identifiers.
		foundUser, err := p.userStore.Find(ctx, recipient)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				slog.DebugContext(ctx, "user not found for identifier enrichment", "eventID", evt.ID, "recipient", recipient.String())
			} else {
				slog.WarnContext(ctx, "error finding user for identifier enrichment", "eventID", evt.ID, "recipient", recipient.String(), "error", err)
			}
			continue
		}

		recipient.Merge(foundUser.Identifiers)
	}

	return notifications, nil
}
