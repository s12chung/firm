package rule

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/s12chung/firm"
)

type parent struct {
	Npm        *npm
	VersionURL *versionURL
	Int        int
}

type (
	npm        struct{}
	versionURL struct{}
)

func TestOneNotNil_ValidateAll(t *testing.T) {
	rule := OneNotNil{Fields: []string{"Npm", "VersionURL"}}
	tcs := []struct {
		name     string
		data     parent
		expected firm.ErrorMap
	}{
		{name: "npm", data: parent{Npm: &npm{}}},
		{name: "version_url", data: parent{VersionURL: &versionURL{}}},
		{name: "none_non_nil", data: parent{}, expected: firm.ErrorMap{oneNotNilName: firm.TemplateError{
			Template:       "must have exactly one of {{.Fields}} non-nil, got {{.NotNil}}",
			TemplateFields: map[string]string{"Fields": "[Npm VersionURL]", "NotNil": "[]"},
		}}},
		{name: "all_non_nil", data: parent{Npm: &npm{}, VersionURL: &versionURL{}}, expected: firm.ErrorMap{oneNotNilName: firm.TemplateError{
			Template:       "must have exactly one of {{.Fields}} non-nil, got {{.NotNil}}",
			TemplateFields: map[string]string{"Fields": "[Npm VersionURL]", "NotNil": "[Npm VersionURL]"},
		}}},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, rule.ValidateValue(reflect.ValueOf(tc.data)))
		})
	}
}

func TestOneNotNil_TypeCheck(t *testing.T) {
	tcs := []struct {
		name         string
		data         any
		fields       []string
		badCondition string
	}{
		{name: "npm", data: parent{}, fields: []string{"Npm"}},
		{name: "all_fields", data: parent{}, fields: []string{"Npm", "VersionURL"}},
		{name: "missing_field", data: parent{}, fields: []string{"Missing"}, badCondition: "has no field, Missing"},
		{name: "not_pointer", data: parent{}, fields: []string{"Int"}, badCondition: "has a non-Pointer field, Int"},
		{name: "duplicate_field", data: parent{}, fields: []string{"Npm", "Npm"}, badCondition: "has duplicate Fields, Npm"},
		{name: "no_fields", data: parent{}, fields: nil, badCondition: "has no Fields"},
		{name: "not_struct", data: 0, fields: []string{"Npm"}, badCondition: "is not a Struct"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			testTypeCheck(t, tc.data, oneNotNilName, tc.badCondition, OneNotNil{Fields: tc.fields})
		})
	}
}

func TestOneNotNil_ErrorMap(t *testing.T) {
	testErrorMap(t, OneNotNil{Fields: []string{"Npm", "VersionURL"}}, "OneNotNil: value must have exactly one of [Npm VersionURL] non-nil")
}
