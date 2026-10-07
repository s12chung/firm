package firm

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorMap_Error(t *testing.T) {
	require := require.New(t)

	errorMap := ErrorMap{
		"A": TemplateError{Template: "field A message"},
		"B": TemplateError{Template: "field B message"},
	}
	require.Equal("A: value field A message, B: value field B message", errorMap.Error())
}

func TestErrorMap_ErrorWith(t *testing.T) {
	require := require.New(t)

	errorMap := ErrorMap{
		"A": TemplateError{Template: "field A message"},
		"B": TemplateError{Template: "field B message"},
	}
	require.Equal("A: value：field A message, B: value：field B message",
		errorMap.ErrorWith("{{.ValueName}}：", ""))
	require.Equal("A: value field A message!, B: value field B message!",
		errorMap.ErrorWith("{{.ValueName}} ", "!"))
}

func TestErrorMap_Merge(t *testing.T) {
	require := require.New(t)

	dest := ErrorMap{
		"PATH.B": TemplateError{Template: "b2"},
		"PATH.C": TemplateError{Template: "c2"},
	}
	dest.Merge("PATH", ErrorMap{
		"A": TemplateError{Template: "a1"},
		"B": TemplateError{Template: "b1"},
	})
	require.Equal(ErrorMap{
		"PATH.A": TemplateError{Template: "a1", ErrorKey: "PATH.A"},
		"PATH.B": TemplateError{Template: "b1", ErrorKey: "PATH.B"},
		"PATH.C": TemplateError{Template: "c2"},
	}, dest)
}

func TestErrorMap_Clone(t *testing.T) {
	require := require.New(t)

	original := ErrorMap{"A": TemplateError{Template: "a", TemplateFields: map[string]string{"Field": "value"}}}
	clone := original.Clone()

	require.Equal(original, clone)

	templateError := clone["A"]
	templateError.Template = "mutated"
	templateError.TemplateFields["Field"] = "mutated"
	clone["A"] = templateError
	require.Equal(ErrorMap{"A": TemplateError{Template: "a", TemplateFields: map[string]string{"Field": "value"}}}, original)
}

func TestErrorMap_withValue(t *testing.T) {
	require := require.New(t)

	shared := ErrorMap{
		"Captured":    {Template: "captured"},
		"NotCaptured": {Template: "not captured"},
	}
	capturedValue := reflect.ValueOf("captured")
	captured := shared["Captured"]
	captured.value = capturedValue
	shared["Captured"] = captured

	value := reflect.ValueOf("value")
	withValue := shared.withValue(value)

	require.Equal(value.Interface(), withValue["NotCaptured"].value.Interface())
	// the innermost capture wins--outer validators do not clobber it
	require.Equal(capturedValue.Interface(), withValue["Captured"].value.Interface())
	// the shared rule ErrorMap is not mutated
	require.False(shared["NotCaptured"].value.IsValid())
}

func TestErrorMap_ToNil(t *testing.T) {
	tcs := []struct {
		name     string
		errorMap ErrorMap
		isNil    bool
	}{
		{name: "not_empty", errorMap: ErrorMap{"testy": TemplateError{}}},
		{name: "empty", errorMap: ErrorMap{}, isNil: true},
		{name: "nil", errorMap: nil, isNil: true},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			expected := tc.errorMap
			if tc.isNil {
				expected = nil
			}
			require.Equal(expected, tc.errorMap.ToNil())
		})
	}
}

var fullTemplateErrorKey = ErrorKey("pkger.Mover.Parent.MyField.TheError")
var errTemplateError = TemplateError{
	ErrorKey: fullTemplateErrorKey,
	Template: "has no {{ .Him }} and {{ .Her }} since it's of type: {{.RootTypeName}}",
	TemplateFields: map[string]string{
		"Him": "Jack",
		"Her": "Jill",
	},
}

func fullTemplate() TemplateError { return errTemplateError }

