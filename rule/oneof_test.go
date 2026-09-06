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
			rule := OneOf[string]{Values: tc.values}
			var expected firm.ErrorMap
			if tc.hasError {
				expected = rule.ErrorMap()
			}
			require.Equal(t, expected, rule.Validate(tc.data))
			require.Equal(t, expected, rule.ValidateValue(reflect.ValueOf(tc.data)))
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

func TestOneOf_ErrorMap(t *testing.T) {
	testErrorMap(t, OneOf[string]{Values: []string{"a", "b"}}, "OneOf: value is not one of [a b]")
	testErrorMap(t, OneOf[int]{Values: []int{1, 2, 3}}, "OneOf: value is not one of [1 2 3]")
}
