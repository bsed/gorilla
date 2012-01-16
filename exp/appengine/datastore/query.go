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

	//"appengine"
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
	var err ErrMulti
	if q.kind != "" {
		query.Kind = proto.String(q.kind)
	} else {
		err = append(err, os.NewError("datastore: empty query kind"))
	}
	if q.ancestor != nil {
		query.Ancestor = q.ancestor.toProto()
	}
	if q.filter != nil {
		query.Filter = make([]*pb.Query_Filter, len(q.filter))
		for i, f := range q.filter {
			var filter pb.Query_Filter
			if e := f.toProto(&filter); e != nil {
				err = append(err, e)
			}
			query.Filter[i] = &filter
		}
	}
	if q.order != nil {
		query.Order = make([]*pb.Query_Order, len(q.order))
		for i, o := range q.order {
			var order pb.Query_Order
			if e := o.toProto(&order); e != nil {
				err = append(err, e)
			}
			query.Order[i] = &order
		}
	}
	if len(err) > 0 {
		q.err = err
	}
	q.proto = &query
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

type queryFilter struct {
	property string
	value    interface{}
}

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

type queryOrder string

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

func validateLimit(limit int) os.Error {
	if limit < 0 {
		return os.NewError("datastore: negative query limit")
	}
	if limit > math.MaxInt32 {
		return os.NewError("datastore: query limit overflow")
	}
	return nil
}

func validateOffset(offset int) os.Error {
	if offset < 0 {
		return os.NewError("datastore: negative query offset")
	}
	if offset > math.MaxInt32 {
		return os.NewError("datastore: query offset overflow")
	}
	return nil
}
