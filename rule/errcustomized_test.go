package rule

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/s12chung/firm"
)

func TestErrCustomized_ValidateValue(t *testing.T) {
	tcs := []struct {
		name      string
		data      any
		customErr func(firm.ErrorMap) firm.ErrorMap
		errorMap  firm.ErrorMap
	}{
		{
			name: "rule_passes", data: 1,
			customErr: func(firm.ErrorMap) firm.ErrorMap { return firm.ErrorMap{"NotCalled": {}} },
		},
		{
			name: "template_change", data: 0,
			customErr: func(m firm.ErrorMap) firm.ErrorMap {
				err := m["A"]
				err.Template += ", a lot"
				m["A"] = err
				return m
			},
			errorMap: firm.ErrorMap{
				"A": firm.TemplateError{Template: "is a, a lot"},
				"B": firm.TemplateError{Template: "is b"},
			},
		},
		{
			name: "rekey_and_delete", data: 0,
			customErr: func(m firm.ErrorMap) firm.ErrorMap {
				delete(m, "B")
				return firm.ErrorMap{"My": m["A"]}
			},
			errorMap: firm.ErrorMap{"My": firm.TemplateError{Template: "is a"}},
		},
		{name: "customErr_passes", data: 0, customErr: func(firm.ErrorMap) firm.ErrorMap { return nil }},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			rule := ErrCustomized{Rule: multiErrorRule{}, CustomErr: tc.customErr}
			require.Equal(t, tc.errorMap, rule.ValidateValue(reflect.ValueOf(tc.data)))
		})
	}
}

func TestErrCustomized_ErrorMapIsolation(t *testing.T) {
	rule := ErrCustomized{
		// Present shares a package-level ErrorMap, so CustomErr mutations must not leak
		Rule: Present{},
		CustomErr: func(m firm.ErrorMap) firm.ErrorMap {
			err := m["Present"]
			err.Template = "mutated"
			m["Present"] = err
			return m
		},
	}

	require.Equal(t, firm.ErrorMap{"Present": firm.TemplateError{Template: "mutated"}}, rule.ValidateValue(reflect.ValueOf("")))
	require.Equal(t, "Present: value is not present", Present{}.ErrorMap().Error())
}

func TestErrCustomized_TypeCheck(t *testing.T) {
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
			customized := ErrCustomized{Rule: OneOf[int]{}, CustomErr: func(m firm.ErrorMap) firm.ErrorMap { return m }}
			testTypeCheck(t, tc.data, "OneOf", tc.badCondition, customized)
		})
	}
}

func TestErrCustomized_ErrorMap(t *testing.T) {
	rule := ErrCustomized{
		Rule: OneOf[string]{Values: []string{"a"}},
		CustomErr: func(m firm.ErrorMap) firm.ErrorMap {
			err := m[oneOfName]
			err.Template = "is none of {{.Values}}"
			return firm.ErrorMap{"MyValues": err}
		},
	}

	testErrorMap(t, rule, `MyValues: value is none of ["a"]`)
	require.Equal(t, rule.ValidateValue(reflect.ValueOf("c")), rule.ErrorMap())
}
