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
func newIterator(c appengine.Context, q *Query, o *QueryOptions) *Iterator {
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
	// Query doesn't know about context so we must set app and namespace.
	req.App = proto.String(c.FullyQualifiedAppID())
	if err := c.Call("datastore_v3", "RunQuery", &req, &res, nil); err != nil {
		return &Iterator{err: err}
	}
	return &Iterator{
		c:      c,
		query:  q,
		res:    res,
		limit:  *req.Limit,
		offset: *req.Offset,
	}
}

// Iterator is the result of running a query.
type Iterator struct {
	c      appengine.Context
	query  *Query
	res    pb.QueryResult
	limit  int32
	offset int32
	pos    int
	err    os.Error
}

// Next returns the key of the next result. When there are no more results,
// Done is returned as the error.
// If the query is not keys only, it also loads the entity
// stored for that key into the struct pointer or Map dst, with the same
// semantics and possible errors as for the Get function.
// If the query is keys only, it is valid to pass a nil interface{} for dst.
func (q *Iterator) Next(dst interface{}) (*Key, os.Error) {
	k, e, err := q.next()
	if err != nil || e == nil {
		return k, err
	}
	q.pos += 1
	return k, loadEntity(dst, e)
}

// Cursor returns the query cursor positioned after the last query result.
func (q *Iterator) Cursor() *Cursor {
	if err := q.nextBatch(); err != nil {
		q.err = err
	}
	if q.res.CompiledCursor != nil {
		return &Cursor{q.res.CompiledCursor}
	}
	return nil
}

// cursor returns the cursor that points to the result at the given index
// for the current query.
func (q *Iterator) CursorAt(index int) *Cursor {
	// TODO: validate index.
	// TODO: don't need RPC in all cases.
	options := &QueryOptions{
		limit:    0,
		offset:   index,
		keysOnly: true,
		compile:  true,
	}
	return q.query.Run(q.c, options).Cursor()
}

// Private methods ------------------------------------------------------------

func (q *Iterator) next() (*Key, *pb.EntityProto, os.Error) {
	if q.err != nil {
		return nil, nil, q.err
	}
	if err := q.nextBatch(); err != nil {
		q.err = err
		return nil, nil, err
	}
	// Pop the EntityProto from the front of q.res.Result and
	// extract its key.
	var e *pb.EntityProto
	e, q.res.Result = q.res.Result[0], q.res.Result[1:]
	if e.Key == nil {
		return nil, nil, os.NewError("datastore: internal error: server did not return a key")
	}
	k, err := protoToKey(e.Key)
	if err != nil || k.Incomplete() {
		return nil, nil, os.NewError("datastore: internal error: server returned an invalid key")
	}
	if proto.GetBool(q.res.KeysOnly) {
		return k, nil, nil
	}
	return k, e, nil
}

// nextBatch issues datastore_v3/Next RPCs as necessary.
func (q *Iterator) nextBatch() os.Error {
	for len(q.res.Result) == 0 {
		if !proto.GetBool(q.res.MoreResults) {
			q.err = Done
			return q.err
		}
		q.offset -= proto.GetInt32(q.res.SkippedResults)
		if q.offset < 0 {
			q.offset = 0
		}
		if err := callNext(q.c, &q.res, q.offset, q.limit, zeroLimitMeansUnlimited); err != nil {
			q.err = err
			return q.err
		}
		// For an Iterator, a zero limit means unlimited.
		if q.limit == 0 {
			continue
		}
		q.limit -= int32(len(q.res.Result))
		if q.limit > 0 {
			continue
		}
		q.limit = 0
		if proto.GetBool(q.res.MoreResults) {
			q.err = os.NewError("datastore: internal error: limit exhausted but more_results is true")
			return q.err
		}
	}
	return nil
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

// advance returns a new Cursor advanced by the given offset.
func (c *Cursor) advance(ctx appengine.Context, query *Query, offset int) *Cursor {
	options := &QueryOptions{
		limit:       0,
		offset:      offset,
		keysOnly:    true,
		compile:     true,
		startCursor: c,
	}
	return query.Run(ctx, options).Cursor()
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
