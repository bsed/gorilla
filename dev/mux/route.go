// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Matcher types try to match a request.
type Matcher interface {
	Match(*http.Request) (*RouteMatch, bool)
}

// RouteMatch stores the matched route, handler and variables.
type RouteMatch struct {
	Route   *Route
	Handler http.Handler
	Vars    map[string]string
}

// Route stores conditions to match a request.
type Route struct {
	handler     http.Handler
	hostRegexp  *routeRegexp
	pathRegexp  *routeRegexp
	matchers    []Matcher
	strictSlash bool
	err         ErrMulti
}

// Match matches this route against the request.
func (r *Route) Match(req *http.Request) (*RouteMatch, bool) {
	// Match host.
	var hostVars []string
	if r.hostRegexp != nil {
		hostVars = r.hostRegexp.regexp.FindStringSubmatch(req.URL.Host)
		if hostVars == nil {
			return nil, false
		}
	}
	// Match path.
	var pathVars []string
	if r.pathRegexp != nil {
		pathVars = r.pathRegexp.regexp.FindStringSubmatch(req.URL.Path)
		if pathVars == nil {
			return nil, false
		}
	}
	// Match extra matchers, including subroutes.
	var match *RouteMatch
	if r.matchers != nil {
		var ok bool
		for _, matcher := range r.matchers {
			if match, ok = matcher.Match(req); !ok {
				return nil, false
			} else if match != nil {
				break
			}
		}
	}

	// Yay, we have a match. Let's collect info about it.

	// Subrouter didn't return a match, so create one.
	if match == nil {
		match = &RouteMatch{
			Route:   r,
			Handler: r.handler,
			Vars:    make(map[string]string),
		}
	}
	// Store host variables.
	if hostVars != nil {
		for k, v := range r.hostRegexp.varsN {
			match.Vars[v] = hostVars[k+1]
		}
	}
	// Store path variables.
	if pathVars != nil {
		for k, v := range r.pathRegexp.varsN {
			match.Vars[v] = pathVars[k+1]
		}
		// Check if we should redirect.
		if r.strictSlash {
			p1 := strings.HasSuffix(req.URL.Path, "/")
			p2 := strings.HasSuffix(r.pathRegexp.template, "/")
			if p1 != p2 {
				ru, _ := url.Parse(req.URL.String())
				if p1 {
					ru.Path = ru.Path[:len(ru.Path)-1]
				} else {
					ru.Path += "/"
				}
				redirectURL := ru.String()
				match.Handler = http.RedirectHandler(redirectURL, 301)
			}
		}
	}
	// Done!
	return match, true
}

func (r *Route) Host(tpl string) *Route {
	if r.hostRegexp != nil {
		err := fmt.Errorf("mux: host already defined for %q", tpl)
		r.err = append(r.err, err)
		return r
	}
	hostRegexp, err := newRouteRegexp(tpl, "[^.]+", false, false)
	if err != nil {
		r.err = append(r.err, err)
		return r
	}
	if r.pathRegexp != nil {
		if err = uniqueVars(hostRegexp.varsN, r.pathRegexp.varsN); err != nil {
			r.err = append(r.err, err)
			return r
		}
	}
	r.hostRegexp = hostRegexp
	return r
}

func (r *Route) Path(tpl string) *Route {
	return r.path(tpl, false)
}

func (r *Route) PathPrefix(tpl string) *Route {
	return r.path(tpl, true)
}

// path adds a Path or PathPrefix matcher to the route.
func (r *Route) path(tpl string, matchPrefix bool) *Route {
	if r.pathRegexp != nil {
		err := fmt.Errorf("mux: path already defined for %q", tpl)
		r.err = append(r.err, err)
		return r
	}
	pathRegexp, err := newRouteRegexp(tpl, "[^/]+", matchPrefix, r.strictSlash)
	if err != nil {
		r.err = append(r.err, err)
		return r
	}
	if r.hostRegexp != nil {
		if err = uniqueVars(pathRegexp.varsN, r.hostRegexp.varsN); err != nil {
			r.err = append(r.err, err)
			return r
		}
	}
	r.pathRegexp = pathRegexp
	return r
}

func uniqueVars(s1, s2 []string) error {
	vars := make(map[string]bool)
	for _, s := range s1 {
		vars[s] = true
	}
	for _, s := range s2 {
		if vars[s] {
			return fmt.Errorf("mux: duplicated variable %q", s)
		}
	}
	return nil
}
