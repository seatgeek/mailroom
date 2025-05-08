// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package server

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/notification"
	"github.com/seatgeek/mailroom/pkg/notifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	someEvent := event.Event{
		Context: event.Context{
			ID:     "a1c11a53-c4be-488f-89b6-f83bf2d48dab",
			Type:   "com.example.event",
			Source: event.MustSource("example.com"),
		},
		Data: "some payload",
	}
	someNotifications := []event.Notification{
		notification.NewBuilder(someEvent.Context).Build(),
	}
	someError := errors.New("some error")

	tests := []struct {
		name           string
		parser         event.Parser
		processors     []event.Processor
		notifier       notifier.Notifier
		wantStatusCode int
	}{
		{
			name:           "happy path",
			parser:         parserThatReturns(t, &someEvent, nil),
			processors:     []event.Processor{processorThatReturns(t, someNotifications, nil)},
			notifier:       notifierThatReturns(t, nil),
			wantStatusCode: 202,
		},
		{
			name:           "uninterested event",
			parser:         parserThatReturns(t, nil, nil),
			wantStatusCode: 200,
		},
		{
			name:           "no notifications generated",
			parser:         parserThatReturns(t, &someEvent, nil),
			processors:     []event.Processor{processorThatReturns(t, nil, nil)},
			notifier:       notifierThatReturns(t, nil),
			wantStatusCode: 200,
		},
		{
			name:           "parse error",
			parser:         parserThatReturns(t, nil, someError),
			wantStatusCode: 500,
		},
		{
			name: "parse error with custom HTTP status code",
			parser: parserThatReturns(t, nil, &Error{
				Code:   400,
				Reason: someError,
			}),
			wantStatusCode: 400,
		},
		{
			name:           "notifier error",
			parser:         parserThatReturns(t, &someEvent, nil),
			processors:     []event.Processor{processorThatReturns(t, someNotifications, nil)},
			notifier:       notifierThatReturns(t, someError),
			wantStatusCode: 202,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := CreateEventProcessingHandler(tc.parser, tc.processors, tc.notifier)

			writer := httptest.NewRecorder()

			handler(writer, httptest.NewRequest("POST", "/some-handler", nil))

			assert.Equal(t, tc.wantStatusCode, writer.Code)
		})
	}
}

func parserThatReturns(t *testing.T, evt *event.Event, err error) event.Parser {
	t.Helper()

	parser := event.NewMockParser(t)
	parser.EXPECT().Key().Return("some-parser")
	parser.EXPECT().Parse(mock.Anything).Return(evt, err)

	return parser
}

func processorThatReturns(t *testing.T, notifs []event.Notification, err error) event.Processor {
	t.Helper()

	processor := event.NewMockProcessor(t)
	processor.EXPECT().Process(mock.Anything, mock.Anything, mock.Anything).Return(notifs, err)

	return processor
}

func notifierThatReturns(t *testing.T, err error) notifier.Notifier {
	t.Helper()

	notif := notifier.MockNotifier{}
	notif.EXPECT().Push(mock.Anything, mock.Anything).Return(err)

	return &notif
}
