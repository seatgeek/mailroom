// Copyright 2025 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package event_test

import (
	"context"
	"testing"
	"time"

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/identifier"
	"github.com/stretchr/testify/assert"
)

func TestContext_WithID(t *testing.T) {
	t.Parallel()

	originalContext := event.Context{
		ID:      "original-id",
		Source:  event.MustSource("https://www.example.com/foo"),
		Type:    "com.example.event",
		Subject: "subject",
		Time:    time.Now(),
	}

	newContext := originalContext.WithID("new-id")

	assert.NotSame(t, &originalContext, &newContext)
	assert.NotEqual(t, originalContext, newContext)

	assert.Equal(t, event.ID("original-id"), originalContext.ID)
	assert.Equal(t, event.ID("new-id"), newContext.ID)

	assert.Equal(t, originalContext.Source, newContext.Source)
	assert.Equal(t, originalContext.Type, newContext.Type)
	assert.Equal(t, originalContext.Subject, newContext.Subject)
	assert.Equal(t, originalContext.Time, newContext.Time)
}

func TestContext_WithSource(t *testing.T) {
	t.Parallel()

	originalSource := event.MustSource("https://www.example.com/foo")
	originalContext := event.Context{
		ID:      "original-id",
		Source:  originalSource,
		Type:    "com.example.event",
		Subject: "subject",
		Time:    time.Now(),
	}

	newSource := event.MustSource("https://www.example.com/bar")
	newContext := originalContext.WithSource(newSource)

	assert.NotSame(t, &originalContext, &newContext)
	assert.NotEqual(t, originalContext, newContext)

	assert.Equal(t, originalSource, originalContext.Source)
	assert.Equal(t, newSource, newContext.Source)

	assert.Equal(t, originalContext.ID, newContext.ID)
	assert.Equal(t, originalContext.Type, newContext.Type)
	assert.Equal(t, originalContext.Subject, newContext.Subject)
	assert.Equal(t, originalContext.Time, newContext.Time)
}

func TestContext_WithType(t *testing.T) {
	t.Parallel()

	originalContext := event.Context{
		ID:      "original-id",
		Source:  event.MustSource("https://www.example.com/foo"),
		Type:    "com.example.foo",
		Subject: "subject",
		Time:    time.Now(),
	}

	newContext := originalContext.WithType("com.example.bar")

	assert.NotSame(t, &originalContext, &newContext)
	assert.NotEqual(t, originalContext, newContext)

	assert.Equal(t, event.Type("com.example.foo"), originalContext.Type)
	assert.Equal(t, event.Type("com.example.bar"), newContext.Type)

	assert.Equal(t, originalContext.ID, newContext.ID)
	assert.Equal(t, originalContext.Source, newContext.Source)
	assert.Equal(t, originalContext.Subject, newContext.Subject)
	assert.Equal(t, originalContext.Time, newContext.Time)
}

func TestContext_WithSubject(t *testing.T) {
	t.Parallel()

	originalContext := event.Context{
		ID:      "original-id",
		Source:  event.MustSource("https://www.example.com/foo"),
		Type:    "com.example.event",
		Subject: "original-subject",
		Time:    time.Now(),
	}

	newContext := originalContext.WithSubject("new-subject")

	assert.NotSame(t, &originalContext, &newContext)
	assert.NotEqual(t, originalContext, newContext)

	assert.Equal(t, "original-subject", originalContext.Subject)
	assert.Equal(t, "new-subject", newContext.Subject)

	assert.Equal(t, originalContext.ID, newContext.ID)
	assert.Equal(t, originalContext.Source, newContext.Source)
	assert.Equal(t, originalContext.Type, newContext.Type)
	assert.Equal(t, originalContext.Time, newContext.Time)
}

func TestContext_WithTime(t *testing.T) {
	t.Parallel()

	originalTime := time.Now()
	originalContext := event.Context{
		ID:      "original-id",
		Source:  event.MustSource("https://www.example.com/foo"),
		Type:    "com.example.event",
		Subject: "subject",
		Time:    originalTime,
	}

	newTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	newContext := originalContext.WithTime(newTime)

	assert.NotSame(t, &originalContext, &newContext)
	assert.NotEqual(t, originalContext, newContext)

	assert.Equal(t, originalTime, originalContext.Time)
	assert.Equal(t, newTime, newContext.Time)

	assert.Equal(t, originalContext.ID, newContext.ID)
	assert.Equal(t, originalContext.Source, newContext.Source)
	assert.Equal(t, originalContext.Type, newContext.Type)
	assert.Equal(t, originalContext.Subject, newContext.Subject)
}

