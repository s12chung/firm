package rule

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/s12chung/firm"
)

func TestLen_ValidateAll(t *testing.T) {
	rule := Len{Min: 2, Max: 3}
	tcs := []struct {
		name     string
		data     string
		hasError bool
	}{
		{name: "at_min", data: "ab"},
		{name: "at_max", data: "abc"},
		{name: "too_short", data: "a", hasError: true},
		{name: "too_long", data: "abcd", hasError: true},
		{name: "empty", data: "", hasError: true},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var expected firm.ErrorMap
			if tc.hasError {
				expected = rule.ErrorMap()
			}
			require.Equal(t, expected, rule.ValidateValue(reflect.ValueOf(tc.data)))
		})
	}
}

func TestLen_ValidateAll_Kinds(t *testing.T) {
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	tcs := []struct {
		name string
		data any
	}{
		{name: "string", data: "ab"},
		{name: "slice", data: []int{1, 2}},
		{name: "array", data: [2]int{}},
		{name: "map", data: map[string]int{"a": 1, "b": 2}},
		{name: "chan", data: ch},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require.Nil(t, Len{Min: 2}.ValidateValue(reflect.ValueOf(tc.data)))
		})
	}
}

func TestLen_ValidateAll_Constraints(t *testing.T) {
	tcs := []struct {
		name     string
		rule     Len
		length   int
		hasError bool
	}{
		{name: "is_match", rule: Len{Is: 3}, length: 3},
		{name: "is_no_match", rule: Len{Is: 3}, length: 2, hasError: true},
		{name: "is_unset", rule: Len{Is: 0}, length: 100},
		{name: "min_ok", rule: Len{Min: 2}, length: 2},
		{name: "min_fail", rule: Len{Min: 2}, length: 1, hasError: true},
		{name: "max_ok", rule: Len{Max: 2}, length: 2},
		{name: "max_fail", rule: Len{Max: 2}, length: 3, hasError: true},
		{name: "range_at_min", rule: Len{Min: 2, Max: 4}, length: 2},
		{name: "range_at_max", rule: Len{Min: 2, Max: 4}, length: 4},
		{name: "range_below", rule: Len{Min: 2, Max: 4}, length: 1, hasError: true},
		{name: "range_above", rule: Len{Min: 2, Max: 4}, length: 5, hasError: true},
		{name: "unset", rule: Len{}, length: 100},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var expected firm.ErrorMap
			if tc.hasError {
				expected = tc.rule.ErrorMap()
			}
			data := strings.Repeat("a", tc.length)
			require.Equal(t, expected, tc.rule.ValidateValue(reflect.ValueOf(data)))
		})
	}
}

func TestLen_TypeCheck(t *testing.T) {
	ch := make(chan int, 1)
	badCondition := "does not have a length (not a Slice, Array, Channel, Map or String)"

	tcs := []struct {
		name         string
		data         any
		badCondition string
	}{
		{name: "string", data: ""},
		{name: "slice", data: []int{}},
		{name: "array", data: [0]int{}},
		{name: "map", data: map[string]int{}},
		{name: "chan", data: ch},
		{name: "int", data: 0, badCondition: badCondition},
		{name: "struct", data: struct{}{}, badCondition: badCondition},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			testTypeCheck(t, tc.data, lenName, tc.badCondition, Len{})
		})
	}
}

func TestLen_ErrorMap(t *testing.T) {
	testErrorMap(t, Len{Is: 6}, "Len: value does not have a length of 6")
	testErrorMap(t, Len{Min: 2}, "Len: value is too short (minimum length is 2)")
	testErrorMap(t, Len{Max: 5}, "Len: value is too long (maximum length is 5)")
	testErrorMap(t, Len{Min: 2, Max: 5}, "Len: value does not have a length between 2 and 5")
}
