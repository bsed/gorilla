// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package datastore

import (
	"fmt"
	"testing"
)

func TestCursor(t *testing.T) {
	c := getContext(t)
	defer c.Close()

	e := &struct{}{}
	keys := make([]*Key, 50)
	entities := make([]interface{}, 50)
	for i := 0; i < 50; i++ {
		keys[i] = NewKey(c, "A", fmt.Sprintf("%03d", i), 0, nil)
		entities[i] = e
	}

	if _, err := PutMulti(c, keys, entities); err != nil {
		t.Errorf("Error on PutMulti(): %v\n", err)
	}

	q1 := NewQuery("A")
	i1 := q1.Run(c, NewQueryOptions(0, 0).Compile(true))

	i2 := q1.Run(c, NewQueryOptions(1, 0).Compile(true).Cursor(i1.CursorAt(5)))
	k2, _ := i2.Next(struct{}{})
	if k2.StringID() != "005" {
		t.Errorf("Expected %q string id, got %q", "005", k2.StringID())
	}

	i3 := q1.Run(c, NewQueryOptions(1, 0).Compile(true).Cursor(i1.CursorAt(42)))
	k3, _ := i3.Next(struct{}{})
	if k3.StringID() != "042" {
		t.Errorf("Expected %q string id, got %q", "042", k3.StringID())
	}
}
