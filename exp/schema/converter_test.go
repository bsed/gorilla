// Copyright 2011 Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"os"
	"testing"
)

func TestStringMapConverter(t *testing.T) {
	// Statically verify that StringMapConverter implements Converter.
	var _ Converter = (*StringMapConverter)(nil)

	c := NewStringMapConverter(map[string][]string{
		"v1": {"true", "false", "true"},
		"v2": {"4.2", "4.3", "4.4"},
		"v3": {"-42", "-43", "-44"},
		"v4": {"foo", "bar", "baz"},
		"v5": {"42", "43", "44"},
	})

	var err os.Error
	var v1 bool
	var v2 float64
	var v3 int64
	var v4 string
	var v5 uint64
	var v6 []bool
	var v7 []float64
	var v8 []int64
	var v9 []string
	var v10 []uint64

	v1, err = c.Bool("v1")
	if err != nil || v1 != true {
		t.Errorf("Error converting v1: %v (Error: %v)", v1, err)
	}
	v2, err = c.Float("v2")
	if err != nil || v2 != 4.2 {
		t.Errorf("Error converting v2: %v (Error: %v)", v2, err)
	}
	v3, err = c.Int("v3")
	if err != nil || v3 != -42 {
		t.Errorf("Error converting v3: %v (Error: %v)", v3, err)
	}
	v4, err = c.String("v4")
	if err != nil || v4 != "foo" {
		t.Errorf("Error converting v4: %v (Error: %v)", v4, err)
	}
	v5, err = c.Uint("v5")
	if err != nil || v5 != 42 {
		t.Errorf("Error converting v5: %v (Error: %v)", v5, err)
	}

	v1, err = c.Bool("v5")
	if err == nil {
		t.Errorf("Error converting v1: %v (Error: %v)", v1, err)
	}
	v2, err = c.Float("v1")
	if err == nil {
		t.Errorf("Error converting v2: %v (Error: %v)", v2, err)
	}
	v3, err = c.Int("v2")
	if err == nil {
		t.Errorf("Error converting v3: %v (Error: %v)", v3, err)
	}
	v5, err = c.Uint("v4")
	if err == nil {
		t.Errorf("Error converting v5: %v (Error: %v)", v5, err)
	}

	v6, err = c.BoolMulti("v1")
	if err != nil || len(v6) != 3 || v6[0] != true || v6[1] != false || v6[2] != true {
		t.Errorf("Error converting v1: %v (Error: %v)", v6, err)
	}
	v7, err = c.FloatMulti("v2")
	if err != nil || len(v7) != 3 || v7[0] != 4.2 || v7[1] != 4.3 || v7[2] != 4.4 {
		t.Errorf("Error converting v2: %v (Error: %v)", v7, err)
	}
	v8, err = c.IntMulti("v3")
	if err != nil || len(v8) != 3 || v8[0] != -42 || v8[1] != -43 || v8[2] != -44 {
		t.Errorf("Error converting v3: %v (Error: %v)", v8, err)
	}
	v9, err = c.StringMulti("v4")
	if err != nil || len(v9) != 3 || v9[0] != "foo" || v9[1] != "bar" || v9[2] != "baz" {
		t.Errorf("Error converting v4: %v (Error: %v)", v9, err)
	}
	v10, err = c.UintMulti("v5")
	if err != nil || len(v10) != 3 || v10[0] != 42 || v10[1] != 43 || v10[2] != 44 {
		t.Errorf("Error converting v5: %v (Error: %v)", v10, err)
	}
}
