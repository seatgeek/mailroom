// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package mailroom

import (
	"context"
	"testing"
	"time"

	"github.com/seatgeek/mailroom/mailroom/source"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	s := New()

	assert.NotNil(t, s)
	assert.Equal(t, "0.0.0.0:8000", s.listenAddr)
}

func TestWithSources(t *testing.T) {
	src1 := &source.Source{ID: "foo"}
	src2 := &source.Source{ID: "bar"}
	s := New(WithSources(src1, src2))

	assert.NotNil(t, s)
	assert.Contains(t, s.sources, src1)
	assert.Contains(t, s.sources, src2)
}

func TestRun(t *testing.T) {
	s := New(WithListenAddr(":0"))
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()

	err := s.Run(ctx)

	assert.Nil(t, err)
}
