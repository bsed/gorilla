// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package datastore

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"

	"appengine"
	"goprotobuf.googlecode.com/hg/proto"

	pb "appengine_internal/datastore"
)

// ----------------------------------------------------------------------------
// Query
// ----------------------------------------------------------------------------

// NewQuery creates a new Query for a specific entity kind.
func NewQuery(kind string) *Query {
	return &Query{kind: kind}
}

// Query represents a datastore query, and is immutable.
type Query struct {
	kind     string
	ancestor *Key
	filter   []queryFilter
	order    []queryOrder

	// For innternal use.
	proto    *pb.Query
	err      os.Error
}

// Kind sets the entity kind for the Query.
func (q *Query) Kind(kind string) *Query {
	c := q.clone()
	c.kind = kind
	return c
}

// Ancestor sets the ancestor filter for the Query.
func (q *Query) Ancestor(ancestor *Key) *Query {
	c := q.clone()
	c.ancestor = ancestor
	return c
}

// Filter adds a field-based filter to the Query.
// The filterStr argument must be a field name followed by optional space,
// followed by an operator, one of ">", "<", ">=", "<=", or "=".
// Fields are compared against the provided value using the operator.
// Multiple filters are AND'ed together.
func (q *Query) Filter(filterStr string, value interface{}) *Query {
	c := q.clone()
	c.filter = append(c.filter, queryFilter{filterStr, value})
	return c
}

// Order adds a field-based sort to the query.
// Orders are applied in the order they are added.
// The default order is ascending; to sort in descending
// order prefix the fieldName with a minus sign (-).
func (q *Query) Order(order string) *Query {
	c := q.clone()
	c.order = append(c.order, queryOrder(order))
	return c
}

// String returns a representation of the query in a readable format.
func (q *Query) String() string {
	buf := bytes.NewBufferString("")
	var hasWhere bool
	if q.kind != "" {
		fmt.Fprintf(buf, "SELECT * FROM %v", q.kind)
	}
	if q.ancestor != nil {
		fmt.Fprintf(buf, " WHERE ANCESTOR IS KEY('%v')", q.ancestor.Encode())
		hasWhere = true
	}
	if q.filter != nil {
		for i, filter := range q.filter {
			if !hasWhere {
				buf.WriteString(" WHERE")
				hasWhere = true
			} else if hasWhere || i > 0 {
				buf.WriteString(" AND")
			}
			// TODO value doesn't follow GQL strictly.
			fmt.Fprintf(buf, " %v%#v", filter.property, filter.value)
		}
	}
	if q.order != nil {
		buf.WriteString(" ORDER BY")
		for i, order := range q.order {
			if i > 0 {
				buf.WriteString(",")
			}
			property := string(order)
			direction := "ASC"
			if strings.HasPrefix(property, "-") {
				property = property[1:]
				direction = "DESC"
			}
			fmt.Fprintf(buf, " %v %v", property, direction)
		}
	}
	return buf.String()
}

// Fetching -------------------------------------------------------------------

//SDK methods:
//func (q *Query) Count(c appengine.Context) (int, os.Error)
//func (q *Query) GetAll(c appengine.Context, dst interface{}) ([]*Key, os.Error)
//func (q *Query) Run(c appengine.Context) *Iterator

// Run runs the query in the given context. TODO.
func (q *Query) Run(c appengine.Context, options *FetchOptions) *Iterator {
	var req pb.Query
	if err := q.toProto(&req); err != nil {
		return &Iterator{err: q.err}
	}
	if options == nil {
		options = &FetchOptions{}
	} else if options.err != nil {
		return &Iterator{err: options.err}
	}
	return nil
}

// Private methods ------------------------------------------------------------

// clone returns a copy of the query.
func (q *Query) clone() *Query {
	return &Query{
		kind:     q.kind,
		ancestor: q.ancestor,
		filter:   q.filter,
		order:    q.order,
	}
}

// toProto converts the query to a protocol buffer.
func (q *Query) toProto(dst *pb.Query) os.Error {
	if q.proto == nil {
		q.setProto()
	}
	if q.err != nil {
		return q.err
	}
	dst.Kind = q.proto.Kind
	dst.Ancestor = q.proto.Ancestor
	dst.Filter = q.proto.Filter
	dst.Order = q.proto.Order
	return nil
}

func (q *Query) setProto() {
	var query pb.Query
	var errMulti ErrMulti
	if q.kind != "" {
		query.Kind = proto.String(q.kind)
	} else {
		errMulti = append(errMulti, os.NewError("datastore: empty query kind"))
	}
	if q.ancestor != nil {
		query.Ancestor = q.ancestor.toProto()
	}
	if q.filter != nil {
		query.Filter = make([]*pb.Query_Filter, len(q.filter))
		for i, f := range q.filter {
			var filter pb.Query_Filter
			if e := f.toProto(&filter); e != nil {
				errMulti = append(errMulti, e)
			}
			query.Filter[i] = &filter
		}
	}
	if q.order != nil {
		query.Order = make([]*pb.Query_Order, len(q.order))
		for i, o := range q.order {
			var order pb.Query_Order
			if e := o.toProto(&order); e != nil {
				errMulti = append(errMulti, e)
			}
			query.Order[i] = &order
		}
	}
	if len(errMulti) > 0 {
		q.err = errMulti
	}
	q.proto = &query
}

// ----------------------------------------------------------------------------
// FetchOptions
// ----------------------------------------------------------------------------

// NewFetchOptions creates a new configuration to run a query.
func NewFetchOptions(limit int, offset int) *FetchOptions {
	o := &FetchOptions{}
	o.setLimit(limit)
	o.setOffset(offset)
	return o
}

