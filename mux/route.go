// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Route stores information to match a request and build URLs.
type Route struct {
	// Reference to the router where the route was registered.
	router *Router
	// Request handler for the route.
	handler http.Handler
	// List of matchers.
	matchers []matcher
	// Manager for the variables from host and path.
	regexp *routeRegexpGroup
	// If true, when the path pattern is "/path/", accessing "/path" will
	// redirect to the former and vice-versa.
	strictSlash bool
	// The name used to build URLs.
	name string
	// Error resulted from building a route.
	err error
}

// Match matches the route against the request.
func (r *Route) Match(req *http.Request, match *RouteMatch) bool {
	return r.match(req, match)
}

// match matches the route against the request.
func (r *Route) match(req *http.Request, match *RouteMatch) bool {
	// Match everything.
	for _, m := range r.matchers {
		if matched := m.match(req, match); !matched {
			return false
		}
	}
	// Yay, we have a match. Let's collect some info about it.
	if match.Route == nil {
		match.Route = r
	}
	if match.Handler == nil {
		match.Handler = r.handler
	}
	if match.Vars == nil {
		match.Vars = make(map[string]string)
	}
	// Set variables.
	if r.regexp != nil {
		r.regexp.setMatch(req, match, r)
	}
	return true
}

// ----------------------------------------------------------------------------
// Route attributes
// ----------------------------------------------------------------------------

// GetError returns an error resulted from building the route, if any.
func (r *Route) GetError() error {
	return r.err
}

// Handler --------------------------------------------------------------------

// Handler sets a handler for the route.
func (r *Route) Handler(handler http.Handler) *Route {
	r.handler = handler
	return r
}

// HandlerFunc sets a handler function for the route.
func (r *Route) HandlerFunc(f func(http.ResponseWriter, *http.Request)) *Route {
	r.handler = http.HandlerFunc(f)
	return r
}

// GetHandler returns the handler for the route, if any.
func (r *Route) GetHandler() http.Handler {
	return r.handler
}

// Name -----------------------------------------------------------------------

// Name sets the name for the route, used to build URLs.
// If the name was registered already it will be overwritten.
func (r *Route) Name(name string) *Route {
	if r.name != "" {
		r.err = fmt.Errorf("mux: route already has name %q, can't set %q",
			r.name, name)
		return r
	}
	if r.router == nil {
		// During tests router is not always set.
		r.router = new(Router)
	}
	r.name = name
	r.router.getNamedRoutes()[name] = r
	return r
}

// GetName returns the name for the route, if any.
func (r *Route) GetName() string {
	return r.name
}

// ----------------------------------------------------------------------------
// Matchers
// ----------------------------------------------------------------------------

// matcher types try to match a request.
type matcher interface {
	match(*http.Request, *RouteMatch) bool
}

// addMatcher adds a matcher to the route.
func (r *Route) addMatcher(m matcher) *Route {
	r.matchers = append(r.matchers, m)
	return r
}

// addRegexpMatcher adds a host or path matcher and builder to a route.
func (r *Route) addRegexpMatcher(tpl string, matchHost, matchPrefix bool) error {
	if r.regexp == nil {
		r.regexp = new(routeRegexpGroup)
	}
	rr, err := newRouteRegexp(tpl, matchHost, matchPrefix, r.strictSlash)
	if err != nil {
		return err
	}
	if matchHost {
		if r.regexp.path != nil {
			if err = uniqueVars(rr.varsN, r.regexp.path.varsN); err != nil {
				return err
			}
		}
		r.regexp.host = rr
	} else {
		if r.regexp.host != nil {
			if err = uniqueVars(rr.varsN, r.regexp.host.varsN); err != nil {
				return err
			}
		}
		r.regexp.path = rr
	}
	r.addMatcher(rr)
	return nil
}

// Headers --------------------------------------------------------------------

// headerMatcher matches the request against header values.
type headerMatcher map[string]string

func (m headerMatcher) match(r *http.Request, match *RouteMatch) bool {
	return matchMap(m, r.Header, true)
}

