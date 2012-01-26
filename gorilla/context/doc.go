// Copyright 2011 Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gorilla/context provides a container to store values for a request.

A context stores global variables used during a request. For example, a router
can set variables extracted from the URL and later application handlers can
access those values. There are several others common cases.

Here's the basic usage: first define the keys that you will need. The key
type is interface{} so a key can be of any type that supports equality.
Here we define a key using a custom int type to avoid name collisions:

	package foo

	type contextKey int

	const Key1 contextKey = 0

Then somewhere in the package set a variable in the context. Context variables
are bound to a http.Request object, so you need a request instance to set a
value:

	context.DefaultContext.Set(request, Key1, "bar")

The application can later access the variable using the same key you provided:

	func MyHandler(w http.ResponseWriter, r *http.Request) {
		// val is "bar".
		val = context.DefaultContext.Get(r, foo.Key1)

		// ...
	}

You can store any type in the request context, because it accepts and returns
interface{}. To enforce a given type, a good idea is to make the key private
and wrap the getter and setter to accept and return values of a specific type:

	type contextKey int

	const key1 contextKey = 0

	// GetKey1 returns a value for this package from the request context.
	func GetKey1(request *http.Request) SomeType {
		rv := context.DefaultContext.Get(request, key1)
		if rv != nil {
			return rv.(SomeType)
		}
		return nil
	}

	// SetKey1 sets a value for this package in the request context.
	func SetKey1(request *http.Request, val SomeType) {
		context.DefaultContext.Set(request, key1, val)
	}

A context must be cleared at the end of a request, to remove all values
that were stored. This is done in a http.Handler, after a request was served.
Just call Clear() passing the request:

	context.DefaultContext.Clear(request)

The package gorilla/mux clears the default context, so if you are using the
default handler from there you don't need to do anything: context variables
will be deleted at the end of a request.
*/
package context
