// Copyright 2011 Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flattener

import (
	"fmt"
	//"reflect"
	"testing"
)

type A struct {
	A1 string   `schema:"a1"`
	A2 []string `schema:"a2"`
}

func TestFlatten(t *testing.T) {
	a1 := &A{
		A1: "lalala",
		A2:	[]string{"a", "b", "c"},
	}
	a2 := &a1
	items, err := Flatten(a2)
	if err != nil {
		t.Errorf("Error: %v", err.String())
	}
	fmt.Printf("%#v\n", items)
	for _, item := range items {
		fmt.Printf("%v\n", item)
	}
}
