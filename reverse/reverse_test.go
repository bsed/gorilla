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
