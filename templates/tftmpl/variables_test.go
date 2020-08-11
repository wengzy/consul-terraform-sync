package tftmpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func TestVariableTypeString(t *testing.T) {
	testCases := []struct {
		name     string
		val      cty.Value
		vType    cty.Type
		expected string
	}{
		{
			"unknown",
			cty.UnknownAsNull(cty.Value{}),
			cty.UnknownAsNull(cty.Value{}).Type(),
			"unknown",
		}, {
			"null bool",
			cty.NullVal(cty.Bool),
			cty.Bool,
			"bool",
		}, {
			"bool",
			cty.BoolVal(true),
			cty.Bool,
			"bool",
		}, {
			"null string",
			cty.NullVal(cty.String),
			cty.String,
			"string",
		}, {
			"string",
			cty.StringVal("this is a string"),
			cty.String,
			"string",
		}, {
			"null number",
			cty.NullVal(cty.Number),
			cty.Number,
			"number",
		}, {
			"number",
			cty.NumberFloatVal(12.34),
			cty.Number,
			"number",
		}, {
			"null list",
			cty.UnknownAsNull(cty.Value{}),
			cty.List(cty.String),
			"list(any)",
		}, {
			"empty list",
			cty.ListValEmpty(cty.Object(nil)),
			cty.List(cty.Object(nil)),
			"list(any)",
		}, {
			"list",
			cty.ListVal([]cty.Value{cty.True, cty.False, cty.True}),
			cty.List(cty.Bool),
			"list(any)",
		}, {
			"null set",
			cty.UnknownAsNull(cty.Value{}),
			cty.Set(cty.String),
			"set(any)",
		}, {
			"empty set",
			cty.SetValEmpty(cty.Object(nil)),
			cty.Set(cty.Object(nil)),
			"set(any)",
		}, {
			"set",
			cty.SetVal([]cty.Value{cty.True, cty.False, cty.True}),
			cty.Set(cty.Bool),
			"set(any)",
		}, {
			"null map",
			cty.UnknownAsNull(cty.Value{}),
			cty.Map(cty.String),
			"map(any)",
		}, {
			"empty map",
			cty.MapValEmpty(cty.Object(nil)),
			cty.Map(cty.Object(nil)),
			"map(any)",
		}, {
			"map",
			cty.MapVal(map[string]cty.Value{"a": cty.True, "b": cty.False, "c": cty.True}),
			cty.Map(cty.Bool),
			"map(any)",
		}, {
			"null tuple",
			cty.UnknownAsNull(cty.Value{}),
			cty.Tuple(nil),
			"tuple([])",
		}, {
			"empty tuple",
			cty.TupleVal([]cty.Value{}),
			cty.EmptyTuple,
			"tuple([])",
		}, {
			"tuple",
			cty.TupleVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.NumberIntVal(1),
				cty.False,
				cty.ListVal([]cty.Value{cty.NumberIntVal(2), cty.NumberIntVal(3)})}),
			cty.Tuple([]cty.Type{cty.String, cty.String, cty.Number, cty.Bool, cty.List(cty.Number)}),
			"tuple([string, string, number, bool, list(any)])",
		}, {
			"null object",
			cty.UnknownAsNull(cty.Value{}),
			cty.Object(map[string]cty.Type{}),
			"object({})",
		}, {
			"empty object",
			cty.ObjectVal(nil),
			cty.EmptyObject,
			"object({})",
		}, {
			"object",
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.True,
				"b": cty.NumberIntVal(1),
				"c": cty.ListVal([]cty.Value{cty.StringVal("x"), cty.StringVal("y"), cty.StringVal("z")}),
			}),
			cty.Object(map[string]cty.Type{
				"a": cty.Bool,
				"b": cty.Number,
				"c": cty.List(cty.String),
			}),
			`object({
a = bool
b = number
c = list(any)
})`,
		}, {
			"nested object",
			cty.ObjectVal(map[string]cty.Value{
				"tup": cty.TupleVal([]cty.Value{
					cty.StringVal("a"),
					cty.StringVal("b"),
					cty.NumberIntVal(1),
				}),
				"obj": cty.ObjectVal(map[string]cty.Value{
					"a": cty.True,
					"b": cty.NumberIntVal(1),
					"c": cty.ListVal([]cty.Value{cty.StringVal("x"), cty.StringVal("y"), cty.StringVal("z")}),
				}),
			}),
			cty.Object(map[string]cty.Type{
				"tup": cty.Tuple([]cty.Type{cty.String, cty.String, cty.Number}),
				"obj": cty.Object(map[string]cty.Type{
					"a": cty.Bool,
					"b": cty.Number,
					"c": cty.List(cty.String),
				}),
			}),
			`object({
obj = object({
a = bool
b = number
c = list(any)
})
tup = tuple([string, string, number])
})`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := variableTypeString(tc.val, tc.vType)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