func TestTemplateError_Error(t *testing.T) {
	tcs := []struct {
		name     string
		template string
		fields   map[string]string
		expected string
	}{
		{name: "everything", expected: "MyField has no Jack and Jill since it's of type: pkger.Mover"},
		{name: "empty_error_key", expected: "value has no Jack and Jill since it's of type: NoType"},
		{name: "missing_one", expected: "MyField has no <no value> and Jill since it's of type: pkger.Mover", fields: map[string]string{
			"Her": "Jill",
		}},
		{name: "bad_template", template: "{{ a }}", expected: "{{ a }} (bad format)"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			templateError := fullTemplate()
			if tc.name == "empty_error_key" {
				templateError.ErrorKey = ""
			}
			if tc.fields != nil {
				templateError.TemplateFields = tc.fields
			}
			if tc.template != "" {
				templateError.Template = tc.template
			}

			require.Equal(tc.expected, templateError.Error())
		})
	}
}

func TestTemplateError_ErrorWith(t *testing.T) {
	tcs := []struct {
		name     string
		prefix   string
		suffix   string
		template string
		expected string
	}{
		{name: "default", prefix: "{{.ValueName}} ", expected: "MyField has no Jack and Jill since it's of type: pkger.Mover"},
		{name: "custom_separator", prefix: "{{.ValueName}}：", expected: "MyField：has no Jack and Jill since it's of type: pkger.Mover"},
		{name: "other_fields", prefix: "[{{.RootTypeName}}] ", expected: "[pkger.Mover] has no Jack and Jill since it's of type: pkger.Mover"},
		{name: "suffix", prefix: "{{.ValueName}} ", suffix: " (like {{ .Him }})",
			expected: "MyField has no Jack and Jill since it's of type: pkger.Mover (like Jack)"},
		{name: "none", expected: "has no Jack and Jill since it's of type: pkger.Mover"},
		{name: "subject_last", template: "{{ .Him }}が必要な{{.ValueName}}ではありません", expected: "Jackが必要なMyFieldではありません"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			templateError := fullTemplate()
			if tc.template != "" {
				templateError.Template = tc.template
			}
			require.Equal(tc.expected, templateError.ErrorWith(tc.prefix, tc.suffix))
		})
	}
}

func TestTemplateError_valueString(t *testing.T) {
	tcs := []struct {
		name     string
		value    reflect.Value
		limit    int
		expected string
	}{
		{name: "string", value: reflect.ValueOf("Noun"), expected: `"Noun"`},
		{name: "empty_string", value: reflect.ValueOf(""), expected: `""`},
		{name: "spaces", value: reflect.ValueOf("  "), expected: `"  "`},
		{name: "capped_string", value: reflect.ValueOf(strings.Repeat("a", 60)), expected: `"` + strings.Repeat("a", valueCap) + `"...`},
		{name: "rune_boundary", value: reflect.ValueOf(strings.Repeat("あ", 20)), limit: 50, expected: `"` + strings.Repeat("あ", 16) + `"...`},
		{name: "custom_limit", value: reflect.ValueOf("Noun"), limit: 2, expected: `"No"...`},
		{name: "int", value: reflect.ValueOf(42), expected: "42"},
		{name: "float", value: reflect.ValueOf(3.5), expected: "3.5"},
		{name: "bool", value: reflect.ValueOf(false), expected: "false"},
		{name: "slice", value: reflect.ValueOf([]string{"a"}), expected: "[]string (len 1)"},
		{name: "map", value: reflect.ValueOf(map[string]int{}), expected: "map[string]int (len 0)"},
		{name: "struct", value: reflect.ValueOf(TemplateError{}), expected: "firm.TemplateError"},
		{name: "nil_ptr", value: reflect.ValueOf((*int)(nil)), expected: "<nil>"},
		{name: "not_captured", value: reflect.Value{}, expected: "<no value>"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			templateError := TemplateError{value: tc.value}
			if tc.limit > 0 {
				require.Equal(tc.expected, templateError.valueString(tc.limit))
			} else {
				require.Equal(tc.expected, templateError.valueString())
			}
		})
	}
}

