// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package datastore

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"

	"appengine"
	"goprotobuf.googlecode.com/hg/proto"

	pb "appengine_internal/datastore"
)

// TODO
// ====
// - Async calls:
//   - http://code.google.com/appengine/docs/python/datastore/async.html
// - IN, OR and != filters
// - Maybe split Query.toProto() in smaller functions to perform checkings
//   related to datastore restrictions:
//   - http://code.google.com/appengine/docs/python/datastore/queries.html#Restrictions_on_Queries

// ----------------------------------------------------------------------------
// Query
// ----------------------------------------------------------------------------

// NewQuery creates a new Query for a specific entity kind.
func NewQuery(kind string) *Query {
	return &Query{kind: kind}
}

// Query represents a datastore query, and is immutable.
// Methods don't modify the query; instead they return a modified copy.
type Query struct {
	kind     string
	ancestor *Key
	filter   []*queryFilter
	order    []*queryOrder

	valid bool
	err   os.Error
}

// Kind sets the entity kind for the Query.
func (q *Query) Kind(kind string) *Query {
	c := *q
	c.kind = kind
	return &c
}

// Ancestor sets the ancestor filter for the Query.
func (q *Query) Ancestor(ancestor *Key) *Query {
	c := *q
	c.ancestor = ancestor
	return &c
}

// Filter adds a field-based filter to the Query.
// The filterStr argument must be a field name followed by optional space,
// followed by an operator, one of ">", "<", ">=", "<=", or "=".
// Fields are compared against the provided value using the operator.
// Multiple filters are AND'ed together.
func (q *Query) Filter(filterStr string, value interface{}) *Query {
	c := *q
	c.filter = append(c.filter, newQueryFilter(filterStr, value))
	return &c
}

// Order adds a field-based sort to the query.
// Orders are applied in the order they are added.
// The default order is ascending; to sort in descending
// order prefix the fieldName with a minus sign (-).
func (q *Query) Order(order string) *Query {
	c := *q
	c.order = append(c.order, newQueryOrder(order))
	return &c
}

// Error validates the query and returns an ErrMulti if errors were found.
func (q *Query) Error() os.Error {
	if err := q.validate(); err != nil {
		return err
	}
	return nil
}

// Run ------------------------------------------------------------------------

// Run runs the query in the given context.
func (q *Query) Run(c appengine.Context, options *QueryOptions) *Iterator {
	if options == nil {
		options = &QueryOptions{}
	}
	return newIterator(c, q, options, "RunQuery")
}

// Private methods ------------------------------------------------------------

// toProto converts the query to a protocol buffer.
//
// Values are stored as defined by the user and validation only happens here.
// It returns an ErrMulti with all encountered errors, if any.
func (q *Query) toProto(dst *pb.Query) os.Error {
	// Backend will complain about these; should we check before the RPC call?
	// os.NewError("datastore: kindless queries only support __key__ filters")
	// os.NewError("datastore: kindless queries only support ascending __key__ order")
	if err := q.validate(); err != nil {
		return err
	}
	if q.kind != "" {
		dst.Kind = proto.String(q.kind)
	}
	if q.ancestor != nil {
		dst.Ancestor = keyToProto(q.ancestor)
	}
	if q.filter != nil {
		dst.Filter = make([]*pb.Query_Filter, len(q.filter))
		for i, f := range q.filter {
			dst.Filter[i] = f.proto
		}
	}
	if q.order != nil {
		dst.Order = make([]*pb.Query_Order, len(q.order))
		for i, o := range q.order {
			dst.Order[i] = o.proto
		}
	}
	return nil
}

