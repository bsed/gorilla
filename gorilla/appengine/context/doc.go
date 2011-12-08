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
		c *context.Context
	}

	func NewContext(r *http.Request) *Context {
		return &Context{
			c: *context.Context{appengine.NewContext(r)},
		}
	}

You can now add custom fields and methods to Context and pass it to all
functions that accept appengine.Context.

This is convenient because since we commonly pass an appengine.Context instance
as argument, we can use it to also pass app-specific types such as registries
or whatever the app needs.
*/
package context
