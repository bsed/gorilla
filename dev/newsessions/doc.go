// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

/*
API in a nutshell:

	import (
		"net/http"
		"code.google.com/p/gorilla/sessions"
	)

	var store = NewCookieStore([]byte("something-very-secret"))

	func MyHandler(w http.ResponseWriter, r *http.Request) {
		// Get a session. We're ignoring the error from decoding an existing
		// session: Get() always returns a session anyway, even if empty.
		s, _ := store.Get(r, "session-name")
		// Set some session values.
		s.Values["foo"] = "bar"
		s.Values[42] = 43
		// Save it.
		sessions.Save(r, w)
	}
*/