// validate checks the query and returns an ErrMulti if there are errors.
func (q *Query) validate() os.Error {
	// Because the query is immutable we only need to validate once,
	// so we store a flag marking it as valid.
	if q.valid {
		return nil
	} else if q.err != nil {
		return q.err
	}
	var err ErrMulti
	if q.ancestor != nil && q.ancestor.Incomplete() {
		err = append(err,
			os.NewError("datastore: incomplete query ancestor key"))
	}
	if q.filter != nil {
		for _, f := range q.filter {
			if f.err != nil {
				err = append(err, f.err)
			}
		}
	}
	if q.order != nil {
		for _, o := range q.order {
			if o.err != nil {
				err = append(err, o.err)
			}
		}
	}
	if len(err) > 0 {
		q.err = err
		return err
	}
	q.valid = true
	return nil
}

// ----------------------------------------------------------------------------
// QueryOptions
// ----------------------------------------------------------------------------

// NewQueryOptions creates a new configuration to run a query.
func NewQueryOptions(limit int, offset int) *QueryOptions {
	return &QueryOptions{limit: limit, offset: offset}
}

// QueryOptions defines a configuration to run a query, and is immutable.
// Methods don't modify the options; instead they return a modified copy.
type QueryOptions struct {
	limit        int
	offset       int
	keysOnly     bool
	compile      bool
	startCursor  *Cursor
	endCursor    *Cursor
	namespace    string // temporarily here until supported by appengine.Context
	batchSize    int    // hint for the number of results returned per RPC
	prefetchSize int    // hint for the number of results in the first RPC

	valid bool
	err   os.Error
}

// Limit sets the maximum number of keys/entities to return.
// A zero value means unlimited. A negative value is invalid.
func (o *QueryOptions) Limit(limit int) *QueryOptions {
	c := *o
	c.limit = limit
	return &c
}

// Offset sets how many keys to skip over before returning results.
// A negative value is invalid.
func (o *QueryOptions) Offset(offset int) *QueryOptions {
	c := *o
	c.offset = offset
	return &c
}

// KeysOnly configures the query to return keys, instead of keys and entities.
func (o *QueryOptions) KeysOnly(keysOnly bool) *QueryOptions {
	c := *o
	c.keysOnly = keysOnly
	return &c
}

// Compile configures the query to produce cursors.
func (o *QueryOptions) Compile(compile bool) *QueryOptions {
	c := *o
	c.compile = compile
	return &c
}

// Cursor sets the cursor position to start the query.
func (o *QueryOptions) Cursor(cursor *Cursor) *QueryOptions {
	c := *o
	c.startCursor = cursor
	return &c
}

// Namespace sets the namespace for the query.
//
// This is a temporary function to fill the gap until appengine.Context
// supports namespaces.
func (o *QueryOptions) Namespace(namespace string) *QueryOptions {
	c := *o
	c.namespace = namespace
	return &c
}

// Error validates the options and returns an ErrMulti if errors were found.
func (o *QueryOptions) Error() os.Error {
	if err := o.validate(); err != nil {
		return err
	}
	return nil
}

// Private methods ------------------------------------------------------------

// toProto converts the query to a protocol buffer.
//
// Values are stored as defined by the user and validation only happens here.
// It returns an ErrMulti with all encountered errors, if any.
//
// TODO: zero limit policy
func (o *QueryOptions) toProto(dst *pb.Query) os.Error {
	if err := o.validate(); err != nil {
		return err
	}
	dst.Limit = proto.Int32(int32(o.limit))
	dst.Offset = proto.Int32(int32(o.offset))
	dst.KeysOnly = proto.Bool(o.keysOnly)
	dst.Compile = proto.Bool(o.compile)
	if o.startCursor != nil {
		dst.CompiledCursor = o.startCursor.compiledCursor
	}
	if o.endCursor != nil {
		dst.EndCompiledCursor = o.endCursor.compiledCursor
	}
	if o.namespace != "" {
		dst.NameSpace = proto.String(o.namespace)
	}
	return nil
}