func TestContext_WithLabels(t *testing.T) {
	t.Parallel()

	originalLabels := map[string]string{
		"environment": "test",
		"version":     "1.0.0",
	}
	originalContext := event.Context{
		ID:      "original-id",
		Source:  event.MustSource("https://www.example.com/foo"),
		Type:    "com.example.event",
		Subject: "subject",
		Time:    time.Now(),
		Labels:  originalLabels,
	}

	newLabels := map[string]string{
		"environment": "production",
		"region":      "us-east-1",
	}
	newContext := originalContext.WithLabels(newLabels)

	assert.NotSame(t, &originalContext, &newContext)
	assert.NotEqual(t, originalContext, newContext)

	assert.Equal(t, originalLabels, originalContext.Labels)
	assert.Equal(t, newLabels, newContext.Labels)

	// Modifying the original labels map does not affect the new context
	originalLabels["new-key"] = "new-value"
	assert.NotContains(t, newContext.Labels, "new-key")

	// Modifying the newLabels map does affect the new context
	newLabels["another-key"] = "another-value"
	assert.Contains(t, newContext.Labels, "another-key")
	assert.Equal(t, "another-value", newContext.Labels["another-key"])

	assert.Equal(t, originalContext.ID, newContext.ID)
	assert.Equal(t, originalContext.Source, newContext.Source)
	assert.Equal(t, originalContext.Type, newContext.Type)
	assert.Equal(t, originalContext.Subject, newContext.Subject)
	assert.Equal(t, originalContext.Time, newContext.Time)

	// Test with nil labels
	contextWithNilLabels := event.Context{
		ID:     "test-id",
		Source: event.MustSource("https://example.com"),
		Type:   "test.type",
	}

	newContextFromNil := contextWithNilLabels.WithLabels(map[string]string{"key": "value"})
	assert.Nil(t, contextWithNilLabels.Labels)
	assert.Equal(t, map[string]string{"key": "value"}, newContextFromNil.Labels)
}

func TestContext_Copy(t *testing.T) {
	t.Parallel()

	originalContext := event.Context{
		ID:      "original-id",
		Source:  event.MustSource("https://www.example.com/foo"),
		Type:    "com.example.event",
		Subject: "subject",
		Time:    time.Now(),
		Labels:  map[string]string{"key": "value"},
	}

	copiedContext := originalContext.Copy()

	assert.NotSame(t, &originalContext, &copiedContext)
	assert.Equal(t, originalContext, copiedContext)

	assert.Equal(t, originalContext.ID, copiedContext.ID)
	assert.Equal(t, originalContext.Source, copiedContext.Source)
	assert.Equal(t, originalContext.Type, copiedContext.Type)
	assert.Equal(t, originalContext.Subject, copiedContext.Subject)
	assert.Equal(t, originalContext.Time, copiedContext.Time)
	assert.Equal(t, originalContext.Labels, copiedContext.Labels)

	originalContext.Labels["new-key"] = "new-value"
	assert.NotContains(t, copiedContext.Labels, "new-key")
	assert.Equal(t, "new-value", originalContext.Labels["new-key"])

	copiedContext.Labels["another-key"] = "another-value"
	assert.NotContains(t, originalContext.Labels, "another-key")
	assert.Equal(t, "another-value", copiedContext.Labels["another-key"])
}

func TestSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		uri     string
		wantNil bool
	}{
		{
			name: "valid uri with DNS authority",
			uri:  "https://www.example.com/foo",
		},
		{
			name: "mailto",
			uri:  "mailto:codell@seatgeek.com",
		},
		{
			name: "universally-unique URN with a UUID",
			uri:  "urn:uuid:6e8bc430-9c3a-11d9-9669-0800200c9a66",
		},
		{
			name: "application-specific identifier",
			uri:  "/cloudevents/spec/pull/123",
		},
		{
			name:    "empty uri",
			uri:     "",
			wantNil: true,
		},
		{
			name:    "invalid uri",
			uri:     ":This is not a URI:",
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			source := event.NewSource(tc.uri)

			if tc.wantNil {
				assert.Nil(t, source)
				return
			}

			assert.NotNil(t, source)
			assert.Equal(t, tc.uri, source.String())

			// We'll also test the MustSource function here since the cases are the same
			if tc.wantNil {
				assert.Panics(t, func() {
					event.MustSource(tc.uri)
				})
			} else {
				assert.NotPanics(t, func() {
					event.MustSource(tc.uri)
				})
			}
		})
	}
}

func TestProcessorFunc(t *testing.T) {
	t.Parallel()

	// A simple processor that drops notifications without recipients
	dropWithoutRecipients := func(ctx context.Context, evt event.Event, notifications []event.Notification) ([]event.Notification, error) {
		ret := make([]event.Notification, 0, len(notifications))
		for _, n := range notifications {
			if n.Recipient().Len() > 0 {
				ret = append(ret, n)
			}
		}

		return ret, nil
	}

	notificationWithRecipient := event.NewMockNotification(t)
	notificationWithRecipient.EXPECT().Recipient().Return(identifier.NewSet(identifier.New(identifier.GenericID, "123")))

	notificationWithoutRecipient := event.NewMockNotification(t)
	notificationWithoutRecipient.EXPECT().Recipient().Return(identifier.NewSet())

	// Call the processor function
	notifications, err := event.ProcessorFunc(dropWithoutRecipients).Process(t.Context(), event.Event{}, []event.Notification{notificationWithRecipient, notificationWithoutRecipient})

	assert.NoError(t, err)
	assert.Len(t, notifications, 1)
	assert.Contains(t, notifications, notificationWithRecipient)
	assert.NotContains(t, notifications, notificationWithoutRecipient)
}
