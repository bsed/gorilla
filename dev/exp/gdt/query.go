package gdt

import (
	"errors"
	"fmt"
	"strings"

	"code.google.com/p/goprotobuf/proto"
	"appengine"
	pb "appengine_internal/datastore"
)

// ----------------------------------------------------------------------------
// BaseQuery
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

type queryOperator int

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

const (
	QueryDirectionAscending queryDirection = iota
	QueryDirectionDescending
)

var queryDirectionToProto = map[queryDirection]*pb.Query_Order_Direction{
	QueryDirectionAscending:  pb.NewQuery_Order_Direction(pb.Query_Order_ASCENDING),
	QueryDirectionDescending: pb.NewQuery_Order_Direction(pb.Query_Order_DESCENDING),
}

func NewBaseQuery() *BaseQuery {
	return &BaseQuery{pbq: new(pb.Query)}
}

type BaseQuery struct {
	pbq *pb.Query
	err error
}

func (q *BaseQuery) Clone() *BaseQuery {
	return &BaseQuery{pbq: &(*q.pbq), err: q.err}
}

func (q *BaseQuery) Namespace(namespace string) *BaseQuery {
	if q.err == nil {
		q.pbq.NameSpace = proto.String(namespace)
	}
	return q
}

func (q *BaseQuery) Kind(kind string) *BaseQuery {
	if q.err == nil {
		q.pbq.Kind = proto.String(kind)
	}
	return q
}

func (q *BaseQuery) Filter(property string, operator queryOperator,
	value interface{}) *BaseQuery {
	if q.err == nil {
		var p *pb.Property
		p, q.err = valueToProto(property, value)
		if q.err == nil {
			q.pbq.Filter = append(q.pbq.Filter, &pb.Query_Filter{
				Op:       queryOperatorToProto[operator],
				Property: []*pb.Property{p},
			})
		}
	}
	return q
}

func (q *BaseQuery) Order(property string, direction queryDirection) *BaseQuery {
	if q.err == nil {
		q.pbq.Order = append(q.pbq.Order, &pb.Query_Order{
			Property:  proto.String(property),
			Direction: queryDirectionToProto[direction],
		})
	}
	return q
}

func (q *BaseQuery) Ancestor(key *Key) *BaseQuery {
	if q.err == nil {
		if key == nil {
			q.pbq.Ancestor = nil
		} else {
			q.pbq.Ancestor = key.toProto()
		}
	}
	return q
}

// TODO
func (q *BaseQuery) Run(c appengine.Context, o *QueryOptions) *Iterator {
	// Make a copy of the query.
	pbq := *q.pbq
	if err := q.toProto(&pbq, o, zeroLimitMeansUnlimited); err != nil {
		return &Iterator{err: q.err}
	}
	return nil
}

func (q *BaseQuery) toProto(pbq *pb.Query, o *QueryOptions, zlp zeroLimitPolicy) error {
	if pbq.Kind == nil {
		return errors.New("gdt: empty query kind")
	}
	if q.err != nil {
		return q.err
	}
	if o.err != nil {
		return o.err
	}
	if o.keysOnly {
		pbq.KeysOnly = proto.Bool(true)
		pbq.RequirePerfectPlan = proto.Bool(true)
	}
	if o.limit != 0 || zlp == zeroLimitMeansZero {
		pbq.Limit = proto.Int32(o.limit)
	}
	if o.offset != 0 {
		pbq.Offset = proto.Int32(o.offset)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Query
// ----------------------------------------------------------------------------

func NewQuery(kind string) *Query {
	return &Query{base: NewBaseQuery().Kind(kind)}
}

type Query struct {
	base    *BaseQuery
	aliases map[string]string
}

func (q *Query) SetPropertyAliases(aliases map[string]string) *Query {
	q.aliases = aliases
	return q
}

func (q *Query) getPropertyName(alias string) string {
	if q.aliases != nil {
		if name, ok := q.aliases[alias]; ok {
			return name
		}
	}
	return alias
}

func (q *Query) Clone() *Query {
	return &Query{base: q.base.Clone(), aliases: q.aliases}
}

func (q *Query) Namespace(namespace string) *Query {
	q.base.Namespace(namespace)
	return q
}

func (q *Query) Kind(kind string) *Query {
	q.base.Kind(kind)
	return q
}

func (q *Query) Filter(filter string, value interface{}) *Query {
	property := strings.TrimRight(filter, " ><=")
	var operator queryOperator
	switch strings.TrimSpace(filter[len(property):]) {
		case "<":
			operator = QueryOperatorLessThan
		case "<=":
			operator = QueryOperatorLessThanOrEqual
		case "=":
			operator = QueryOperatorEqual
		case ">=":
			operator = QueryOperatorGreaterThanOrEqual
		case ">":
			operator = QueryOperatorGreaterThan
		default:
			q.base.err = fmt.Errorf("gdt: invalid query filter %q", filter)
			return q
	}
	q.base.Filter(q.getPropertyName(property), operator, value)
	return q
}

func (q *Query) Order(order string) *Query {
	property := order
	direction := QueryDirectionAscending
	if strings.HasPrefix(order, "-") {
		property = strings.TrimSpace(order[1:])
		direction = QueryDirectionDescending
	}
	q.base.Order(q.getPropertyName(property), direction)
	return q
}

func (q *Query) Ancestor(key *Key) *Query {
	q.base.Ancestor(key)
	return q
}

func (q *Query) Run(c appengine.Context, o *QueryOptions) *Iterator {
	return q.base.Run(c, o)
}

// TODO
func (q *Query) GetAll(c appengine.Context, dst interface{}) ([]*Key, error) {
	return nil, nil
}

// TODO
func (q *Query) Count(c appengine.Context) (int, error) {
	return 0, nil
}

// ----------------------------------------------------------------------------
// QueryOptions
// ----------------------------------------------------------------------------

// TODO
type QueryOptions struct {
	keysOnly bool
	limit    int32
	offset   int32
	err      error
}

// ----------------------------------------------------------------------------
// Iterator
// ----------------------------------------------------------------------------

// TODO
type Iterator struct {
	err error
}
