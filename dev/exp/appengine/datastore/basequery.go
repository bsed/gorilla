// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package datastore

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strings"

	"code.google.com/p/goprotobuf/proto"

	"appengine"
	pb "appengine_internal/datastore"
)

type queryOperator int

// Filter operators.
const (
	QueryOperatorLessThan queryOperator = iota
	QueryOperatorLessThanOrEqual
	QueryOperatorEqual
	QueryOperatorGreaterThanOrEqual
	QueryOperatorGreaterThan
)

var queryOperatorToProto = map[queryOperator]*pb.Query_Filter_Operator{
	QueryOperatorLessThan:           pb.NewQuery_Filter_Operator(pb.Query_Filter_LESS_THAN),
	QueryOperatorLessThanOrEqual:    pb.NewQuery_Filter_Operator(pb.Query_Filter_LESS_THAN_OR_EQUAL),
	QueryOperatorEqual:              pb.NewQuery_Filter_Operator(pb.Query_Filter_EQUAL),
	QueryOperatorGreaterThanOrEqual: pb.NewQuery_Filter_Operator(pb.Query_Filter_GREATER_THAN_OR_EQUAL),
	QueryOperatorGreaterThan:        pb.NewQuery_Filter_Operator(pb.Query_Filter_GREATER_THAN),
}

type queryDirection int

// Order directions.
const (
	QueryDirectionAscending queryDirection = iota
	QueryDirectionDescending
)

var queryDirectionToProto = map[queryDirection]*pb.Query_Order_Direction{
	QueryDirectionAscending:  pb.NewQuery_Order_Direction(pb.Query_Order_ASCENDING),
	QueryDirectionDescending: pb.NewQuery_Order_Direction(pb.Query_Order_DESCENDING),
}

// ----------------------------------------------------------------------------
// BaseQuery
// ----------------------------------------------------------------------------

// NewBaseQuery returns a new BaseQuery.
func NewBaseQuery() *BaseQuery {
	return &BaseQuery{pbq: new(pb.Query)}
}

// BaseQuery deals with protocol buffers so that Query doesn't have to.
type BaseQuery struct {
	pbq *pb.Query
	err error
}

// Clone returns a copy of the query.
func (q *BaseQuery) Clone() *BaseQuery {
	return &BaseQuery{pbq: &(*q.pbq), err: q.err}
}

// Namespace sets the namespace for the query.
//
// This is a temporary function to fill the gap until appengine.Context
// supports namespaces.
func (q *BaseQuery) Namespace(namespace string) *BaseQuery {
	if q.err == nil {
		if namespace == "" {
			q.pbq.NameSpace = nil
		} else {
			q.pbq.NameSpace = proto.String(namespace)
		}
	}
	return q
}

// Ancestor sets the ancestor filter for the query.
func (q *BaseQuery) Ancestor(key *Key) *BaseQuery {
	if q.err == nil {
		if key == nil {
			q.pbq.Ancestor = nil
		} else {
			if key.Incomplete() {
				q.err = errors.New("datastore: incomplete query ancestor key")
			} else {
				q.pbq.Ancestor = keyToProto(key)
			}
		}
	}
	return q
}

// Kind sets the entity kind for the query.
func (q *BaseQuery) Kind(kind string) *BaseQuery {
	if q.err == nil {
		if kind == "" {
			q.pbq.Kind = nil
		} else {
			q.pbq.Kind = proto.String(kind)
		}
	}
	return q
}

// Filter adds a field-based filter to the query.
func (q *BaseQuery) Filter(property string, operator queryOperator,
	value interface{}) *BaseQuery {
	if q.err == nil {
		var p *pb.Property
		p, q.err = valueToProto(property, value, false)
		if q.err == nil {
			q.pbq.Filter = append(q.pbq.Filter, &pb.Query_Filter{
				Op:       queryOperatorToProto[operator],
				Property: []*pb.Property{p},
			})
		}
	}
	return q
}

// Order adds a field-based sort to the query.
func (q *BaseQuery) Order(property string, direction queryDirection) *BaseQuery {
	if q.err == nil {
		q.pbq.Order = append(q.pbq.Order, &pb.Query_Order{
			Property:  proto.String(property),
			Direction: queryDirectionToProto[direction],
		})
	}
	return q
}

// Limit sets the maximum number of keys/entities to return.
// A zero value means unlimited. A negative value is invalid.
func (q *BaseQuery) Limit(limit int) *BaseQuery {
	if q.err == nil {
		if limit == 0 {
			q.pbq.Limit = nil
		} else if q.err = validateInt32(limit, "limit"); q.err == nil {
			q.pbq.Limit = proto.Int32(int32(limit))
		}
	}
	return q
}

// Offset sets how many keys to skip over before returning results.
// A negative value is invalid.
func (q *BaseQuery) Offset(offset int) *BaseQuery {
	if q.err == nil {
		if offset == 0 {
			q.pbq.Offset = nil
		} else if q.err = validateInt32(offset, "offset"); q.err == nil {
			q.pbq.Offset = proto.Int32(int32(offset))
		}
	}
	return q
}

// KeysOnly configures the query to return keys, instead of keys and entities.
func (q *BaseQuery) KeysOnly(keysOnly bool) *BaseQuery {
	if q.err == nil {
		q.pbq.KeysOnly = proto.Bool(keysOnly)
		q.pbq.RequirePerfectPlan = proto.Bool(keysOnly)
	}
	return q
}

