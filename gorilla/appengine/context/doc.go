// Copyright 2011 Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gorilla/appengine/context provides a convenience wrapper to serve as
base for extended contexts.

It wraps the concrete implementation of appengine.Context from the App Engine
SDK to allow it to be embedded in custom contexts. For example, to create your
own context:

	import (
		"http"
		"appengine"
		"gorilla.googlecode.com/hg/gorilla/appengine/context"
	)

	type Context struct {
		context.Context
	}

	func NewContext(r *http.Request) *Context {
		return &Context{
			context.Context: context.Context{appengine.NewContext(r)},
		}
	}

You can now add custom fields and methods to Context and pass it to all
functions that accept appengine.Context.

This is convenient because we can attach to the context app-specific types
such as registries or whatever the app needs. And the implementation of the
appengine.Context interface remains encapsulated in a single place.
*/
package context
