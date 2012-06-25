package pluralforms

import (
	"fmt"
	"strconv"
)

// Operator precedence levels.
var precedence = map[tokenType]int{
	tokenNot:   6,
	tokenMul:   5,
	tokenDiv:   5,
	tokenMod:   5,
	tokenAdd:   4,
	tokenSub:   4,
	tokenEq:    3,
	tokenNotEq: 3,
	tokenGt:    3,
	tokenGte:   3,
	tokenLt:    3,
	tokenLte:   3,
	tokenOr:    2,
	tokenAnd:   1,
}

// Map of operators that are right-associative. We don't have any. :P
var rightAssociativity = map[tokenType]bool{}

// ----------------------------------------------------------------------------

// parse parses an expression and returns a parse tree.
func parse(expr string) (node, error) {
	p := &parser{stream: newTokenStream(expr)}
	return p.parse()
}

// parser parses basic arithmetic expressions and returns a parse tree.
//
// It uses the recursive descent "precedence climbing" algorithm from:
//
//     http://www.engr.mun.ca/~theo/Misc/exp_parsing.htm
type parser struct {
	stream *tokenStream
}

// expect consumes the next token if it matches the given type, or returns
// an error.
func (p *parser) expect(t tokenType) error {
	next := p.stream.pop()
	if next.typ == t {
		return nil
	}
	p.stream.push(next)
	return fmt.Errorf("Expected token %q, got %q", t, next.typ)
}

// parse consumes the token stream and returns a parse tree.
func (p *parser) parse() (node, error) {
	n, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	if err := p.expect(tokenEOF); err != nil {
		return nil, err
	}
	return n, nil
}

// parseExpression parses and returns an expression node.
func (p *parser) parseExpression(prec int) (node, error) {
	n, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	var t token
	for {
		t = p.stream.pop()
		q := precedence[t.typ]
		if !isBinaryOp(t) || q < prec {
			break
		}
		if !rightAssociativity[t.typ] {
			q += 1
		}
		n1, err := p.parseExpression(q)
		if err != nil {
			return nil, err
		}
		n = newBinaryOpNode(t, n, n1)
	}
	p.stream.push(t)
	return n, nil
}

// parsePrimary parses and returns a primary node.
func (p *parser) parsePrimary() (node, error) {
	t := p.stream.pop()
	if isUnaryOp(t) {
		n, err := p.parseExpression(precedence[t.typ])
		if err != nil {
			return nil, err
		}
		return newUnaryOpNode(t, n), nil
	} else if t.typ == tokenLeftParen {
		n, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		if err := p.expect(tokenRightParen); err != nil {
			return nil, err
		}
		return n, nil
	} else if isValue(t) {
		n, err := newValueNode(t)
		if err != nil {
			return nil, err
		}
		return n, nil
	}
	return nil, fmt.Errorf("Unexpected token %q", t)
}

// ----------------------------------------------------------------------------

// isBinaryOp returns true if the given token is a binary operator.
func isBinaryOp(t token) bool {
	switch t.typ {
	case tokenMul, tokenDiv, tokenMod,
		tokenAdd, tokenSub,
		tokenEq, tokenNotEq, tokenGt, tokenGte, tokenLt, tokenLte,
		tokenOr, tokenAnd:
		return true
	}
	return false
}

// newBinaryOpNode returns a tree for the given binary operator
// and child nodes.
func newBinaryOpNode(t token, n1, n2 node) node {
	return &binaryOpNode{op: t.typ, n1: n1, n2: n2}
}

// isUnaryOp returns true if the given token is a unary operator.
func isUnaryOp(t token) bool {
	switch t.typ {
	case tokenNot:
		return true
	}
	return false
}

// newUnaryOpNode returns a tree for the given unary operator and child node.
func newUnaryOpNode(t token, n1 node) node {
	return &unaryOpNode{op: t.typ, n1: n1}
}

// isValue returns true if the given token is a literal or variable.
func isValue(t token) bool {
	switch t.typ {
	case tokenBool, tokenInt, tokenVar:
		return true
	}
	return false
}

// newValueNode returns a node for the given literal or variable.
func newValueNode(t token) (node, error) {
	// bool is never extracted directly from the expression
	switch t.typ {
	case tokenInt:
		return newIntNode(t.val)
	case tokenVar:
		return &varNode{}, nil
	}
	panic("unreachable")
}

// ----------------------------------------------------------------------------

type node interface {
	Eval(ctx int) node
	BinaryOp(ctx int, op tokenType, n2 node) node
	UnaryOp(ctx int, op tokenType) node
	String() string
}

// ----------------------------------------------------------------------------

