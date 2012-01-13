// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package schema

import (
	"fmt"
	"os"
)

// ErrMulti stores multiple errors.
//
// ErrMulti was borrowed appengine/datastore, from the App Engine SDK.
type ErrMulti []os.Error

// String returns a string representation of the error.
func (m ErrMulti) String() string {
	s, n := "", 0
	for _, e := range m {
		if e == nil {
			continue
		}
		if n == 0 {
			s = e.String()
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
