// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package datastore

import (
	"testing"
)

func getKeyMap(t *testing.T, iter *Iterator) map[string]*Key {
	m := make(map[string]*Key)
	for {
		key, err := iter.Next(&struct{}{})
		if err != nil {
			if err == Done {
				break
			}
			t.Errorf("Error on Run(): %v\n", err)
			break
		}
		m[key.Encode()] = key
	}
	return m
}

func TestKindlessQuery(t *testing.T) {
	c := getContext(t)
	defer c.Close()

	k1 := NewKey(c, "A", "a", 0, nil)
	k2 := NewKey(c, "B", "b", 0, nil)
	k3 := NewKey(c, "C", "c", 0, nil)
	e := &struct{}{}
	if _, err := PutMulti(c, []*Key{k1, k2, k3}, []interface{}{e, e, e}); err != nil {
		t.Errorf("Error on PutMulti(): %v\n", err)
	}

	// Order on __key__ ascending.
	q1 := NewQuery("").Order("__key__")
	m1 := getKeyMap(t, q1.Run(c, NewQueryOptions(10, 0)))
	if len(m1) != 3 || m1[k1.Encode()] == nil || m1[k2.Encode()] == nil || m1[k3.Encode()] == nil {
		t.Errorf("Expected 3 results, got %v\n", m1)
	}

	// Filter on __key__.
	q2 := q1.Filter("__key__>", k1)
	m2 := getKeyMap(t, q2.Run(c, NewQueryOptions(10, 0)))
	if len(m2) != 2 || m2[k2.Encode()] == nil || m2[k3.Encode()] == nil {
		t.Errorf("Expected 2 results, got %v\n", m2)
	}
}

func TestKindlessAncestorQuery(t *testing.T) {
	c := getContext(t)
	defer c.Close()

	k1 := NewKey(c, "A", "a", 0, nil)
	k2 := NewKey(c, "B", "b", 0, k1)
	k3 := NewKey(c, "C", "c", 0, k2)
	e := &struct{}{}
	if _, err := PutMulti(c, []*Key{k1, k2, k3}, []interface{}{e, e, e}); err != nil {
		t.Errorf("Error on PutMulti(): %v\n", err)
	}

	// Order on __key__ ascending.
	q1 := NewQuery("").Order("__key__").Ancestor(k1)
	m1 := getKeyMap(t, q1.Run(c, NewQueryOptions(10, 0)))
	if len(m1) != 3 || m1[k1.Encode()] == nil || m1[k2.Encode()] == nil || m1[k3.Encode()] == nil {
		t.Errorf("Expected 3 results, got %v\n", m1)
	}

	// Filter on __key__.
	q2 := q1.Filter("__key__>", k1)
	m2 := getKeyMap(t, q2.Run(c, NewQueryOptions(10, 0)))
	if len(m2) != 2 || m2[k2.Encode()] == nil || m2[k3.Encode()] == nil {
		t.Errorf("Expected 2 results, got %v\n", m2)
	}
}
