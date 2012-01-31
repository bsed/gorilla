// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gorilla/rpc/json provides a codec for JSON-RPC over HTTP services.

Check the gorilla/rpc documentation for more details:

	http://gorilla-web.appspot.com/pkg/gorilla/rpc

To register the codec in a RPC server:

	import (
		"http"
		"code.google.com/p/gorilla/rpc"
		"code.google.com/p/gorilla/rpc/json"
	)

	func init() {
		s := rpc.NewServer()
		s.RegisterCodec(json.NewCodec(), "application/json")
		// [...]
		http.Handle("/rpc", s)
	}

A codec is tied to a content type. In the example above, the server will use
the JSON codec for requests with "application/json" as the value for the
"Content-Type" header.
*/
package json