// FetchOptions defines a configuration to run a query, and is immutable.
type FetchOptions struct {
	limit       int32
	offset      int32
	keysOnly    bool
	compile     bool
	startCursor *Cursor
	endCursor   *Cursor
	// TODO?
	// batchSize: int, hint for the number of results returned per RPC
	// prefetchSize: int, hint for the number of results in the first RPC

	err os.Error
}

// Limit sets the maximum number of keys/entities to return.
// A zero value means unlimited. A negative value is invalid.
func (o *FetchOptions) Limit(limit int) *FetchOptions {
	c := o.clone()
	c.setLimit(limit)
	return c
}

// Offset sets how many keys to skip over before returning results.
// A negative value is invalid.
func (o *FetchOptions) Offset(offset int) *FetchOptions {
	c := o.clone()
	c.setOffset(offset)
	return c
}

// KeysOnly configures the query to return keys, instead of keys and entities.
func (o *FetchOptions) KeysOnly(keysOnly bool) *FetchOptions {
	c := o.clone()
	c.keysOnly = keysOnly
	return c
}

// Compile configures the query to produce cursors.
func (o *FetchOptions) Compile(compile bool) *FetchOptions {
	c := o.clone()
	c.compile = compile
	return c
}

// Cursor sets the cursor position to start the query.
func (o *FetchOptions) Cursor(cursor *Cursor) *FetchOptions {
	// TODO: When a cursor is set, should we automatically configure it
	// to produce cursors?
	c := o.clone()
	c.startCursor = cursor
	return c
}

// Private methods ------------------------------------------------------------

// clone returns a copy of the fetch options.
func (o *FetchOptions) clone() *FetchOptions {
	return &FetchOptions{
		limit:       o.limit,
		offset:      o.offset,
		keysOnly:    o.keysOnly,
		compile:     o.compile,
		startCursor: o.startCursor,
		endCursor:   o.endCursor,
	}
}

// setLimit sets the limit field checking if it is a valid value.
func (o *FetchOptions) setLimit(limit int) {
	if limit32, err := validInt32(limit, "limit"); err != nil {
		errMulti := o.err.(ErrMulti)
		errMulti = append(errMulti, err)
		o.err = errMulti
	} else {
		o.limit = limit32
	}
}

// setOffset sets the offset field checking if it is a valid value.
func (o *FetchOptions) setOffset(offset int) {
	if offset32, err := validInt32(offset, "offset"); err != nil {
		errMulti := o.err.(ErrMulti)
		errMulti = append(errMulti, err)
		o.err = errMulti
	} else {
		o.offset = offset32
	}
}

// ----------------------------------------------------------------------------
// Iterator
// ----------------------------------------------------------------------------

// Iterator. TODO.
type Iterator struct {
	err os.Error
}

// ----------------------------------------------------------------------------
// Cursor
// ----------------------------------------------------------------------------

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

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

var operatorToProto = map[string]*pb.Query_Filter_Operator{
	"<":  pb.NewQuery_Filter_Operator(pb.Query_Filter_LESS_THAN),
	"<=": pb.NewQuery_Filter_Operator(pb.Query_Filter_LESS_THAN_OR_EQUAL),
	"=":  pb.NewQuery_Filter_Operator(pb.Query_Filter_EQUAL),
	">=": pb.NewQuery_Filter_Operator(pb.Query_Filter_GREATER_THAN_OR_EQUAL),
	">":  pb.NewQuery_Filter_Operator(pb.Query_Filter_GREATER_THAN),
}

var orderDirectionToProto = map[string]*pb.Query_Order_Direction{
	"+": pb.NewQuery_Order_Direction(pb.Query_Order_ASCENDING),
	"-": pb.NewQuery_Order_Direction(pb.Query_Order_DESCENDING),
}

// queryFilter stores a query filter as defined by the user.
type queryFilter struct {
	property string
	value    interface{}
}

// toProto converts the filter to a pb.Query_Filter.
func (q queryFilter) toProto(dst *pb.Query_Filter) os.Error {
	filterStr := strings.TrimSpace(q.property)
	if filterStr == "" {
		return os.NewError("datastore: invalid query filter: " + filterStr)
	}
	propStr := strings.TrimRight(filterStr, " ><=")
	if propStr == "" {
		return os.NewError("datastore: empty query filter property")
	}
	opStr := strings.TrimSpace(filterStr[len(propStr):])
	op := operatorToProto[opStr]
	if op == nil {
		return fmt.Errorf("datastore: invalid operator %q in filter %q",
			opStr, filterStr)
	}
	prop, err := valueToProto(propStr, reflect.ValueOf(q.value), false)
	if err != "" {
		return fmt.Errorf("datastore: bad query filter value type: %q", err)
	}
	dst.Op = op
	dst.Property = []*pb.Property{prop}
	return nil
}

// queryOrder stores a query order as defined by the user.
type queryOrder string

// toProto converts the order to a pb.Query_Order.
func (q queryOrder) toProto(dst *pb.Query_Order) os.Error {
	property := strings.TrimSpace(string(q))
	direction := "+"
	if strings.HasPrefix(property, "-") {
		direction = "-"
		property = strings.TrimSpace(property[1:])
	}
	if property == "" {
		return os.NewError("datastore: empty query order property")
	}
	dst.Property = proto.String(property)
	dst.Direction = orderDirectionToProto[direction]
	return nil
}

// validInt32 validates that an int is positive ad doesn't overflow.
func validInt32(value int, name string) (res int32, err os.Error) {
	if value < 0 {
		return res, os.NewError("datastore: negative value for " + name)
	} else if value > math.MaxInt32 {
		return res, os.NewError("datastore: value overflow for " + name)
	}
	return int32(value), nil
}
