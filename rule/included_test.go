package rule

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/s12chung/firm"
)

func TestIncluded_ValidateAll(t *testing.T) {
	tcs := []struct {
		name string
		in   []string

		data     string
		hasError bool
	}{
		{name: "in_set", in: []string{"a", "b"}, data: "a"},
		{name: "in_set_last", in: []string{"a", "b"}, data: "b"},
		{name: "not_in_set", in: []string{"a", "b"}, data: "c", hasError: true},
		{name: "empty_in", in: []string{}, data: "a", hasError: true},
		{name: "nil_in", data: "a", hasError: true},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			rule := Included[string]{In: tc.in}
			var expected firm.ErrorMap
			if tc.hasError {
				expected = rule.ErrorMap()
			}
			require.Equal(t, expected, rule.Validate(tc.data))
			require.Equal(t, expected, rule.ValidateValue(reflect.ValueOf(tc.data)))
		})
	}
}

func TestIncluded_TypeCheck(t *testing.T) {
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
			testTypeCheck(t, tc.data, includedName, tc.badCondition, Included[int]{})
		})
	}
}

func TestIncluded_ErrorMap(t *testing.T) {
	testErrorMap(t, Included[string]{In: []string{"a", "b"}}, "Included: value is not one of [a b]")
	testErrorMap(t, Included[int]{In: []int{1, 2, 3}}, "Included: value is not one of [1 2 3]")
}
