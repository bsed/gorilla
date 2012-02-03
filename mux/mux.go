// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"fmt"
	"net/http"
	"path"

	"code.google.com/p/gorilla/context"
)

// ----------------------------------------------------------------------------
// Context
// ----------------------------------------------------------------------------

type RouteVars map[string]string

type contextKey int

const (
   varsKey contextKey = iota
   routeKey
)

// Vars returns the route variables for the current request, if any.
func Vars(r *http.Request) RouteVars {
	if rv := context.DefaultContext.Get(r, varsKey); rv != nil {
		return rv.(RouteVars)
	}
	return nil
}

// CurrentRoute returns the matched route for the current request, if any.
func CurrentRoute(r *http.Request) *Route {
	if rv := context.DefaultContext.Get(r, routeKey); rv != nil {
		return rv.(*Route)
	}
	return nil
}

func setVars(r *http.Request, val interface{}) {
	context.DefaultContext.Set(r, varsKey, val)
}

func setCurrentRoute(r *http.Request, val interface{}) {
	context.DefaultContext.Set(r, routeKey, val)
}

// ----------------------------------------------------------------------------
// Router
// ----------------------------------------------------------------------------

// Router registers routes to be matched and dispatches a handler.
//
// It implements the http.Handler interface, so it can be registered to serve
// requests:
//
//     var router = new(mux.Router)
//
//     func main() {
//         http.Handle("/", router)
//     }
//
// Or, for Google App Engine, register it in a init() function:
//
//     var router = new(mux.Router)
//
//     func init() {
//         http.Handle("/", router)
//     }
//
// This will send all incoming requests to the router.
type Router struct {
	// Routes by name, for URL building.
	NamedRoutes map[string]*Route
	// Configurable Handler to be used when no route matches.
	NotFoundHandler http.Handler
	// Routes to be matched, in order.
	routes []*Route
	// Reference to the root router, where named routes are stored.
	rootRouter *Router
	// See Route.strictSlash. This defines the default flag for new routes.
	strictSlash bool
	// Manager for the variables from host and path.
	regexp *routeRegexpGroup
}

// Match matches registered routes against the request.
func (r *Router) Match(req *http.Request, match *RouteMatch) bool {
	for _, route := range r.routes {
		if route.err == nil {
			if matched := route.Match(req, match); matched {
				setVars(req, match.Vars)
				setCurrentRoute(req, match.Route)
				return true
			}
		}
	}
	return false
}

// ServeHTTP dispatches the handler registered in the matched route.
//
// When there is a match, the route variables can be retrieved calling
// mux.Vars(request).
func (r *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// Clean path to canonical form and redirect.
	// (this comes from the http package)
	if p := cleanPath(request.URL.Path); p != request.URL.Path {
		writer.Header().Set("Location", p)
		writer.WriteHeader(http.StatusMovedPermanently)
		return
	}
	var match RouteMatch
	var handler http.Handler
	if matched := r.Match(request, &match); matched {
		handler = match.Handler
	}
	if handler == nil {
		if r.NotFoundHandler == nil {
			r.NotFoundHandler = http.NotFoundHandler()
		}
		handler = r.NotFoundHandler
	}
	defer context.DefaultContext.Clear(request)
	handler.ServeHTTP(writer, request)
}

// AddRoute registers a route in the router.
func (r *Router) AddRoute(route *Route) *Router {
	if r.routes == nil {
		r.routes = make([]*Route, 0)
	}
	route.router = r
	r.routes = append(r.routes, route)
	return r
}

// RedirectSlash defines the default RedirectSlash behavior for new routes.
//
// See Route.RedirectSlash.
func (r *Router) RedirectSlash(value bool) *Router {
	r.strictSlash = value
	return r
}

// root returns the root router, where named routes are stored.
func (r *Router) root() *Router {
	if r.rootRouter == nil {
		return r
	}
	return r.rootRouter
}

// Convenience route factories ------------------------------------------------

// NewRoute creates an empty route and registers it in the router.
func (r *Router) NewRoute() *Route {
	route := new(Route)
	route.strictSlash = r.strictSlash
	if r.regexp != nil {
		route.regexp = new(routeRegexpGroup)
		route.regexp.host = r.regexp.host
		route.regexp.path = r.regexp.path
	}
	r.AddRoute(route)
	return route
}

// Handle registers a new route and sets a path and handler.
//
// See also: Route.Handle().
func (r *Router) Handle(path string, handler http.Handler) *Route {
	return r.NewRoute().Handle(path, handler)
}

// HandleFunc registers a new route and sets a path and handler function.
//
// See also: Route.HandleFunc().
func (r *Router) HandleFunc(path string, handler func(http.ResponseWriter,
	*http.Request)) *Route {
	return r.NewRoute().HandleFunc(path, handler)
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// cleanPath returns the canonical path for p, eliminating . and .. elements.
//
// Extracted from the http package.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

// uniqueVars returns an error if two slices contain duplicated strings.
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

// stringMapFromPairs converts variadic string parameters to a string map.
func stringMapFromPairs(pairs ...string) (map[string]string, error) {
	length := len(pairs)
	if length%2 != 0 {
		return nil, fmt.Errorf("mux: parameters must be multiple of 2, got %v",
			pairs)
	}
	m := make(map[string]string, length/2)
	for i := 0; i < length; i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return m, nil
}

// matchInArray returns true if the given string value is in the array.
func matchInArray(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

// matchMap returns true if the given key/value pairs exist in a given map.
func matchMap(toCheck map[string]string, toMatch map[string][]string,
	canonicalKey bool) bool {
	for k, v := range toCheck {
		// Check if key exists.
		if canonicalKey {
			k = http.CanonicalHeaderKey(k)
		}
		if values := toMatch[k]; values == nil {
			return false
		} else if v != "" {
			// If value was defined as an empty string we only check that the
			// key exists. Otherwise we also check if the value exists.
			valueExists := false
			for _, value := range values {
				if v == value {
					valueExists = true
					break
				}
			}
			if !valueExists {
				return false
			}
		}
	}
	return true
}
