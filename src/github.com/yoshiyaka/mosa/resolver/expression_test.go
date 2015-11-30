package resolver

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

	{Expression{0, "<", 4, 5}, Bool(true)},
	{Expression{0, "<=", 4, 5}, Bool(true)},
	{Expression{0, ">", 4, 5}, Bool(false)},
	{Expression{0, ">=", 4, 5}, Bool(false)},

	{Expression{0, "<", "aa", "ab"}, Bool(true)},
	{Expression{0, "<=", "ba", "ab"}, Bool(false)},
	{Expression{0, ">", "ba", "ab"}, Bool(true)},
	{Expression{0, ">=", "aa", "ab"}, Bool(false)},

	{Expression{0, "&&", Bool(false), Bool(false)}, Bool(false)},
	{Expression{0, "&&", Bool(false), Bool(true)}, Bool(false)},
	{Expression{0, "&&", Bool(true), Bool(false)}, Bool(false)},
	{Expression{0, "&&", Bool(true), Bool(true)}, Bool(true)},
	{Expression{0, "||", Bool(false), Bool(false)}, Bool(false)},
	{Expression{0, "||", Bool(false), Bool(true)}, Bool(true)},
	{Expression{0, "||", Bool(true), Bool(false)}, Bool(true)},
	{Expression{0, "||", Bool(true), Bool(true)}, Bool(true)},

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
			t.Errorf("Got bad value for %v: %v", test.expression, val)
		}
	}
}

var badExpressionTests = []struct {
	expression    Expression
	expectedError string
}{
	{Expression{0, "+", 4, "string"}, "Bad types (int, string) supplied for operation '+' at t.ms:0"},
	{Expression{0, "-", 4, "string"}, "Bad types (int, string) supplied for operation '-' at t.ms:0"},
	{Expression{0, "*", 4, "string"}, "Bad types (int, string) supplied for operation '*' at t.ms:0"},
	{Expression{0, "/", 4, "string"}, "Bad types (int, string) supplied for operation '/' at t.ms:0"},

	{Expression{0, "*", "s1", "s2"}, "Bad types (string, string) supplied for operation '*' at t.ms:0"},
	{Expression{0, "/", "s1", "s2"}, "Bad types (string, string) supplied for operation '/' at t.ms:0"},
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