func TestTemplateError_valueTemplateFunc(t *testing.T) {
	require := require.New(t)

	templateError := TemplateError{
		ErrorKey:       fullTemplateErrorKey,
		Template:       "is not one of {{.Values}}: {{value}}",
		TemplateFields: map[string]string{"Values": "Noun, Verb"},
		value:          reflect.ValueOf("Noun"),
	}
	require.Equal(`MyField is not one of Noun, Verb: "Noun"`, templateError.Error())
	require.Equal(`MyField is not one of Noun, Verb: "Noun" (got "No"...)`,
		templateError.ErrorWith(valueNamePrefix, " (got {{value 2}})"))
	require.Equal(`MyField is nil: <no value>`,
		TemplateError{ErrorKey: fullTemplateErrorKey, Template: "is nil: {{value}}"}.Error())
}

func TestErrorKey_RootTypeName(t *testing.T) {
	tcs := []struct {
		name     string
		errorKey ErrorKey
		expected string
	}{
		{name: "deep", errorKey: "firm.parent.Field.[0].InnerField.TheError", expected: "firm.parent"},
		{name: "field", errorKey: "firm.parent.Field.TheError", expected: "firm.parent"},
		{name: "field", errorKey: "firm.parent.Field.TheError", expected: "firm.parent"},
		{name: "self", errorKey: "firm.parent.TheError", expected: "firm.parent"},
		{name: "map_key_separator", errorKey: "firm.parent.Map.[a.b].TheError", expected: "firm.parent"},
		{name: "composite_root", errorKey: "map[string]firm.Child.[k].TheError", expected: "map[string]firm.Child"},
		{name: "empty", errorKey: "", expected: ""},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expected, tc.errorKey.RootTypeName())
		})
	}
}

func TestErrorKey_ValueName(t *testing.T) {
	tcs := []struct {
		name     string
		errorKey ErrorKey
		expected string
	}{
		{name: "deep", errorKey: "firm.parent.Field.[0].InnerField.TheError", expected: "InnerField"},
		{name: "slice", errorKey: "firm.parent.Field.[0].TheError", expected: "Field[0]"},
		{name: "nested_slice", errorKey: "firm.parent.Field.[0].[1].TheError", expected: "Field[0][1]"},
		{name: "field", errorKey: "firm.parent.Field.TheError", expected: "Field"},
		{name: "self", errorKey: "firm.parent.TheError", expected: "firm.parent"},
		{name: "map_key_separator", errorKey: "firm.parent.Map.[a.b].TheError", expected: "Map[a.b]"},
		{name: "composite_root", errorKey: "map[string]firm.Child.[k].TheError", expected: "[k]"},
		{name: "just_type", errorKey: "firm.parent", expected: ""},
		{name: "empty", errorKey: "", expected: ""},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expected, tc.errorKey.ValueName())
		})
	}
}

func TestErrorKey_ErrorName(t *testing.T) {
	tcs := []struct {
		name     string
		errorKey ErrorKey
		expected string
	}{
		{name: "deep", errorKey: "firm.parent.Field.[0].InnerField.TheError", expected: "TheError"},
		{name: "field", errorKey: "firm.parent.Field.TheError", expected: "TheError"},
		{name: "self", errorKey: "firm.parent.TheError", expected: "TheError"},
		{name: "map_key_separator", errorKey: "firm.parent.Map.[a.b].TheError", expected: "TheError"},
		{name: "empty", errorKey: "", expected: ""},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expected, tc.errorKey.ErrorName())
		})
	}
}

func TestRuleTypeError_TemplateError(t *testing.T) {
	require := require.New(t)
	require.Equal("value is not a string, got int",
		NewRuleTypeError("MyRule", reflect.TypeFor[int](), "is not a string").TemplateError().Error())
}

func TestRuleTypeError_Error(t *testing.T) {
	require := require.New(t)
	require.Equal("MyRule: value is not a string, got int",
		NewRuleTypeError("MyRule", reflect.TypeFor[int](), "is not a string").Error())
}
