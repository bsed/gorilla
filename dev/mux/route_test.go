// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"testing"
)

func TestUniqueVars(t *testing.T) {
	s1 := []string{"a", "b", "c"}
	s2 := []string{"d", "e", "f"}
	if err := uniqueVars(s1, s2); err != nil {
		t.Errorf("should be unique: %v %v", s1, s2)
	}
	if err := uniqueVars(s1, s1); err == nil {
		t.Errorf("should not be unique: %v %v", s1, s1)
	}
	if err := uniqueVars(s2, s2); err == nil {
		t.Errorf("should not be unique: %v %v", s2, s2)
	}
}
