// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

/*
Package gorilla/appengine/datastore offers an alternative API to the
App Engine datastore.

This is a work in progress. The intention of this package is to feed the
discussion in the hope to push the standard App Engine API forward.

This API is similar to the standard one but it is not compatible, sometimes
in subtle ways. New supported features are:

	- Kindless Queries: http://goo.gl/11hei
	- Kindless Ancestor Queries: http://goo.gl/UyaRV
	- Query Cursors: http://goo.gl/eqUaj
	- Datastore Multitenancy: http://goo.gl/0wRJ8
*/
package datastore
