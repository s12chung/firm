package rule

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/s12chung/firm"
)

func andRules() AndWrap {
	return And(Len{Is: 5}, Match{Regexp: regexp.MustCompile(`^z`)})
}

func TestAnd_ValidateValue(t *testing.T) {
	tcs := []struct {
		name     string
		rule     AndWrap
		data     any
		errorMap firm.ErrorMap
	}{
		{name: "all_rules_pass", data: "zebra", rule: andRules()},
		{
			name: "first_rule_fails", data: "zeb", rule: andRules(),
			errorMap: firm.ErrorMap{"Len": firm.TemplateError{
				TemplateFields: map[string]string{"Is": "5"},
				Template:       "does not have a length of {{.Is}}",
				ErrorKey:       "Len",
			}},
		},
		{
			name: "second_rule_fails", data: "hello", rule: andRules(),
			errorMap: firm.ErrorMap{"Match": firm.TemplateError{
				TemplateFields: map[string]string{"Regexp": "^z"},
				Template:       "does not match {{.Regexp}}",
				ErrorKey:       "Match",
			}},
		},
		{name: "all_rules_fail", data: "abc", rule: andRules(), errorMap: andRules().ErrorMap()},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.errorMap, tc.rule.ValidateValue(reflect.ValueOf(tc.data)))
		})
	}
}

func TestAnd_WithoutRules(t *testing.T) {
	require.PanicsWithValue(t, "And() called without rules", func() { And() })
}

//nolint:dupl // symmetric with TestOr_TypeCheck
func TestAnd_TypeCheck(t *testing.T) {
	stringRules := And(TrimPresent{}, Present{})
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
			rule:         And(Len{Is: 5}, OneOf[int]{Values: []int{1}}),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			testTypeCheck(t, tc.data, tc.ruleName, tc.badCondition, tc.rule)
		})
	}
}

func TestAnd_ErrorMap(t *testing.T) {
	rule := And(Len{Is: 5}, Match{Regexp: regexp.MustCompile(`^z`)})
	testErrorMap(t, rule, `Len: value does not have a length of 5, Match: value does not match ^z`)
	require.Equal(t, rule.ValidateValue(reflect.ValueOf("abc")), rule.ErrorMap())
}
