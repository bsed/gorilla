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

// zeroLimitPolicy defines how to interpret a zero query/cursor limit. In some
// contexts, it means an unlimited query (to follow Go's idiom of a zero value
// being a useful default value). In other contexts, it means a literal zero,
// such as when issuing a query count, no actual entity data is wanted, only
// the number of skipped results.
type zeroLimitPolicy int

const (
	zeroLimitMeansUnlimited zeroLimitPolicy = iota
	zeroLimitMeansZero
)

// Done is returned when a query iteration has completed.
var Done = os.NewError("datastore: query has no more results")

// callNext issues a datastore_v3/Next RPC to advance a cursor, such as that
// returned by a query with more results.
func callNext(c appengine.Context, res *pb.QueryResult, offset, limit int32, zlp zeroLimitPolicy) os.Error {
	if res.Cursor == nil {
		return os.NewError("datastore: internal error: server did not return a cursor")
	}
	// TODO: should I eventually call datastore_v3/DeleteCursor on the cursor?
	req := &pb.NextRequest{
		Cursor: res.Cursor,
		Offset: proto.Int32(offset),
	}
	if limit != 0 || zlp == zeroLimitMeansZero {
		req.Count = proto.Int32(limit)
	}
	if res.CompiledCursor != nil {
		req.Compile = proto.Bool(true)
	}
	res.Reset()
	return c.Call("datastore_v3", "Next", req, res, nil)
}

// newIterator returns a new Iterator.
func newIterator(c appengine.Context, q *Query, o *QueryOptions, method string) *Iterator {
	// TODO: get namespace from context once it supports it.
	// TODO: zero limit policy
	var req pb.Query
	var res pb.QueryResult
	if err := q.toProto(&req); err != nil {
		return &Iterator{err: err}
	}
	if err := o.toProto(&req); err != nil {
		return &Iterator{err: err}
	}
	// Query doesn't know about context so we must set the app field.
	req.App = proto.String(c.FullyQualifiedAppID())
	if err := c.Call("datastore_v3", method, &req, &res, nil); err != nil {
		return &Iterator{err: err}
	}
	return &Iterator{
		c:      c,
		res:    res,
		limit:  *req.Limit,
		offset: *req.Offset,
	}
}

// Iterator is the result of running a query.
type Iterator struct {
	c      appengine.Context
	res    pb.QueryResult
	limit  int32
	offset int32
	err    os.Error
}

// Next returns the key of the next result. When there are no more results,
// Done is returned as the error.
// If the query is not keys only, it also loads the entity
// stored for that key into the struct pointer or Map dst, with the same
// semantics and possible errors as for the Get function.
// If the query is keys only, it is valid to pass a nil interface{} for dst.
func (t *Iterator) Next(dst interface{}) (*Key, os.Error) {
	k, e, err := t.next()
	if err != nil || e == nil {
		return k, err
	}
	return k, loadEntity(dst, e)
}

func (t *Iterator) next() (*Key, *pb.EntityProto, os.Error) {
	if t.err != nil {
		return nil, nil, t.err
	}

	// Issue datastore_v3/Next RPCs as necessary.
	for len(t.res.Result) == 0 {
		if !proto.GetBool(t.res.MoreResults) {
			t.err = Done
			return nil, nil, t.err
		}
		t.offset -= proto.GetInt32(t.res.SkippedResults)
		if t.offset < 0 {
			t.offset = 0
		}
		if err := callNext(t.c, &t.res, t.offset, t.limit, zeroLimitMeansUnlimited); err != nil {
			t.err = err
			return nil, nil, t.err
		}
		// For an Iterator, a zero limit means unlimited.
		if t.limit == 0 {
			continue
		}
		t.limit -= int32(len(t.res.Result))
		if t.limit > 0 {
			continue
		}
		t.limit = 0
		if proto.GetBool(t.res.MoreResults) {
			t.err = os.NewError("datastore: internal error: limit exhausted but more_results is true")
			return nil, nil, t.err
		}
	}

	// Pop the EntityProto from the front of t.res.Result and
	// extract its key.
	var e *pb.EntityProto
	e, t.res.Result = t.res.Result[0], t.res.Result[1:]
	if e.Key == nil {
		return nil, nil, os.NewError("datastore: internal error: server did not return a key")
	}
	k, err := protoToKey(e.Key)
	if err != nil || k.Incomplete() {
		return nil, nil, os.NewError("datastore: internal error: server returned an invalid key")
	}
	if proto.GetBool(t.res.KeysOnly) {
		return k, nil, nil
	}
	return k, e, nil
}

// ----------------------------------------------------------------------------
// Cursor
// ----------------------------------------------------------------------------

// Cursor represents a compiled query cursor.
type Cursor struct {
	compiledCursor *pb.CompiledCursor
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