var invalidExpression = errorNode("Invalid expression")

type errorNode string

func (n errorNode) Eval(ctx int) node {
	return n
}

func (n errorNode) BinaryOp(ctx int, op tokenType, n2 node) node {
	return n
}

func (n errorNode) UnaryOp(ctx int, op tokenType) node {
	return n
}

func (n errorNode) String() string {
	return string(n)
}

// ----------------------------------------------------------------------------

type binaryOpNode struct {
	op tokenType
	n1 node
	n2 node
}

func (n *binaryOpNode) Eval(ctx int) node {
	return n.n1.BinaryOp(ctx, n.op, n.n2)
}

func (n *binaryOpNode) BinaryOp(ctx int, op tokenType, n2 node) node {
	return n.Eval(ctx).BinaryOp(ctx, op, n2)
}

func (n *binaryOpNode) UnaryOp(ctx int, op tokenType) node {
	return n.Eval(ctx).UnaryOp(ctx, op)
}

func (n *binaryOpNode) String() string {
	return fmt.Sprintf("<%s%s%s>", n.n1, n.op, n.n2)
}

// ----------------------------------------------------------------------------

type unaryOpNode struct {
	op tokenType
	n1 node
}

func (n *unaryOpNode) Eval(ctx int) node {
	return n.n1.UnaryOp(ctx, n.op)
}

func (n *unaryOpNode) BinaryOp(ctx int, op tokenType, n2 node) node {
	return n.Eval(ctx).BinaryOp(ctx, op, n2)
}

func (n *unaryOpNode) UnaryOp(ctx int, op tokenType) node {
	return n.Eval(ctx).UnaryOp(ctx, op)
}

func (n *unaryOpNode) String() string {
	return fmt.Sprintf("<%s%s>", n.op, n.n1)
}

// ----------------------------------------------------------------------------

type boolNode bool

func (n boolNode) Eval(ctx int) node {
	return n
}

func (x boolNode) BinaryOp(ctx int, op tokenType, n2 node) node {
	switch y := n2.(type) {
	case errorNode:
		return y
	case *binaryOpNode:
		return y.BinaryOp(ctx, op, x)
	case boolNode:
		switch op {
		case tokenAnd:
			return boolNode(x && y)
		case tokenOr:
			return boolNode(x || y)
		case tokenEq:
			return boolNode(x == y)
		case tokenNotEq:
			return boolNode(x != y)
		}
	}
	return invalidExpression
}

func (n boolNode) UnaryOp(ctx int, op tokenType) node {
	// Shortcut. We only have one unary op, which is "logical not", and only
	// works for bool.
	return !n
}

func (n boolNode) String() string {
	return fmt.Sprintf("%v", bool(n))
}

// ----------------------------------------------------------------------------

func newIntNode(src string) (intNode, error) {
	if value, err := strconv.ParseInt(src, 10, 0); err == nil {
		return intNode(value), nil
	}
	return 0, fmt.Errorf("Invalid int %q", src)
}

type intNode int

func (n intNode) Eval(ctx int) node {
	return n
}

func (x intNode) BinaryOp(ctx int, op tokenType, n2 node) node {
	switch y := n2.(type) {
	case errorNode:
		return y
	case *varNode:
		return x.BinaryOp(ctx, op, intNode(ctx))
	case *binaryOpNode:
		return y.BinaryOp(ctx, op, x)
	case intNode:
		switch op {
		case tokenMul:
			return x * y
		case tokenDiv:
			return x / y
		case tokenMod:
			return x % y
		case tokenAdd:
			return x + y
		case tokenSub:
			return x - y
		case tokenEq:
			return boolNode(x == y)
		case tokenNotEq:
			return boolNode(x != y)
		case tokenGt:
			return boolNode(x > y)
		case tokenGte:
			return boolNode(x >= y)
		case tokenLt:
			return boolNode(x < y)
		case tokenLte:
			return boolNode(x <= y)
		}
	}
	return invalidExpression
}

func (n intNode) UnaryOp(ctx int, op tokenType) node {
	return invalidExpression
}

func (n intNode) String() string {
	return fmt.Sprintf("%v", int(n))
}

// ----------------------------------------------------------------------------

type varNode struct{}

func (n *varNode) Eval(ctx int) node {
	return intNode(ctx)
}

func (n *varNode) BinaryOp(ctx int, op tokenType, n2 node) node {
	return n.Eval(ctx).BinaryOp(ctx, op, n2)
}

func (n *varNode) UnaryOp(ctx int, op tokenType) node {
	return invalidExpression
}

func (n *varNode) String() string {
	return "n"
}
