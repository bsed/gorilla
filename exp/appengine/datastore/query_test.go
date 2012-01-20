// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package datastore

import (
	"fmt"
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

func TestQueryGQL(t *testing.T) {
	c := getContext(t)
	defer c.Close()

	k := NewKey(c, "Kind", "", 42, nil)

	q1 := NewQuery("Kind").
		  Ancestor(k).
		  Filter("f1=", "v1").
		  Filter("f2=", true).
		  Order("f3").
		  Order("-f4")
	s1 := q1.GQL(nil)
	expect1 := fmt.Sprintf("SELECT * FROM Kind WHERE ANCESTOR IS KEY('%v') AND f1=\"v1\" AND f2=true ORDER BY f3 ASC, f4 DESC", k.Encode())
	if s1 != expect1 {
		t.Errorf("Unexpected Query.String()\nresult: %v\nexpect: %v\n", s1, expect1)
	}

	q2 := q1.Ancestor(nil).Filter("f5 >=", 42)
	s2 := q2.GQL(nil)
	expect2 := fmt.Sprintf("SELECT * FROM Kind WHERE f1=\"v1\" AND f2=true AND f5>=42 ORDER BY f3 ASC, f4 DESC")
	if s2 != expect2 {
		t.Errorf("Unexpected Query.String()\nresult: %v\nexpect: %v\n", s2, expect2)
	}

	s3 := q1.GQL(nil)
	if s3 != expect1 {
		t.Errorf("Unexpected Query.String()\nresult: %v\nexpect: %v\n", s3, expect1)
	}

	options := NewQueryOptions(200, 100)
	s4 := q1.GQL(options)
	expect4 := " LIMIT 100,200"
	if s4 != expect1 + expect4 {
		t.Errorf("Unexpected Query.String()\nresult: %v\nexpect: %v\n", s4, expect1 + expect4)
	}

	options = options.Limit(0)
	s5 := q1.GQL(options)
	expect5 := " OFFSET 100"
	if s5 != expect1 + expect5 {
		t.Errorf("Unexpected Query.String()\nresult: %v\nexpect: %v\n", s5, expect1 + expect5)
	}
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

	// Kindless ancestor query.
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
