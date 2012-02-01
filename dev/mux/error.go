// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"fmt"
)

// ErrMulti stores multiple errors.
type ErrMulti []error

// String returns a string representation of the error.
func (m ErrMulti) Error() string {
	s, n := "", 0
	for _, e := range m {
		if e == nil {
			continue
		}
		if n == 0 {
			s = e.Error()
		}
		n++
	}
	switch n {
	case 0:
		return "(0 errors)"
	case 1:
		return s
	case 2:
		return s + " (and 1 other error)"
	}
	return fmt.Sprintf("%s (and %d other errors)", s, n-1)
}
