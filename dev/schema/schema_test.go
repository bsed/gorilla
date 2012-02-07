// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"reflect"
	"testing"
)

type Struct1 struct {
	F1 Struct2
}
type Struct2 struct {
	F2 Struct3 `schema:"f2"`
}
type Struct3 struct {
	F3 string
	F4 []int
	F5 map[string]string
}

func TestSchema(t *testing.T) {
	s1 := &Struct1{}
	v := reflect.ValueOf(s1)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	values := map[string][]string{
		"F1.f2.F3":     {"Hello, world."},
		"F1.f2.F4":     {"42", "43", "44"},
		"F1.f2.F5.foo": {"bar"},
		"F1.f2.F5.baz": {"ding"},
	}
	LoadStruct(values, s1)

	t.Errorf("V: %v", s1.F1.F2.F3)
	t.Errorf("V: %v", s1.F1.F2.F4)
	t.Errorf("V: %v", s1.F1.F2.F5)
}
