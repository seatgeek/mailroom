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

	someNotifications := []common.Notification{
		notification.NewBuilder("a1c11a53-c4be-488f-89b6-f83bf2d48dab", "com.example.event").Build(),
	}
	someError := errors.New("some error")

	tests := []struct {
		name     string
		source   source.Source
		notifier notifier.Notifier
		wantErr  error
	}{
		{
			name:     "happy path",
			source:   sourceThatReturns(t, someNotifications, nil),
			notifier: notifierThatReturns(t, nil),
			wantErr:  nil,
		},
		{
			name:     "no notifications generated",
			source:   sourceThatReturns(t, nil, nil),
			notifier: notifierThatReturns(t, nil),
			wantErr:  nil,
		},
		{
			name:    "parse error",
			source:  sourceThatReturns(t, nil, someError),
			wantErr: someError,
		},
		{
			name:     "notifier error",
			source:   sourceThatReturns(t, someNotifications, nil),
			notifier: notifierThatReturns(t, someError),
			wantErr:  someError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := CreateEventHandler(
				context.Background(),
				tc.source,
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

func sourceThatReturns(t *testing.T, notifs []common.Notification, err error) source.Source {
	t.Helper()

	src := source.MockSource{}
	src.On("Key").Return("some-source")
	src.On("Parse", mock.Anything).Return(notifs, err)

	return &src
}

func notifierThatReturns(t *testing.T, err error) notifier.Notifier {
	t.Helper()

	notif := notifier.MockNotifier{}
	notif.On("Push", mock.Anything, mock.Anything).Return(err)

	return &notif
}
