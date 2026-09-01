package firm_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/s12chung/firm"
	"github.com/s12chung/firm/rule"
)

type nonExport struct {
	privateChild

	Public  string
	private string
}

type privateChild struct {
	privateC string
}

var notEmpty = nonExport{Public: "not_empty", private: "not_empty", privateChild: privateChild{privateC: "not_empty"}}

func errorMap(field, name firm.ErrorKey, template string) firm.ErrorMap {
	if name == "" && template == "" {
		name = "TrimPresent"
		template = "is just spaces or empty"
	}
	key := "firm_test.nonExport." + field + "." + name
	return firm.ErrorMap{key: firm.TemplateError{Template: template, ErrorKey: key}}
}

func TestFieldsWithErrPkg(t *testing.T) {
	tcs := []struct {
		name           string
		ruleMap        firm.RuleMap
		failErr        error
		constructorErr error
	}{
		{name: "exported_field", ruleMap: firm.RuleMap{"Public": {rule.TrimPresent{}}},
			failErr: errorMap("Public", "", "")},
		{name: "non_exported_field", ruleMap: firm.RuleMap{"private": {rule.TrimPresent{}}},
			constructorErr: errors.New("Fields: field, private, is unexported in type: firm_test.nonExport")},
		{name: "non_exported_child", ruleMap: firm.RuleMap{"privateChild": {rule.Present{}}},
			constructorErr: errors.New("Fields: field, privateChild, is unexported in type: firm_test.nonExport")},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			validator, err := firm.FieldsWithErr[nonExport](tc.ruleMap)
			if tc.constructorErr != nil {
				require.Equal(tc.constructorErr, err)
				return
			}
			require.NoError(err)
			require.Nil(validator.ValidateAny(notEmpty))
			require.Equal(tc.failErr, validator.ValidateAny(nonExport{}))
		})
	}
}

// positiveVldr is a custom firm.Validator implemented with only firm's public impl helpers
type positiveVldr struct {
	firm.Rule
}

var (
	_ firm.Validator           = positiveVldr{}
	_ firm.ValidatorTyped[int] = positiveVldr{}
)

func (p positiveVldr) ValidateAny(data any) firm.ErrorMap { return firm.ImplValidateAny(p, data) }

func (p positiveVldr) ValidateValue(value reflect.Value) firm.ErrorMap {
	return firm.ImplValidateValue(p, value)
}

func (p positiveVldr) ValidateMerge(value reflect.Value, key string, errorMap firm.ErrorMap) {
	firm.ImplValidateMerge(value, key, errorMap, []firm.Rule{p.Rule})
}

func (p positiveVldr) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	return firm.TypeCheck("positiveVldr", typ, reflect.TypeFor[int](), "")
}

func (p positiveVldr) Validate(data int) firm.ErrorMap { return firm.ImplValidate(p, data) }

func keyed(key firm.ErrorKey, templateError firm.TemplateError) firm.TemplateError {
	templateError.ErrorKey = key
	return templateError
}

func TestCustomValidatorPkg(t *testing.T) {
	require := require.New(t)

	v := positiveVldr{rule.Greater[int]{To: 0}}
	greaterErr := firm.TemplateError{
		TemplateFields: map[string]string{"To": "0"},
		Template:       "is not greater than {{.To}}",
	}
	typeCheckErr := firm.TemplateError{
		TemplateFields: map[string]string{"ValueTypeName": "string"},
		Template:       "is not matching type int, got {{.ValueTypeName}}",
	}

	// ImplValidate - typed Validate(), no TypeCheck
	require.Nil(v.Validate(1))
	require.Equal(firm.ErrorMap{"int.Greater": keyed("int.Greater", greaterErr)}, v.Validate(-1))

	// ImplValidateAny - unsafe values are indirected, invalid values and TypeCheck are answered
	require.Nil(v.ValidateAny(1))
	require.Equal(firm.ErrorMap{"int.Greater": keyed("int.Greater", greaterErr)}, v.ValidateAny(-1))

	i := -1
	require.Equal(firm.ErrorMap{"int.Greater": keyed("int.Greater", greaterErr)}, v.ValidateAny(&i))
	require.Equal(firm.ErrorMap{"Invalid": firm.TemplateError{Template: "is not valid"}}, v.ValidateAny(nil))

	var pti *int
	require.Equal(firm.ErrorMap{"Invalid": firm.TemplateError{Template: "is not valid"}}, v.ValidateAny(pti))
	require.Equal(firm.ErrorMap{"TypeCheck": typeCheckErr}, v.ValidateAny("str"))

	// ImplValidateValue - safe value, assumes TypeCheck is called
	require.Equal(firm.ErrorMap{"Greater": keyed("Greater", greaterErr)}, v.ValidateValue(reflect.ValueOf(-1)))
	require.Nil(v.ValidateValue(reflect.ValueOf(1)))
	require.Panics(func() { _ = v.ValidateValue(reflect.Value{}) }) // safe values are expected

	// ImplValidateMerge - safe value, assumes TypeCheck is called
	errorMap := firm.ErrorMap{}
	v.ValidateMerge(reflect.ValueOf(-1), "Positive", errorMap)
	require.Panics(func() { v.ValidateMerge(reflect.Value{}, "Positive", errorMap) }) // safe values are expected
	require.Equal(firm.ErrorMap{
		"Positive.Greater": keyed("Positive.Greater", greaterErr),
	}, errorMap)
}

func TestTypeCheckAndBackPkg(t *testing.T) {
	require := require.New(t)

	rules, err := firm.TypeCheckAndBack(reflect.TypeFor[int](), []firm.Rule{rule.Greater[int]{To: 0}}, "")
	require.NoError(err)
	require.Equal([]firm.Rule{rule.Greater[int]{To: 0}}, rules)

	rules, err = firm.TypeCheckAndBack(reflect.TypeFor[string](), []firm.Rule{rule.Greater[int]{To: 0}}, "")
	require.Nil(rules)
	require.EqualError(err, "Greater: value is not a int, got string")

	rules, err = firm.TypeCheckAndBack(reflect.TypeFor[string](), []firm.Rule{rule.Greater[int]{To: 0}}, "Positive: element type")
	require.Nil(rules)
	require.EqualError(err, "Positive: element type: Greater: value is not a int, got string")
}
