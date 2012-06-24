// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pluralforms

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	eof     = -1
	numbers = "0123456789"
	symbols = "*/%+-=!<>|&?:n"
)

// tokenType is the type of lex tokens.
type tokenType int

const (
	tokenError      tokenType = iota
	tokenEOF
	tokenBool
	tokenInt
	tokenVar
	tokenMul        // *
	tokenDiv        // /
	tokenMod        // %
	tokenAdd        // +
	tokenSub        // - (binary)
	tokenEq         // ==
	tokenNotEq      // !=
	tokenGt         // >
	tokenGte        // >=
	tokenLt         // <
	tokenLte        // <=
	tokenOr         // ||
	tokenAnd        // &&
	tokenIf         // ?
	tokenElse       // :
	tokenNot        // !
	tokenLeftParen  // (
	tokenRightParen // )
)

var stringToToken = map[string]tokenType{
	"*":  tokenMul,
	"/":  tokenDiv,
	"%":  tokenMod,
	"+":  tokenAdd,
	"-":  tokenSub,
	"==": tokenEq,
	"!=": tokenNotEq,
	">":  tokenGt,
	">=": tokenGte,
	"<":  tokenLt,
	"<=": tokenLte,
	"||": tokenOr,
	"&&": tokenAnd,
	"?":  tokenIf,
	":":  tokenElse,
	"!":  tokenNot,
	"(":  tokenLeftParen,
	")":  tokenRightParen,
	"n":  tokenVar,
}

// ----------------------------------------------------------------------------

// token is a token returned from the lexer.
type token struct {
	typ tokenType
	val string
}

// ----------------------------------------------------------------------------

func newLexer(input string) *lexer {
	return &lexer{input: input}
}

type lexer struct {
	input string // string being scanned
	pos   int    // current position in the input
	width int    // width of last rune read from input
}

// next returns the next token from the input.
func (l *lexer) next() token {
	for {
		r := l.nextRune()
		switch r {
		case eof:
			return token{typ:tokenEOF}
		case ' ':
			// just ignore spaces.
		case '(':
			return token{typ: tokenLeftParen}
		case ')':
			return token{typ: tokenRightParen}
		default:
			l.backup()
			if s := l.nextRun(numbers); s != "" {
				return token{typ: tokenInt, val: s}
			}
			if s := l.nextRun(symbols); s != "" {
				if typ, ok := stringToToken[s]; ok {
					return token{typ: typ}
				}
			}
			return token{tokenError,
				fmt.Sprintf("Invalid character %q",	string(r))}
		}
	}
	panic("unreachable")
}

// next returns the next rune from the input.
func (l *lexer) nextRune() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// nextRun returns a run of runes from the valid set.
func (l *lexer) nextRun(valid string) string {
	pos := l.pos
	for strings.IndexRune(valid, l.nextRune()) >= 0 {}
	l.backup()
	return l.input[pos:l.pos]
}

// backup steps back one rune. Can only be called once per call of nextRune.
func (l *lexer) backup() {
	l.pos -= l.width
}
