// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pluralforms

import (
	"testing"
)

func TestParse(t *testing.T) {
	// Compare the results of the parsed expression and the precomputed
	// functions.
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
				t.Fatalf("pluralFunc%d: expected %d, got %d during iteration %d.", fNumber, expected, result, i, expr)
			}
			t.Logf("pluralFunc%d: expected %d, got %d during iteration %d.", fNumber, expected, result, i)
		}
		fNumber++
	}
	// Now some bad expressions.
	badExprs := []string{
		"1 *",
		"-1 * 2", // negative numbers are not allowed
		"1 (1)",
		"1 ?",
		"1 ? 2",
		"1 :",
		"1 : 2",
		"2 * (3 * (4 + 5)",
		"2 * (3 * (4 + 5)))",
	}
	for _, expr := range badExprs {
		fn, err := Parse(expr)
		if err == nil {
			for i := 0; i < 200; i++ {
				expected := -1
				result := fn(i)
				if result != expected {
					t.Fatalf("Expected %d, got %d during iteration %d. Expression: %s", expected, result, i, expr)
				}
			}
		}
	}
}
