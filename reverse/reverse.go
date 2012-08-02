// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reverse

import (
	"bytes"
	"fmt"
	"regexp"
	"regexp/syntax"
)

// Regexp stores a regular expression that can be "reverted" or "built":
// outermost capturing groups become placeholders to be filled by variables.
//
// For example, given a Regexp with the pattern `1(\d+)3`, we can call
// re.Revert([]string{"2"}, nil) to get a resulting string "123".
// This also works for named capturing groups: we can revert `1(?P<two>\d+)3`
// calling re.Revert(nil, map[string]string{"two": "2"}).
//
// There are a few limitations that can't be changed:
//
// 1. Nested capturing groups are ignored; only the outermost groups become
// a placeholder. So in `1(\d+([a-z]+))3` there is only one placeholder
// although there are two capturing groups: re.Revert([]string{"2", "a"}, nil)
// results in "123" and not "12a3".
//
// 2. Literals inside capturing groups are ignored; the whole group becomes
// a placeholder.
type Regexp struct {
	compiled *regexp.Regexp // compiled regular expression
	template string         // reverse template
	groups   []string       // order of positional and named capturing groups;
							// names for named and empty strings for positional
	indices  []int          // indices of the outermost groups
}

// Compile compiles the regular expression pattern and creates a template
// to revert it.
func Compile(pattern string) (*Regexp, error) {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	re, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		return nil, err
	}
	tpl := &template{buffer: new(bytes.Buffer)}
	tpl.write(re)
	return &Regexp{
		compiled: compiled,
		template: tpl.buffer.String(),
		groups:   tpl.groups,
		indices:  tpl.indices,
	}, nil
}

// Regexp returns the compiled regular expression to be used for matching.
func (r *Regexp) Regexp() *regexp.Regexp {
	return r.compiled
}

// Groups returns an ordered list of the outermost capturing groups found in
// the regexp, and the indices of these groups.
//
// Positional groups are listed as an empty string and named groups use
// the group name.
func (r *Regexp) Groups() ([]string, []int) {
	return r.groups, r.indices
}

// Revert builds a string for this regexp using the given values.
//
// The args parameter is used for positional and named capturing groups,
// and the kwds parameter is optionally used for named groups only;
// if a name is not provided in kwds, the value is taken from args, in order.
func (r *Regexp) Revert(args []string, kwds map[string]string) (string, error) {
	i := 0
	values := make([]interface{}, len(r.groups))
	for k, v := range r.groups {
		if v != "" && kwds != nil {
			// A named group. Check if it was passed in kwds.
			if tmp, ok := kwds[v]; ok {
				values[k] = tmp
				continue
			}
		}
		if i >= len(args) {
			return "", fmt.Errorf(
				"Not enough values to revert the regexp " +
				"(expected %d variables)", len(r.groups))
		}
		values[k] = args[i]
		i++
	}
	return fmt.Sprintf(r.template, values...), nil
}

// ValidRevert is the same as Revert but it also validates the resulting
// string matching it against the compiled regexp.
func (r *Regexp) ValidRevert(args []string, kwds map[string]string) (string, error) {
	reverse, err := r.Revert(args, kwds)
	if err != nil {
		return "", err
	}
	if !r.compiled.MatchString(reverse) {
		return "", fmt.Errorf("Resulting string doesn't match the regexp: %q",
			reverse)
	}
	return reverse, nil
}

// template builds a reverse template for a regexp.
type template struct {
	buffer  *bytes.Buffer
	groups  []string      // outermost capturing groups: empty string for
						  // positional or name for named groups
	indices []int         // indices of outermost capturing groups
	index   int           // current group index
	level   int           // current capturing group nesting level
}

// write writes a reverse template to the buffer.
func (t *template) write(re *syntax.Regexp) {
	switch re.Op {
	case syntax.OpLiteral:
		if t.level == 0 {
			for _, r := range re.Rune {
				t.buffer.WriteRune(r)
				if r == '%' {
					t.buffer.WriteRune('%')
				}
			}
		}
	case syntax.OpCapture:
		t.level++
		t.index++
		if t.level == 1 {
			t.groups = append(t.groups, re.Name)
			t.indices = append(t.indices, t.index)
			t.buffer.WriteString("%s")
		}
		for _, sub := range re.Sub {
			t.write(sub)
		}
		t.level--
	case syntax.OpConcat:
		for _, sub := range re.Sub {
			t.write(sub)
		}
	}
}
