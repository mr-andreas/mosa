package reducer

import (
	"reflect"
	"testing"

	. "github.com/yoshiyaka/mosa/ast"
)

var expressionTests = []struct {
	expression    Value
	expectedValue Value
}{
	{Expression{0, "+", 4, 5}, 9},
	{Expression{0, "-", 4, 5}, -1},
	{Expression{0, "*", 4, 5}, 20},
	{Expression{0, "/", 6, 2}, 3},

	{Expression{0, "<", 4, 5}, true},
	{Expression{0, "<=", 4, 5}, true},
	{Expression{0, ">", 4, 5}, false},
	{Expression{0, ">=", 4, 5}, false},

	{Expression{0, "<", "aa", "ab"}, true},
	{Expression{0, "<=", "ba", "ab"}, false},
	{Expression{0, ">", "ba", "ab"}, true},
	{Expression{0, ">=", "aa", "ab"}, false},

	{Expression{0, "&&", false, false}, false},
	{Expression{0, "&&", false, true}, false},
	{Expression{0, "&&", true, false}, false},
	{Expression{0, "&&", true, true}, true},
	{Expression{0, "||", false, false}, false},
	{Expression{0, "||", false, true}, true},
	{Expression{0, "||", true, false}, true},
	{Expression{0, "||", true, true}, true},

	{Expression{0, "*", Expression{0, "-", 4, 5}, 5}, -5},

	{
		Expression{0, "+", QuotedString("a"), QuotedString("b")},
		QuotedString("ab"),
	},

	{
		Expression{
			0, "+",
			InterpolatedString{0, []interface{}{"a"}},
			InterpolatedString{0, []interface{}{"b"}},
		},
		QuotedString("ab"),
	},
}

func TestExpressions(t *testing.T) {
	for _, test := range expressionTests {
		ls := newLocalState("test.ms", "test.ms", 0)

		val, err := ls.resolveValue(test.expression, 0)
		if err != nil {
			t.Error("Resolving", test.expression, ", got error:", err.Error())
			continue
		}

		if !reflect.DeepEqual(test.expectedValue, val) {
			t.Error("Got bad value for", test.expression, ": ", val)
		}
	}
}

var badExpressionTests = []struct {
	expression    Expression
	expectedError string
}{
	{Expression{0, "+", 4, "string"}, "Bad types supplied for operation '+' at t.ms:0"},
	{Expression{0, "-", 4, "string"}, "Bad types supplied for operation '-' at t.ms:0"},
	{Expression{0, "*", 4, "string"}, "Bad types supplied for operation '*' at t.ms:0"},
	{Expression{0, "/", 4, "string"}, "Bad types supplied for operation '/' at t.ms:0"},

	{Expression{0, "*", "s1", "s2"}, "Bad types supplied for operation '*' at t.ms:0"},
	{Expression{0, "/", "s1", "s2"}, "Bad types supplied for operation '/' at t.ms:0"},
}

func TestBadExpressions(t *testing.T) {
	for _, test := range badExpressionTests {
		ls := newLocalState("t.ms", "t.ms", 0)

		_, err := ls.resolveValue(test.expression, 0)
		if err == nil || err.Error() != test.expectedError {
			t.Error("Resolving", test.expression, ", got bad error:", err)
			continue
		}
	}
}
