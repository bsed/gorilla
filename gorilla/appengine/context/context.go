// Copyright 2011 Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"os"
	"appengine"
	"appengine_internal"
)

// Context is a convenience to create extended appengine.Context types.
type Context struct {
	// The real App Engine context.
	Context appengine.Context
}

// Debugf formats its arguments according to the format, analogous to
// fmt.Printf, and records the text as a log message at Debug level.
func (c *Context) Debugf(format string, args ...interface{}) {
	c.Context.Debugf(format, args...)
}

// Infof is like Debugf, but at Info level.
func (c *Context) Infof(format string, args ...interface{}) {
	c.Context.Infof(format, args...)
}

// Warningf is like Debugf, but at Warning level.
func (c *Context) Warningf(format string, args ...interface{}) {
	c.Context.Warningf(format, args...)
}

// Errorf is like Debugf, but at Error level.
func (c *Context) Errorf(format string, args ...interface{}) {
	c.Context.Errorf(format, args...)
}

// Criticalf is like Debugf, but at Critical level.
func (c *Context) Criticalf(format string, args ...interface{}) {
	c.Context.Criticalf(format, args...)
}

// AppID is deprecated. Use the AppID function instead.
func (c *Context) AppID() string {
	return c.Context.AppID()
}

// The remaining methods are for internal use only.
// Developer-facing APIs wrap these methods to provide a more friendly API.

// For internal use only.
func (c *Context) Call(service, method string, in, out interface{}, opts *appengine_internal.CallOptions) os.Error {
	return c.Context.Call(service, method, in, out, opts)
}

// For internal use only.
func (c *Context) FullyQualifiedAppID() string {
	return c.Context.FullyQualifiedAppID()
}

// For internal use only.
func (c *Context) Request() interface{} {
	return c.Context.Request()
}
