// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package requestdata

import (
	"net/http"
	"sync"
)

// Original implementation by Brad Fitzpatrick:
// http://groups.google.com/group/golang-nuts/msg/e2d679d303aa5d53

var (
	mutex sync.Mutex
	data  = make(map[*http.Request]map[interface{}]interface{})
)

// Set stores the value for a given key in a given request.
func Set(r *http.Request, key, val interface{}) {
	mutex.Lock()
	if data[r] == nil {
		data[r] = make(map[interface{}]interface{})
	}
	data[r][key] = val
	mutex.Unlock()
}

// Get returns the value stored for a given key in a given request.
func Get(r *http.Request, key interface{}) interface{} {
	mutex.Lock()
	defer mutex.Unlock()
	if data[r] != nil {
		return data[r][key]
	}
	return nil
}

// Delete removes the value stored for a given key in a given request.
func Delete(r *http.Request, key interface{}) {
	mutex.Lock()
	if data[r] != nil {
		delete(data[r], key)
	}
	mutex.Unlock()
}

// Clear removes all values stored for a given request.
//
// This is usually called by a handler wrapper to clean up request
// variables at the end of a request lifetime. See ClearHandler().
func Clear(r *http.Request) {
	mutex.Lock()
	delete(data, r)
	mutex.Unlock()
}

// ClearAll removes all values stored for all requests.
//
// This is not normally used but it is here for completeness.
func ClearAll() {
	mutex.Lock()
	data = make(map[*http.Request]map[interface{}]interface{})
	mutex.Unlock()
}

// ClearHandler wraps an http.Handler and clears request values at the end
// of a request lifetime.
func ClearHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer Clear(r)
		h.ServeHTTP(w, r)
	})
}
