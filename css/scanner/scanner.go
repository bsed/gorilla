// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scanner

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// tokenType identifies the type of lexical tokens.
type tokenType int

// String returns a string representation of the token type.
func (t tokenType) String() string {
	return tokenNames[t]
}

// Token represents a token and the corresponding string.
type Token struct {
	Type   tokenType
	Value  string
	Line   int
	Column int
}

// String returns a string representation of the token.
func (t *Token) String() string {
	if len(t.Value) > 10 {
		return fmt.Sprintf("%s (line: %d, column: %d): %.10q...",
			t.Type, t.Line, t.Column, t.Value)
	}
	return fmt.Sprintf("%s (line: %d, column: %d): %q",
		t.Type, t.Line, t.Column, t.Value)
}

// All tokens -----------------------------------------------------------------

// The complete list of tokens in CSS3.
const (
	// Scanner flags.
	TokenError tokenType = iota
	TokenEOF
	// From now on, only tokens from the CSS specification.
	TokenIdent
	TokenAtKeyword
	TokenString
	TokenHash
	TokenNumber
	TokenPercentage
	TokenDimension
	TokenURI
	TokenUnicodeRange
	TokenCDO
	TokenCDC
	TokenS
	TokenComment
	TokenFunction
	TokenIncludes
	TokenDashMatch
	TokenPrefixMatch
	TokenSuffixMatch
	TokenSubstringMatch
	TokenChar
	TokenBOM
)

// tokenNames maps tokenType's to their names. Used for conversion to string.
var tokenNames = map[tokenType]string{
	TokenError:          "error",
	TokenEOF:            "EOF",
	TokenIdent:          "IDENT",
	TokenAtKeyword:      "ATKEYWORD",
	TokenString:         "STRING",
	TokenHash:           "HASH",
	TokenNumber:         "NUMBER",
	TokenPercentage:     "PERCENTAGE",
	TokenDimension:      "DIMENSION",
	TokenURI:            "URI",
	TokenUnicodeRange:   "UNICODE-RANGE",
	TokenCDO:            "CDO",
	TokenCDC:            "CDC",
	TokenS:              "S",
	TokenComment:        "COMMENT",
	TokenFunction:       "FUNCTION",
	TokenIncludes:       "INCLUDES",
	TokenDashMatch:      "DASHMATCH",
	TokenPrefixMatch:    "PREFIXMATCH",
	TokenSuffixMatch:    "SUFFIXMATCH",
	TokenSubstringMatch: "SUBSTRINGMATCH",
	TokenChar:           "CHAR",
	TokenBOM:            "BOM",
}

// Macros and productions -----------------------------------------------------
// http://www.w3.org/TR/css3-syntax/#tokenization

var macroRegexp = regexp.MustCompile(`\{[a-z]+\}`)

// macros maps macro names to patterns to be expanded.
var macros = map[string]string{
	// must be escaped: `\.+*?()|[]{}^$`
	"ident":      `-?{nmstart}{nmchar}*`,
	"name":       `{nmchar}+`,
	"nmstart":    `[a-zA-Z_]|{nonascii}|{escape}`,
	"nonascii":   "[\u0080-\uD7FF\uE000-\uFFFD\U00010000-\U0010FFFF]",
	"unicode":    `\\[0-9a-fA-F]{1,6}{wc}?`,
	"escape":     "{unicode}|\\[\u0020-\u007E\u0080-\uD7FF\uE000-\uFFFD\U00010000-\U0010FFFF]",
	"nmchar":     `[a-zA-Z0-9_-]|{nonascii}|{escape}`,
	"num":        `[0-9]+|[0-9]*\.[0-9]+`,
	"string":     `"(?:{stringchar}|')*"|'(?:{stringchar}|")*'`,
	"stringchar": `{urlchar}|[ ]|\\{nl}`,
	"urlchar":    "[\u0009\u0021\u0023-\u0026\u0027-\u007E]|{nonascii}|{escape}",
	"nl":         `[\n\r\f]|\r\n`,
	"w":          `{wc}*`,
	"wc":         `[\t\n\f\r ]`,
}

// productions maps the list of tokens to patterns to be expanded.
var productions = map[tokenType]string{
	TokenIdent:          `{ident}`,
	TokenAtKeyword:      `@{ident}`,
	TokenString:         `{string}`,
	TokenHash:           `#{name}`,
	TokenNumber:         `{num}`,
	TokenPercentage:     `{num}%`,
	TokenDimension:      `{num}{ident}`,
	TokenURI:            `url\({w}(?:{string}|{urlchar}*){w}\)`,
	TokenUnicodeRange:   `U\+[0-9A-F\?]{1,6}(?:-[0-9A-F]{1,6})?`,
	TokenCDO:            `<!--`,
	TokenCDC:            `-->`,
	TokenS:              `{wc}+`,
	TokenComment:        `/\*[^\*]*[\*]+(?:[^/][^\*]*[\*]+)*/`,
	TokenFunction:       `{ident}\(`,
	TokenIncludes:       `~=`,
	TokenDashMatch:      `\|=`,
	TokenPrefixMatch:    `\^=`,
	TokenSuffixMatch:    `\$=`,
	TokenSubstringMatch: `\*=`,
	TokenChar:           `[^"']`,
	TokenBOM:            "\uFEFF",
}

// matchers maps the list of tokens to compiled regular expressions.
//
// The map is filled on init() using the macros and productions defined in
// the CSS specification.
var matchers = map[tokenType]*regexp.Regexp{}

// matchOrder is the order to test regexps when first-char shortcuts
// can't be used.
var matchOrder = []tokenType{
	TokenURI,
	TokenFunction,
	TokenUnicodeRange,
	TokenIdent,
	TokenDimension,
	TokenPercentage,
	TokenNumber,
	TokenCDC,
	TokenChar,
}

