package rule

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/s12chung/firm"
)

func TestOneOf_ValidateAll(t *testing.T) {
	tcs := []struct {
		name   string
		values []string

		data     string
		hasError bool
	}{
		{name: "in_values", values: []string{"a", "b"}, data: "a"},
		{name: "in_values_last", values: []string{"a", "b"}, data: "b"},
		{name: "not_in_values", values: []string{"a", "b"}, data: "c", hasError: true},
		{name: "empty_values", values: []string{}, data: "a", hasError: true},
		{name: "nil_values", data: "a", hasError: true},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			rules := map[string]OneOf[string]{
				"Values":     {Values: tc.values},
				"ValuesFunc": {ValuesFunc: func() []string { return tc.values }},
			}
			for name, rule := range rules {
				t.Run(name, func(t *testing.T) {
					var expected firm.ErrorMap
					if tc.hasError {
						expected = rule.ErrorMap()
					}
					require.Equal(t, expected, rule.Validate(tc.data))
					require.Equal(t, expected, rule.ValidateValue(reflect.ValueOf(tc.data)))
				})
			}
		})
	}
}

func TestOneOf_TypeCheck(t *testing.T) {
	i := 0
	badCondition := "is not a int"

	tcs := []struct {
		name         string
		data         any
		badCondition string
	}{
		{name: "matching type", data: 0},
		{name: "matching type pointer", data: &i, badCondition: badCondition},
		{name: "other type", data: "", badCondition: badCondition},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			testTypeCheck(t, tc.data, oneOfName, tc.badCondition, OneOf[int]{})
		})
	}
}

func TestOneOf_BothSet(t *testing.T) {
	bothSet := OneOf[int]{Values: []int{1}, ValuesFunc: func() []int { return []int{2} }}

	testTypeCheck(t, 0, oneOfName, "Values and ValuesFunc must not both be set", bothSet)
}

func TestOneOf_ErrorMap(t *testing.T) {
	valuesFunc := func(values ...string) func() []string {
		return func() []string { return values }
	}
	testErrorMap(t, OneOf[string]{Values: []string{"a", "b"}}, "OneOf: value is not one of [\"a\" \"b\"]")
	testErrorMap(t, OneOf[string]{ValuesFunc: valuesFunc("a", "b")}, "OneOf: value is not one of [\"a\" \"b\"]")
	testErrorMap(t, OneOf[string]{Values: []string{"a", ""}}, "OneOf: value is not one of [\"a\" \"\"]")
	testErrorMap(t, OneOf[string]{ValuesFunc: valuesFunc("a", "")}, "OneOf: value is not one of [\"a\" \"\"]")
	testErrorMap(t, OneOf[int]{Values: []int{1, 2, 3}}, "OneOf: value is not one of [1 2 3]")
	testErrorMap(t, OneOf[int]{ValuesFunc: func() []int { return []int{1, 2, 3} }}, "OneOf: value is not one of [1 2 3]")
}
