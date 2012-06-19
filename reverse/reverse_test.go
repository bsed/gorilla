package reverse

import (
	"testing"
)

type reverseTest struct {
	pattern       string
	validArgs     []string
	validKwds     map[string]string
	validResult   string
	invalidArgs   []string
	invalidKwds   map[string]string
	invalidResult string
}

var reverseTests = []reverseTest{
	reverseTest{
		pattern:       `^1(\d+)3$`,
		validArgs:     []string{"2"},
		validKwds:     nil,
		validResult:   "123",
		invalidArgs:   []string{"a"},
		invalidKwds:   nil,
		invalidResult: "1a3",
	},
	reverseTest{
		pattern:       `^4(?P<foo>\d+)6$`,
		validArgs:     nil,
		validKwds:     map[string]string{"foo": "5"},
		validResult:   "456",
		invalidArgs:   nil,
		invalidKwds:   map[string]string{"foo": "b"},
		invalidResult: "4b6",
	},
	reverseTest{
		pattern:       `^7(?P<foo>\d+)(\d+)0$`,
		validArgs:     []string{"9"},
		validKwds:     map[string]string{"foo": "8"},
		validResult:   "7890",
		invalidArgs:   []string{"d"},
		invalidKwds:   map[string]string{"foo": "c"},
		invalidResult: "7cd0",
	},
	reverseTest{
		pattern:       `(?P<foo>\d+)`,
		validArgs:     []string{"1"},
		validKwds:     nil,
		validResult:   "1",
		invalidArgs:   []string{"a"},
		invalidKwds:   nil,
		invalidResult: "a",
	},
}

func TestReverseRegexp(t *testing.T) {
	for _, test := range reverseTests {
		r, err := Compile(test.pattern)
		if err != nil {
			t.Fatal(err)
		}
		reverse, err := r.Revert(test.validArgs, test.validKwds)
		if err != nil {
			t.Fatal(err)
		}
		if reverse != test.validResult {
			t.Errorf("Expected %q, got %q", test.validResult, reverse)
		}

		reverse, err = r.ValidRevert(test.validArgs, test.validKwds)
		if err != nil {
			t.Errorf("Expected valid %q", test.pattern)
		}

		reverse, err = r.Revert(test.invalidArgs, test.invalidKwds)
		if err != nil {
			t.Fatal(err)
		}
		if reverse != test.invalidResult {
			t.Errorf("Expected %q, got %q", test.invalidResult, reverse)
		}

		reverse, err = r.ValidRevert(test.invalidArgs, test.invalidKwds)
		if err == nil {
			t.Errorf("Expected error for %q", test.pattern)
		}
	}
}

type groupTest struct {
	pattern string
	groups  []string
	indices []int
}

var groupTests = []groupTest{
	groupTest{
		pattern: `^1(\d+)3$`,
		groups:  []string{""},
		indices: []int{1},
	},
	groupTest{
		pattern: `^1(\d+([a-z]+)(\d+([a-z]+)))(?P<foo>\d+)3([a-z]+(\d+))(?P<bar>\d+)$`,
		groups:  []string{"", "foo", "", "bar"},
		indices: []int{1, 5, 6, 8},
	},
}

func TestGroups(t *testing.T) {
	for _, test := range groupTests {
		r, err := Compile(test.pattern)
		if err != nil {
			t.Fatal(err)
		}
		groups, indices := r.Groups()
		if !stringSliceEqual(test.groups, groups) {
			t.Errorf("Expected %v, got %v", test.groups, groups)
		}
		if !intSliceEqual(test.indices, indices) {
			t.Errorf("Expected %v, got %v", test.indices, indices)
		}
	}
}

func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if v != b[k] {
			return false
		}
	}
	return true
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if v != b[k] {
			return false
		}
	}
	return true
}
