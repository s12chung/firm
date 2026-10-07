package rule

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/s12chung/firm"
)

func orRules() OrWrap {
	return Or(Len{Is: 5}, Match{Regexp: regexp.MustCompile(`^z`)})
}

func TestOr_ValidateValue(t *testing.T) {
	tcs := []struct {
		name     string
		rule     OrWrap
		data     any
		errorMap firm.ErrorMap
	}{
		{name: "first_rule", data: "hello", rule: orRules()},
		{name: "second_rule", data: "zup", rule: orRules()},
		{name: "all_rules_fail", data: "abc", rule: orRules(), errorMap: orRules().ErrorMap()},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.errorMap, tc.rule.ValidateValue(reflect.ValueOf(tc.data)))
		})
	}
}

func TestOr_WithoutRules(t *testing.T) {
	require.PanicsWithValue(t, "Or() called without rules", func() { Or() })
}

//nolint:dupl // symmetric with TestAnd_TypeCheck
func TestOr_TypeCheck(t *testing.T) {
	stringRules := Or(TrimPresent{}, Present{})
	tcs := []struct {
		name         string
		data         any
		badCondition string
		ruleName     string
		rule         firm.Rule
	}{
		{name: "string", data: "", ruleName: "TrimPresent", rule: stringRules},
		{name: "not string", data: 1, badCondition: "is not a String", ruleName: "TrimPresent", rule: stringRules},
		{
			name:         "not every rule type passes type check",
			data:         "",
			badCondition: "is not a int",
			ruleName:     "OneOf",
			rule:         Or(Len{Is: 5}, OneOf[int]{Values: []int{1}}),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			testTypeCheck(t, tc.data, tc.ruleName, tc.badCondition, tc.rule)
		})
	}
}

func TestOr_ErrorMap(t *testing.T) {
	rule := Or(Len{Is: 5}, Match{Regexp: regexp.MustCompile(`^z`)})
	testErrorMap(t, rule, `Or: value is not any of [Len: value does not have a length of 5; Match: value does not match ^z]`)
	require.Equal(t, rule.ValidateValue(reflect.ValueOf("abc")), rule.ErrorMap())
}