func init() {
	// replace macros and compile regexps for productions.
	replaceMacro := func(s string) string {
		return "(?:" + macros[s[1:len(s)-1]] + ")"
	}
	for t, s := range productions {
		for macroRegexp.MatchString(s) {
			s = macroRegexp.ReplaceAllStringFunc(s, replaceMacro)
		}
		matchers[t] = regexp.MustCompile("^(?:" + s + ")")
	}
}

// Scanner --------------------------------------------------------------------

// New returns a new CSS scanner for the given input.
func New(input string) *Scanner {
	// Normalize newlines.
	input = strings.Replace(input, "\r\n", "\n", -1)
	return &Scanner{
		input:  input,
		line:   1,
		column: 1,
	}
}

// Scanner scans an input and emits tokens following the CSS3 specification.
type Scanner struct {
	input  string
	pos    int
	line   int
	column int
	last   *Token
}

// Next returns the next token from the input.
//
// At the end of the input the token type is TokenEOF.
//
// If the input can't be tokenized the token type is TokenError. This occurs
// in case of unclosed quotation marks or comments.
func (s *Scanner) Next() *Token {
	if s.last != nil {
		return s.last
	}
	if s.pos >= len(s.input) {
		s.last = &Token{TokenEOF, "", -1, -1}
		return s.last
	}
	input := s.input[s.pos:]
	if s.pos == 0 {
		// Test BOM only at the beginning of the file.
		if strings.HasPrefix(input, "\uFEFF") {
			return s.emitToken(TokenBOM, "\uFEFF")
		}
	}
	// There's a lot we can guess based on the current rune so we'll take this
	// shortcut before testing multiple regexps.
	r := input[0]
	switch r {
	case '\t', '\n', '\f', '\r', ' ':
		// Whitespace.
		return s.emitToken(TokenS, matchers[TokenS].FindString(input))
	case '.':
		// Dot is too common to not have a quick check.
		// We'll test if this is a Char; if it is followed by a number it is a
		// dimension/percentage/number, and this will be matched later.
		if len(input) > 1 && !unicode.IsDigit(rune(input[1])) {
			return s.emitToken(TokenChar, ".")
		}
	case '#':
		// Hash is also a common one. If the regexp doesn't match it is a Char.
		if match := matchers[TokenHash].FindString(input); match != "" {
			return s.emitToken(TokenHash, match)
		}
		return s.emitToken(TokenChar, "#")
	case '@':
		// Another common one. If the regexp doesn't match it is a Char.
		if match := matchers[TokenAtKeyword].FindString(input); match != "" {
			return s.emitToken(TokenAtKeyword, match)
		}
		return s.emitToken(TokenChar, "@")
	case ':', ',', ';', '%', '&', '+', '=', '>', '(', ')', '[', ']', '{', '}':
		// More common chars.
		return s.emitToken(TokenChar, string(r))
	case '"', '\'':
		// String or error.
		match := matchers[TokenString].FindString(input)
		if match != "" {
			return s.emitToken(TokenString, match)
		} else {
			s.last = s.emitToken(TokenError, "unclosed quotation mark")
			return s.last
		}
	case '/':
		if len(input) > 1 && input[1] == '*' {
			// Comment or error.
			match := matchers[TokenComment].FindString(input)
			if match != "" {
				return s.emitToken(TokenComment, match)
			} else {
				s.last = s.emitToken(TokenError, "unclosed comment")
				return s.last
			}
		}
		// A simple char.
		return s.emitToken(TokenChar, "/")
	case '~':
		// Includes or Char.
		return s.emitPrefixOrChar(TokenIncludes, "~=")
	case '|':
		// DashMatch or Char.
		return s.emitPrefixOrChar(TokenDashMatch, "|=")
	case '^':
		// PrefixMatch or Char.
		return s.emitPrefixOrChar(TokenPrefixMatch, "^=")
	case '$':
		// SuffixMatch or Char.
		return s.emitPrefixOrChar(TokenSuffixMatch, "$=")
	case '*':
		// SubstringMatch or Char.
		return s.emitPrefixOrChar(TokenSubstringMatch, "*=")
	case '<':
		// CDO or Char.
		return s.emitPrefixOrChar(TokenCDO, "<!--")
	}
	// Test all regexps, in order.
	for _, token := range matchOrder {
		if match := matchers[token].FindString(input); match != "" {
			return s.emitToken(token, match)
		}
	}
	s.last = s.emitToken(TokenError, "impossible to tokenize")
	return s.last
}

// updatePosition updates input coordinates based on the consumed text.
func (s *Scanner) updatePosition(text string) {
	count := utf8.RuneCountInString(text)
	lines := strings.Count(text, "\n")
	s.line += lines
	if lines == 0 {
		s.column += count
	} else {
		s.column = utf8.RuneCountInString(text[strings.LastIndex(text, "\n"):])
	}
	s.pos += count
}

// emitToken returns a Token for the string v and updates the scanner position.
func (s *Scanner) emitToken(t tokenType, v string) *Token {
	token := &Token{t, v, s.line, s.column}
	s.updatePosition(v)
	return token
}

// emitPrefixOrChar returns a Token for type t if the current position
// matches the given prefix. Otherwise it returns a Char token using the
// first character from the prefix.
func (s *Scanner) emitPrefixOrChar(t tokenType, prefix string) *Token {
	if strings.HasPrefix(s.input[s.pos:], prefix) {
		return s.emitToken(t, prefix)
	}
	return s.emitToken(TokenChar, string(prefix[0]))
}