// Headers adds a matcher for header values.
// It accepts a sequence of key/value pairs to be matched. For example:
//
//     r := new(mux.Router)
//     r.Headers("Content-Type", "application/json",
//               "X-Requested-With", "XMLHttpRequest")
//
// The above route will only match if both request header values match.
//
// It the value is an empty string, it will match any value if the key is set.
func (r *Route) Headers(pairs ...string) *Route {
	if len(pairs) == 0 || r.err != nil {
		return r
	}
	headers, err := mapFromPairs(pairs...)
	if err != nil {
		r.err = err
		return r
	}
	return r.addMatcher(headerMatcher(headers))
}

// Host -----------------------------------------------------------------------

// Host adds a matcher for the URL host.
// It accepts a template with zero or more URL variables enclosed by {}.
// Variables can define an optional regexp pattern to me matched:
//
// - {name} matches anything until the next dot.
//
// - {name:pattern} matches the given regexp pattern.
//
// For example:
//
//     r := new(mux.Router)
//     r.Host("www.domain.com")
//     r.Host("{subdomain}.domain.com")
//     r.Host("{subdomain:[a-z]+}.domain.com")
//
// Variable names must be unique in a given route. They can be retrieved
// calling mux.Vars(request).
func (r *Route) Host(tpl string) *Route {
	if r.err == nil {
		r.err = r.addRegexpMatcher(tpl, true, false)
	}
	return r
}

// MatcherFunc ----------------------------------------------------------------

// MatcherFunc is the function signature used by custom matchers.
type MatcherFunc func(*http.Request, *RouteMatch) bool

func (m MatcherFunc) match(r *http.Request, match *RouteMatch) bool {
	return m(r, match)
}

// MatcherFunc adds a custom function to be used as request matcher.
func (r *Route) MatcherFunc(f MatcherFunc) *Route {
	return r.addMatcher(f)
}

// Methods --------------------------------------------------------------------

// methodMatcher matches the request against HTTP methods.
type methodMatcher []string

func (m methodMatcher) match(r *http.Request, match *RouteMatch) bool {
	return matchInArray(m, r.Method)
}

// Methods adds a matcher for HTTP methods.
// It accepts a sequence of one or more methods to be matched, e.g.:
// "GET", "POST", "PUT".
func (r *Route) Methods(methods ...string) *Route {
	if len(methods) == 0 || r.err != nil {
		return r
	}
	for k, v := range methods {
		methods[k] = strings.ToUpper(v)
	}
	return r.addMatcher(methodMatcher(methods))
}

// Path -----------------------------------------------------------------------

// Path adds a matcher for the URL path.
// It accepts a template with zero or more URL variables enclosed by {}.
// Variables can define an optional regexp pattern to me matched:
//
// - {name} matches anything until the next slash.
//
// - {name:pattern} matches the given regexp pattern.
//
// For example:
//
//     r := new(mux.Router)
//     r.Path("/products/").Handler(ProductsHandler)
//     r.Path("/products/{key}").Handler(ProductsHandler)
//     r.Path("/articles/{category}/{id:[0-9]+}").
//       Handler(ArticleHandler)
//
// Variable names must be unique in a given route. They can be retrieved
// calling mux.Vars(request).
func (r *Route) Path(tpl string) *Route {
	if r.err == nil {
		r.err = r.addRegexpMatcher(tpl, false, false)
	}
	return r
}

// PathPrefix -----------------------------------------------------------------

// PathPrefix adds a matcher for the URL path prefix.
func (r *Route) PathPrefix(tpl string) *Route {
	if r.err == nil {
		r.err = r.addRegexpMatcher(tpl, false, true)
	}
	return r
}

// Query ----------------------------------------------------------------------

// queryMatcher matches the request against URL queries.
type queryMatcher map[string]string

func (m queryMatcher) match(r *http.Request, match *RouteMatch) bool {
	return matchMap(m, r.URL.Query(), false)
}

// Queries adds a matcher for URL query values.
// It accepts a sequence of key/value pairs. For example:
//
//     r := new(mux.Router)
//     r.Queries("foo", "bar", "baz", "ding")
//
// The above route will only match if the URL contains the defined queries
// values, e.g.: ?foo=bar&baz=ding.
//
// It the value is an empty string, it will match any value if the key is set.
func (r *Route) Queries(pairs ...string) *Route {
	if len(pairs) == 0 || r.err != nil {
		return r
	}
	queries, err := mapFromPairs(pairs...)
	if err != nil {
		r.err = err
		return r
	}
	return r.addMatcher(queryMatcher(queries))
}

