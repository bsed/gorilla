// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package datastore

import (
	"fmt"
	"testing"
	"gae-go-testing.googlecode.com/git/appenginetesting"
)

func getContext(t *testing.T) *appenginetesting.Context {
	c, err := appenginetesting.NewContext(nil)
	if err != nil {
		t.Fatalf("NewContext: %v", err)
	}
	return c
}

func TestQueryString(t *testing.T) {
	c := getContext(t)
	defer c.Close()

	k := NewKey(c, "Wiki", "", 42, nil)
	q := NewQuery("Wiki").
		 Ancestor(k).
		 Filter("section=", "golang").
		 Filter("public=", true).
		 Order("title").
		 Order("-updated")

	s := q.String()
	expected := fmt.Sprintf("SELECT * FROM Wiki WHERE ANCESTOR IS KEY('%v') AND section=\"golang\" AND public=true ORDER BY title ASC, updated DESC", k.Encode())
	if s != expected {
		t.Errorf("Unexpected Query.String()\nresult: %v\nexpect: %v\n", s, expected)
	}

	q = q.Ancestor(nil)
	s = q.String()
	expected = fmt.Sprintf("SELECT * FROM Wiki WHERE section=\"golang\" AND public=true ORDER BY title ASC, updated DESC")
	if s != expected {
		t.Errorf("Unexpected Query.String()\nresult: %v\nexpect: %v\n", s, expected)
	}
}
