// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package server

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/seatgeek/mailroom/pkg/common"
	"github.com/seatgeek/mailroom/pkg/event"
	"github.com/seatgeek/mailroom/pkg/handler"
	"github.com/seatgeek/mailroom/pkg/notification"
	"github.com/seatgeek/mailroom/pkg/notifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	someNotifications := []common.Notification{
		notification.NewBuilder(event.Context{ID: "a1c11a53-c4be-488f-89b6-f83bf2d48dab", Type: "com.example.event"}).Build(),
	}
	someError := errors.New("some error")

	tests := []struct {
		name     string
		handler  handler.Handler
		notifier notifier.Notifier
		wantErr  error
	}{
		{
			name:     "happy path",
			handler:  handlerThatReturns(t, someNotifications, nil),
			notifier: notifierThatReturns(t, nil),
			wantErr:  nil,
		},
		{
			name:     "no notifications generated",
			handler:  handlerThatReturns(t, nil, nil),
			notifier: notifierThatReturns(t, nil),
			wantErr:  nil,
		},
		{
			name:    "parse error",
			handler: handlerThatReturns(t, nil, someError),
			wantErr: someError,
		},
		{
			name:     "notifier error",
			handler:  handlerThatReturns(t, someNotifications, nil),
			notifier: notifierThatReturns(t, someError),
			wantErr:  someError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := CreateEventHandler(
				tc.handler,
				tc.notifier,
			)

			writer := httptest.NewRecorder()

			err := handler(writer, httptest.NewRequest("POST", "/some-handler", nil))

			if tc.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, 200, writer.Code)
			} else {
				assert.Errorf(t, err, tc.wantErr.Error())
			}
		})
	}
}

func handlerThatReturns(t *testing.T, notifs []common.Notification, err error) handler.Handler {
	t.Helper()

	src := handler.NewMockHandler(t)
	src.EXPECT().Key().Return("some-handler")
	src.EXPECT().Process(mock.Anything).Return(notifs, err)

	return src
}

func notifierThatReturns(t *testing.T, err error) notifier.Notifier {
	t.Helper()

	notif := notifier.MockNotifier{}
	notif.EXPECT().Push(mock.Anything, mock.Anything).Return(err)

	return &notif
}