// Schemes --------------------------------------------------------------------

// schemeMatcher matches the request against URL schemes.
type schemeMatcher []string

func (m schemeMatcher) match(r *http.Request, match *RouteMatch) bool {
	return matchInArray(m, r.URL.Scheme)
}

// Schemes adds a matcher for URL schemes.
// It accepts a sequence schemes to be matched, e.g.: "http", "https".
func (r *Route) Schemes(schemes ...string) *Route {
	if len(schemes) == 0 || r.err != nil {
		return r
	}
	for k, v := range schemes {
		schemes[k] = strings.ToLower(v)
	}
	return r.addMatcher(schemeMatcher(schemes))
}

// Subrouter ------------------------------------------------------------------

// Subrouter creates a subrouter for the route.
//
// It will test the inner routes only if the parent route matched. For example:
//
//     r := new(mux.Router)
//     subrouter := r.Host("www.domain.com").Subrouter()
//     subrouter.HandleFunc("/products/", ProductsHandler)
//     subrouter.HandleFunc("/products/{key}", ProductHandler)
//     subrouter.HandleFunc("/articles/{category}/{id:[0-9]+}"),
//                          ArticleHandler)
//
// In this example, the routes registered in the subrouter won't be tested
// if the host doesn't match.
func (r *Route) Subrouter() *Router {
	if r.router == nil {
		// During tests router is not always set.
		r.router = new(Router)
	}
	router := &Router{
		namedRoutes: r.router.getNamedRoutes(),
		regexp:      copyRouteRegexpGroup(r.regexp),
	}
	r.addMatcher(router)
	return router
}

// ----------------------------------------------------------------------------
// URL building
// ----------------------------------------------------------------------------

// URL builds a URL for the route.
//
// It accepts a sequence of key/value pairs for the route variables. For
// example, given this route:
//
//     r := new(mux.Router)
//     r.HandleFunc("/articles/{category}/{id:[0-9]+}", ArticleHandler).
//       Name("article")
//
// ...a URL for it can be built using:
//
//     url := r.NamedRoute("article").URL("category", "technology",
//                                        "id", "42")
//
// ...which will return an url.URL with the following path:
//
//     "/articles/technology/42"
//
// This also works for host variables:
//
//     r := new(mux.Router)
//     r.Host("{subdomain}.domain.com").
//       HandleFunc("/articles/{category}/{id:[0-9]+}", ArticleHandler).
//       Name("article")
//
//     // url.String() will be "http://news.domain.com/articles/technology/42"
//     url := r.NamedRoute("article").URL("subdomain", "news",
//                                        "category", "technology",
//                                        "id", "42")
//
// All variable names defined in the route are required, and their values must
// conform to the corresponding patterns, if any.
func (r *Route) URL(pairs ...string) (*url.URL, error) {
	if r.regexp == nil {
		return nil, errors.New("mux: route doesn't have a host or path.")
	}
	var scheme, host, path string
	var err error
	if r.regexp.host != nil {
		// Set a default scheme.
		scheme = "http"
		if host, err = r.regexp.host.url(pairs...); err != nil {
			return nil, err
		}
	}
	if r.regexp.path != nil {
		if path, err = r.regexp.path.url(pairs...); err != nil {
			return nil, err
		}
	}
	return &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}, nil
}

// URLHost builds the host part of the URL for a route. See Route.URL().
//
// The route must have a host defined.
func (r *Route) URLHost(pairs ...string) (*url.URL, error) {
	if r.regexp == nil || r.regexp.host == nil {
		return nil, errors.New("mux: route doesn't have a host.")
	}
	host, err := r.regexp.host.url(pairs...)
	if err != nil {
		return nil, err
	}
	return &url.URL{
		Scheme: "http",
		Host:   host,
	}, nil
}

// URLPath builds the path part of the URL for a route. See Route.URL().
//
// The route must have a path defined.
func (r *Route) URLPath(pairs ...string) (*url.URL, error) {
	if r.regexp == nil || r.regexp.path == nil {
		return nil, errors.New("mux: route doesn't have a path.")
	}
	path, err := r.regexp.path.url(pairs...)
	if err != nil {
		return nil, err
	}
	return &url.URL{
		Path: path,
	}, nil
}
