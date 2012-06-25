// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pluralforms

import (
	"testing"
)

func TestParse(t *testing.T) {
	fNumber := 1
	for expr, fn := range pluralFuncs {
		fn2, err := Parse(expr)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 200; i++ {
			expected := fn(i)
			result := fn2(i)
			if result != expected {
				t.Fatalf("Expected %d, got %d during iteration %d. Expression: %s", expected, result, i, expr)
			}
			t.Logf("pluralFunc%d: expected %d, got %d during iteration %d.", fNumber, expected, result, i)
		}
		fNumber++
	}
}
