package reverse

import (
	"bytes"
	"fmt"
	"regexp"
	"regexp/syntax"
)

// ReverseRegexp stores a regular expression that can be "reverted" or "built":
// outermost capturing groups become placeholders to be filled by variables.
//
// For example, given a ReverseRegexp with the pattern `1(\d+)3`, we can
// call re.Revert([]string{"2"}, nil) to get a resulting string "123".
// This also works for named capturing groups: we can revert `1(?P<two>\d+)3`
// calling re.Revert(nil, map[string]string{"two": "2"}).
//
// There are a few gotchas:
//
// 1. Nested capturing groups are ignored; only the outermost group becomes
// a placeholder. So in `1(\d+([a-z]+))3` there is only one placeholder
// although there are two capturing groups: re.Revert([]string{"2", "a"}, nil)
// results in "123" and not "12a3".
//
// 2. Literals inside capturing groups are ignored; the whole group becomes
// a placeholder.
type ReverseRegexp struct {
	compiled *regexp.Regexp // compiled regular expression
	template string         // reverse template
	groups   []string       // order of positional and named capturing groups;
							// names for named and empty strings for positional
}

// Compile compiles the regular expression pattern and creates a template
// to revert it.
func Compile(pattern string) (*ReverseRegexp, error) {
	re, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		return nil, err
	}
	var template bytes.Buffer
	var groups []string
	writeTemplate(&template, &groups, re)
	return &ReverseRegexp{
		compiled: regexp.MustCompile(pattern),
		template: template.String(),
		groups:   groups,
	}, nil
}

// Regexp returns the compiled regular expression to be used for matching.
func (r *ReverseRegexp) Regexp() *regexp.Regexp {
	return r.compiled
}

// Groups returns an ordered list of the outermost capturing groups found in
// the regexp.
//
// Positional groups are listed as an empty string and named groups use
// the group name.
func (r *ReverseRegexp) Groups() []string {
	return r.groups
}

// Revert builds a string for this regexp using the given values.
//
// The args parameter is used for positional and named capturing groups,
// and the kwds parameter is optionally used for named groups only;
// if a name is not provided in kwds, the value is taken from args, in order.
func (r *ReverseRegexp) Revert(args []string, kwds map[string]string) (string, error) {
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
func (r *ReverseRegexp) ValidRevert(args []string, kwds map[string]string) (string, error) {
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

// writeTemplate writes a reverse template for a regexp to the buffer.
func writeTemplate(b *bytes.Buffer, groups *[]string, re *syntax.Regexp) {
	switch re.Op {
	case syntax.OpLiteral:
		for _, r := range re.Rune {
			b.WriteRune(r)
			if r == '%' {
				b.WriteRune('%')
			}
		}
	case syntax.OpCapture:
		*groups = append(*groups, re.Name)
		b.WriteString("%s")
	case syntax.OpConcat:
		for _, sub := range re.Sub {
			writeTemplate(b, groups, sub)
		}
	}
}
