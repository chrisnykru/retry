// Copyright (c) 2017 OctoLab. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Package retry provides functional mechanism based on context
// to perform actions repetitively until successful.
//
// This package is an extended version of https://godoc.org/github.com/Rican7/retry.
// Copyright © 2016 Trevor N. Suarez (Rican7)
package retry // import "github.com/kamilsk/retry"

import (
	"errors"

	"github.com/kamilsk/retry/strategy"
)

// Action defines a callable function that package retry can handle.
type Action func(attempt uint) error

// Retry takes an action and performs it, repetitively, until successful.
//
// Optionally, strategies may be passed that assess whether or not an attempt
// should be made.
func Retry(deadline <-chan struct{}, action Action, strategies ...strategy.Strategy) error {
	var attempt uint

	if len(strategies) == 0 {
		return action(attempt)
	}

	var err error
	done := make(chan struct{})
	go func() {
		// TODO ctx.Err() should be replaced correctly, atomic is good candidate
		for ; (attempt == 0 || err != nil) && shouldAttempt(attempt, err, strategies...); attempt++ {
			err = action(attempt)
		}
		close(done)
	}()

	select {
	case <-deadline:
		return errTimeout
	case <-done:
		return err
	}
}

var errTimeout = errors.New("operation timeout")

// shouldAttempt evaluates the provided strategies with the given attempt to
// determine if the Retry loop should make another attempt.
func shouldAttempt(attempt uint, err error, strategies ...strategy.Strategy) bool {
	shouldAttempt := true

	for i, repeat := 0, len(strategies); shouldAttempt && i < repeat; i++ {
		shouldAttempt = shouldAttempt && strategies[i](attempt, err)
	}

	return shouldAttempt
}
