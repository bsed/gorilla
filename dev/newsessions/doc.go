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

	var cookieStore = NewCookieStore([]byte("something-very-secret"))

	func MyHandler(w http.ResponseWriter, r *http.Request) {
		if session, err := cookieStore.Get(r, "cookie-name"); err == nil {
			// Set a session value.
			session.Value["foo"] = "bar"
			// Save all sessions.
			sessions.Save(r, w)
		}
	}
*/
