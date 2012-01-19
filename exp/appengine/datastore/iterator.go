// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package datastore

import (
	"encoding/base64"
	"os"
	"strings"

	"appengine"
	"goprotobuf.googlecode.com/hg/proto"

	pb "appengine_internal/datastore"
)

// ----------------------------------------------------------------------------
// Iterator
// ----------------------------------------------------------------------------

func newIterator(c appengine.Context, q *Query, o *QueryOptions, method string) *Iterator {
	// TODO: zero limit policy
	var req pb.Query
	var res pb.QueryResult
	if err := q.toProto(&req); err != nil {
		return &Iterator{err: err}
	}
	if err := o.toProto(&req); err != nil {
		return &Iterator{err: err}
	}
	if err := c.Call("datastore_v3", method, &req, &res, nil); err != nil {
		return &Iterator{err: err}
	}
	return &Iterator{
		c:   c,
		res: &res,
	}
}

type Iterator struct {
	c   appengine.Context
	res *pb.QueryResult
	err os.Error
}

// ----------------------------------------------------------------------------
// Cursor
// ----------------------------------------------------------------------------

// TODO factory function?

// Cursor represents a compiled query cursor.
type Cursor struct {
	compiledCursor *pb.CompiledCursor
}

// String returns a compact representation of the cursor suitable for
// debugging.
func (c *Cursor) String() string {
	if c.compiledCursor != nil {
		return c.compiledCursor.String()
	}
	return ""
}

// Encode returns an opaque representation of the cursor suitable for use in
// HTML and URLs. This is compatible with the Python and Java runtimes.
func (c *Cursor) Encode() string {
	if c.compiledCursor != nil {
		if b, err := proto.Marshal(c.compiledCursor); err == nil {
			// Trailing padding is stripped.
			return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
		}
	}
	// We don't return the error to follow Key.Encode which only
	// returns a string. It is unlikely to happen anyway.
	return ""
}

// DecodeCursor decodes a cursor from the opaque representation returned by
// Cursor.Encode.
func DecodeCursor(encoded string) (*Cursor, os.Error) {
	// Re-add padding.
	if m := len(encoded) % 4; m != 0 {
		encoded += strings.Repeat("=", 4-m)
	}
	b, err := base64.URLEncoding.DecodeString(encoded)
	if err == nil {
		var c pb.CompiledCursor
		err = proto.Unmarshal(b, &c)
		if err == nil {
			return &Cursor{&c}, nil
		}
	}
	return nil, err
}