// Cursor sets the cursor position to start the query.
func (q *BaseQuery) Cursor(cursor *Cursor) *BaseQuery {
	if q.err == nil {
		if cursor == nil {
			q.pbq.CompiledCursor = nil
		} else if cursor.compiledCursor == nil {
			q.err = errors.New("datastore: empty start cursor")
		} else {
			q.pbq.CompiledCursor = cursor.compiledCursor
		}
	}
	if q.err == nil {
		q.pbq.Compile = proto.Bool(true)
	}
	return q
}

// EndCursor sets the cursor position to end the query.
func (q *BaseQuery) EndCursor(cursor *Cursor) *BaseQuery {
	if q.err == nil {
		if cursor == nil {
			q.pbq.EndCompiledCursor = nil
		} else if cursor.compiledCursor == nil {
			q.err = errors.New("datastore: empty end cursor")
		} else {
			q.pbq.EndCompiledCursor = cursor.compiledCursor
		}
	}
	if q.err == nil {
		q.pbq.Compile = proto.Bool(true)
	}
	return q
}

// toProto converts the query to a protocol buffer.
//
// The zeroLimitMeansZero flag defines how to interpret a zero query/cursor
// limit. In some contexts, it means an unlimited query (to follow Go's idiom
// of a zero value being a useful default value). In other contexts, it means
// a literal zero, such as when issuing a query count, no actual entity data
// is wanted, only the number of skipped results.
func (q *BaseQuery) toProto(pbq *pb.Query, zeroLimitMeansZero bool) error {
	if q.err != nil {
		return q.err
	}
	if zeroLimitMeansZero && pbq.Limit == nil {
		pbq.Limit = proto.Int32(0)
	}
	return nil
}

// Run runs the query in the given context.
func (q *BaseQuery) Run(c appengine.Context) *Iterator {
	// Make a copy of the query.
	req := *q.pbq
	if err := q.toProto(&req, false); err != nil {
		return &Iterator{err: q.err}
	}
	var limit, offset int32
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}
	req.App = proto.String(c.FullyQualifiedAppID())
	t := &Iterator{
		c:      c,
		offset: offset,
		limit:  limit,
	}
	if err := c.Call("datastore_v3", "RunQuery", &req, &t.res, nil); err != nil {
		t.err = err
		return t
	}
	return t
}

// validateInt32 validates that an int is positive ad doesn't overflow.
func validateInt32(v int, name string) error {
	if v < 0 {
		return fmt.Errorf("datastore: negative value for %v", name)
	}
	if v > math.MaxInt32 {
		return fmt.Errorf("datastore: value overflow for %v", name)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Iterator
// ----------------------------------------------------------------------------

// Done is returned when a query iteration has completed.
var Done = errors.New("datastore: query has no more results")

// Iterator is the result of running a query.
type Iterator struct {
	c      appengine.Context
	offset int32
	limit  int32
	res    pb.QueryResult
	err    error
}

// Next returns the key of the next result. When there are no more results,
// Done is returned as the error.
// If the query is not keys only, it also loads the entity stored for that key
// into the struct pointer or PropertyLoadSaver dst, with the same semantics
// and possible errors as for the Get function.
// If the query is keys only, it is valid to pass a nil interface{} for dst.
func (t *Iterator) Next(dst interface{}) (*Key, error) {
	k, e, err := t.next()
	if err != nil || e == nil {
		return k, err
	}
	return k, loadEntity(dst, e)
}

func (t *Iterator) next() (*Key, *pb.EntityProto, error) {
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
		if err := callNext(t.c, &t.res, t.offset, t.limit, false); err != nil {
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
			t.err = errors.New("datastore: internal error: limit exhausted but more_results is true")
			return nil, nil, t.err
		}
	}

	// Pop the EntityProto from the front of t.res.Result and
	// extract its key.
	var e *pb.EntityProto
	e, t.res.Result = t.res.Result[0], t.res.Result[1:]
	if e.Key == nil {
		return nil, nil, errors.New("datastore: internal error: server did not return a key")
	}
	k, err := protoToKey(e.Key)
	if err != nil || k.Incomplete() {
		return nil, nil, errors.New("datastore: internal error: server returned an invalid key")
	}
	if proto.GetBool(t.res.KeysOnly) {
		return k, nil, nil
	}
	return k, e, nil
}

// callNext issues a datastore_v3/Next RPC to advance a cursor, such as that
// returned by a query with more results.
func callNext(c appengine.Context, res *pb.QueryResult, offset, limit int32, zeroLimitMeansZero bool) error {
	if res.Cursor == nil {
		return errors.New("datastore: internal error: server did not return a cursor")
	}
	// TODO: should I eventually call datastore_v3/DeleteCursor on the cursor?
	req := &pb.NextRequest{
		Cursor: res.Cursor,
		Offset: proto.Int32(offset),
	}
	if limit != 0 || zeroLimitMeansZero {
		req.Count = proto.Int32(limit)
	}
	if res.CompiledCursor != nil {
		req.Compile = proto.Bool(true)
	}
	res.Reset()
	return c.Call("datastore_v3", "Next", req, res, nil)
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
	/*
		options := &QueryOptions{
			limit:       0,
			// TODO: int32 conversion
			offset:      int32(offset),
			keysOnly:    true,
			compile:     true,
			startCursor: c,
		}
		return query.Run(ctx, options).Cursor()
	*/
	return nil
}

// DecodeCursor decodes a cursor from the opaque representation returned by
// Cursor.Encode.
func DecodeCursor(encoded string) (*Cursor, error) {
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
