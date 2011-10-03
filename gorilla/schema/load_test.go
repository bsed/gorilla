// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"testing"
)

// ----------------------------------------------------------------------------

type TestStruct1 struct {
	F001 bool
	F002 int
	F003 string
	F004 TestStruct2
	F005 **TestStruct2
	F006 map[string]string
	F007 []string
	F008 []int
}

type TestStruct2 struct {
	F001 string
	F002 **TestStruct2
}

func TestLoad(t *testing.T) {
	values := map[string][]string{
		"F001":           {"true"},
		"F002":           {"42"},
		"F003":           {"string 1"},
		"F004.F001":      {"string 2"},
		"F005.F001":      {"string 3"},
		"F005.F002.F001": {"string 4"},
		"F006[foo]":      {"foo value"},
		"F006[bar]":      {"bar value"},
		"F007":           {"value 1", "value 2", "value 3"},
		"F008":           {"42", "43", "44"},
	}

	s := &TestStruct1{}
	Load(s, values)

	if s.F001 != true {
		t.Errorf("Expected %v, got %v.", true, s.F001)
	}
	if s.F002 != 42 {
		t.Errorf("Expected %v, got %v.", 42, s.F002)
	}
	if s.F003 != "string 1" {
		t.Errorf("Expected %v, got %v.", "string 1", s.F003)
	}
	if s.F004.F001 != "string 2" {
		t.Errorf("Expected %v, got %v.", "string 2", s.F004.F001)
	}
	if (*(*s.F005)).F001 != "string 3" {
		t.Errorf("Expected %v, got %v.", "string 3", (*(*s.F005)).F001)
	}
	if (*(*(*(*s.F005)).F002)).F001 != "string 4" {
		t.Errorf("Expected %v, got %v.", "string 4", (*(*(*(*s.F005)).F002)).F001)
	}
	if len(s.F006) != 2 || s.F006["foo"] != "foo value" || s.F006["bar"] != "bar value" {
		t.Errorf("Expected filled map, got %v.", s.F006)
	}
	if len(s.F007) != 3 || s.F007[0] != "value 1" || s.F007[1] != "value 2" || s.F007[2] != "value 3" {
		t.Errorf("Expected %v, got %v.", values["F007"], s.F007)
	}
	if len(s.F008) != 3 || s.F008[0] != 42 || s.F008[1] != 43 || s.F008[2] != 44 {
		t.Errorf("Expected %v, got %v.", values["F008"], s.F008)
	}
}
