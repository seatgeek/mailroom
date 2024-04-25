// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package server

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/notification"
	"github.com/seatgeek/mailroom/mailroom/notifier"
	"github.com/seatgeek/mailroom/mailroom/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	somePayload := &struct{}{}
	someNotifications := []common.Notification{
		notification.NewBuilder("a1c11a53-c4be-488f-89b6-f83bf2d48dab", "com.example.event").Build(),
	}
	someError := errors.New("some error")

	tests := []struct {
		name      string
		parser    source.PayloadParser
		generator source.NotificationGenerator
		notifier  notifier.Notifier
		wantErr   error
	}{
		{
			name:      "happy path",
			parser:    parserThatReturns(t, somePayload, nil),
			generator: generatorThatReturns(t, someNotifications, nil),
			notifier:  notifierThatReturns(t, nil),
			wantErr:   nil,
		},
		{
			name:    "parser error",
			parser:  parserThatReturns(t, nil, someError),
			wantErr: someError,
		},
		{
			name:      "generator error",
			parser:    parserThatReturns(t, somePayload, nil),
			generator: generatorThatReturns(t, nil, someError),
			wantErr:   someError,
		},
		{
			name:      "notifier error",
			parser:    parserThatReturns(t, somePayload, nil),
			generator: generatorThatReturns(t, someNotifications, nil),
			notifier:  notifierThatReturns(t, someError),
			wantErr:   someError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := CreateEventHandler(
				context.Background(),
				&source.Source{
					Key:       "some-source",
					Parser:    tc.parser,
					Generator: tc.generator,
				},
				tc.notifier,
			)

			writer := httptest.NewRecorder()

			err := handler(writer, httptest.NewRequest("POST", "/some-source", nil))

			if tc.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, 200, writer.Code)
			} else {
				assert.Errorf(t, err, tc.wantErr.Error())
			}
		})
	}
}

func parserThatReturns(t *testing.T, payload *struct{}, err error) source.PayloadParser {
	t.Helper()

	parser := source.MockPayloadParser{}
	parser.On("Parse", mock.Anything).Return(payload, err)

	return &parser
}

func generatorThatReturns(t *testing.T, notifications []common.Notification, err error) source.NotificationGenerator {
	t.Helper()

	generator := source.MockNotificationGenerator{}
	generator.On("Generate", mock.Anything).Return(notifications, err)

	return &generator
}

func notifierThatReturns(t *testing.T, err error) notifier.Notifier {
	t.Helper()

	notif := notifier.MockNotifier{}
	notif.On("Push", mock.Anything, mock.Anything).Return(err)

	return &notif
}