// validate checks all options and returns an ErrMulti if there are errors.
func (o *QueryOptions) validate() os.Error {
	// Because the options are immutable we only need to validate once,
	// so we store a flag marking them as valid.
	if o.valid {
		return nil
	} else if o.err != nil {
		return o.err
	}
	var err ErrMulti
	if e := validInt32(o.limit, "limit"); e != nil {
		err = append(err, e)
	}
	if e := validInt32(o.offset, "offset"); e != nil {
		err = append(err, e)
	}
	if e := validInt32(o.batchSize, "batchSize"); e != nil {
		err = append(err, e)
	}
	if e := validInt32(o.prefetchSize, "prefetchSize"); e != nil {
		err = append(err, e)
	}
	if o.startCursor != nil && o.startCursor.compiledCursor == nil {
		err = append(err, os.NewError("datastore: empty start cursor."))
	}
	if o.endCursor != nil && o.endCursor.compiledCursor == nil {
		err = append(err, os.NewError("datastore: empty end cursor."))
	}
	if len(err) > 0 {
		o.err = err
		return err
	}
	o.valid = true
	return nil
}

// ----------------------------------------------------------------------------
// queryFilter
// ----------------------------------------------------------------------------

var queryFilterOperatorToProto = map[string]*pb.Query_Filter_Operator{
	"<":  pb.NewQuery_Filter_Operator(pb.Query_Filter_LESS_THAN),
	"<=": pb.NewQuery_Filter_Operator(pb.Query_Filter_LESS_THAN_OR_EQUAL),
	"=":  pb.NewQuery_Filter_Operator(pb.Query_Filter_EQUAL),
	">=": pb.NewQuery_Filter_Operator(pb.Query_Filter_GREATER_THAN_OR_EQUAL),
	">":  pb.NewQuery_Filter_Operator(pb.Query_Filter_GREATER_THAN),
}

func newQueryFilter(filterStr string, value interface{}) *queryFilter {
	q := &queryFilter{}
	filterStr = strings.TrimSpace(filterStr)
	if filterStr == "" {
		q.err = os.NewError("datastore: empty query filter")
		return q
	}
	propStr := strings.TrimRight(filterStr, " ><=")
	if propStr == "" {
		q.err = os.NewError("datastore: empty query filter property")
		return q
	}
	opStr := strings.TrimSpace(filterStr[len(propStr):])
	op := queryFilterOperatorToProto[opStr]
	if op == nil {
		q.err = fmt.Errorf("datastore: invalid operator %q in filter %q",
			opStr, filterStr)
		return q
	}
	prop, errStr := valueToProto(propStr, reflect.ValueOf(value), false)
	if errStr != "" {
		q.err = fmt.Errorf("datastore: bad query filter value: %q", errStr)
		return q
	}
	q.proto = &pb.Query_Filter{
		Op:       op,
		Property: []*pb.Property{prop},
	}
	return q
}

// queryFilter stores a query filter as protobuff and any possible errors.
type queryFilter struct {
	proto *pb.Query_Filter
	err   os.Error
}

// ----------------------------------------------------------------------------
// queryOrder
// ----------------------------------------------------------------------------

var queryOrderDirectionToProto = map[string]*pb.Query_Order_Direction{
	"+": pb.NewQuery_Order_Direction(pb.Query_Order_ASCENDING),
	"-": pb.NewQuery_Order_Direction(pb.Query_Order_DESCENDING),
}

func newQueryOrder(order string) *queryOrder {
	q := &queryOrder{}
	property := strings.TrimSpace(order)
	direction := "+"
	if property[0] == '-' {
		property = strings.TrimSpace(property[1:])
		direction = "-"
	}
	if property == "" {
		q.err = os.NewError("datastore: empty query order property")
		return q
	}
	q.proto = &pb.Query_Order{
		Property:  proto.String(property),
		Direction: queryOrderDirectionToProto[direction],
	}
	return q
}

// queryOrder stores a query filter as protobuff and any possible errors.
type queryOrder struct {
	proto *pb.Query_Order
	err   os.Error
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// validInt32 validates that an int is positive ad doesn't overflow.
func validInt32(value int, name string) os.Error {
	if value < 0 {
		return os.NewError("datastore: negative value for " + name)
	} else if value > math.MaxInt32 {
		return os.NewError("datastore: value overflow for " + name)
	}
	return nil
}
