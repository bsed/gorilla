// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reverse

import (
	"net/http"
	"testing"
)

func testMatcher(t *testing.T, name string, m Matcher, r *http.Request, expect bool) {
	result := m.Match(r)
	if result != expect {
		t.Errorf("%s: got %v, expected %v", name, result, expect)
	}
}

func TestHost(t *testing.T) {
	type test struct {
		host   string
		rURL   string
		expect bool
	}
	tests := []test{
		{"domain.com", "http://domain.com", true},
		{"domain.com", "http://other.com", false},
	}
	for _, v := range tests {
		r, err := http.NewRequest("GET", v.rURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		testMatcher(t, "Host", NewHost(v.host), r, v.expect)
	}
}

func TestMethod(t *testing.T) {
	type test struct {
		methods []string
		rMethod string
		expect  bool
	}
	tests := []test{
		{[]string{"GET", "POST"}, "GET", true},
		{[]string{"GET", "POST"}, "POST", true},
		{[]string{"get", "post"}, "GET", true},
		{[]string{"get", "post"}, "POST", true},
		{[]string{"POST", "PUT"}, "GET", false},
	}
	for _, v := range tests {
		r, err := http.NewRequest(v.rMethod, "http://domain.com", nil)
		if err != nil {
			t.Fatal(err)
		}
		testMatcher(t, "Method", NewMethod(v.methods), r, v.expect)
	}
}

func TestScheme(t *testing.T) {
	type test struct {
		schemes []string
		rURL    string
		expect  bool
	}
	tests := []test{
		{[]string{"http", "https"}, "http://domain.com", true},
		{[]string{"http", "https"}, "https://domain.com", true},
		{[]string{"https"}, "http://domain.com", false},
	}
	for _, v := range tests {
		r, err := http.NewRequest("GET", v.rURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		testMatcher(t, "Scheme", NewScheme(v.schemes), r, v.expect)
	}
}
