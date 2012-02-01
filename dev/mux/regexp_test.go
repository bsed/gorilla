// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"testing"
)

type regexpTest struct {
	tpl string
	url string
	res map[string]string
}

// ----------------------------------------------------------------------------
// Host regexp
// ----------------------------------------------------------------------------

var hostRegexpTests = []regexpTest{
	{
		tpl: "{foo:[a-z][a-z][a-z]}.{bar:[a-z][a-z][a-z]}.{baz:[a-z][a-z][a-z]}",
		url: "abc.def.ghi",
		res: map[string]string{"foo": "abc", "bar": "def", "baz": "ghi"},
	},
	{
		tpl: "{foo:[a-z][a-z][a-z]}.{bar:[a-z][a-z][a-z]}.{baz:[a-z][a-z][a-z]}",
		url: "a.b.c",
		res: nil,
	},
}

func TestHostRegexp(t *testing.T) {
	testRegexp(t, hostRegexpTests, "[^.]+", false, false)
}

// ----------------------------------------------------------------------------
// Path regexp
// ----------------------------------------------------------------------------

var pathRegexpTests = []regexpTest{
	{
		tpl: "/{foo:[0-9][0-9][0-9]}/{bar:[0-9][0-9][0-9]}/{baz:[0-9][0-9][0-9]}",
		url: "/123/456/789",
		res: map[string]string{"foo": "123", "bar": "456", "baz": "789"},
	},
	{
		tpl: "/{foo:[0-9][0-9][0-9]}/{bar:[0-9][0-9][0-9]}/{baz:[0-9][0-9][0-9]}",
		url: "/1/2/3",
		res: nil,
	},
}

func TestPathRegexp(t *testing.T) {
	testRegexp(t, pathRegexpTests, "[^/]+", false, false)
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// match matches a string and returns a map of variables.
func match(r *routeRegexp, s string) map[string]string {
	matches := r.regexp.FindStringSubmatch(s)
	if matches != nil {
		vars := make(map[string]string)
		for k, v := range r.varsN {
			// Skip first match which is the whole matched string.
			vars[v] = matches[k+1]
		}
		return vars
	}
	return nil
}

func isEqualStringMap(m1, m2 map[string]string) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v := range m1 {
		if v != m2[k] {
			return false
		}
	}
	return true
}

func testRegexp(t *testing.T, tests []regexpTest, defaultPattern string,
	matchPrefix, strictSlash bool) {
	for _, values := range tests {
		r, _ := newRouteRegexp(values.tpl, defaultPattern, matchPrefix,
			strictSlash)
		m := match(r, values.url)
		hasMatch := m != nil
		shouldMatch := values.res != nil
		if hasMatch != shouldMatch {
			msg := "Should not match"
			if shouldMatch {
				msg = "Should match"
			}
			t.Errorf("%v -- Path: %v URL: %v", msg, r.template, values.url)
		} else {
			if !isEqualStringMap(m, values.res) {
				t.Errorf("Result is not equal for %q: expected %v, got %v", values.tpl, values.res, m)
			}
		}
	}
}
