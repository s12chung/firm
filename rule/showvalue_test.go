package rule

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/s12chung/firm"
)

func TestShowValue_ValidateValue(t *testing.T) {
	require := require.New(t)

	shown := ShowValue(multiErrorRule{})
	require.Nil(shown.ValidateValue(reflect.ValueOf(1)))
	require.Equal(firm.ErrorMap{
		"A": firm.TemplateError{Template: "is a: {{value}}"},
		"B": firm.TemplateError{Template: "is b: {{value}}"},
	}, shown.ValidateValue(reflect.ValueOf(0)))
}

func TestShowValue_ErrorMapIsolation(t *testing.T) {
	require := require.New(t)

	// Present shares a package-level ErrorMap, so the suffix must not leak into it
	shown := ShowValue(Present{})

	require.Equal(firm.ErrorMap{"Present": firm.TemplateError{Template: "is not present: {{value}}"}},
		shown.ValidateValue(reflect.ValueOf("")))
	require.Equal("Present: value is not present", Present{}.ErrorMap().Error())
}

func TestShowValue_TypeCheck(t *testing.T) {
	shown := ShowValue(OneOf[int]{})
	testTypeCheck(t, 0, "OneOf", "", shown)
	testTypeCheck(t, "", "OneOf", "is not a int", shown)
}

func TestShowValue_validates(t *testing.T) {
	require := require.New(t)

	type config struct{ Str string }
	validator := firm.Fields[config](firm.RuleMap{
		"Str": {ShowValue(OneOf[string]{Values: []string{"a"}})},
	})
	require.Nil(validator.Validate(config{Str: "a"}))

	errMap := validator.Validate(config{Str: "b"})
	require.Len(errMap, 1)
	for _, err := range errMap {
		require.Equal(`Str is not one of ["a"]: "b"`, err.Error())
	}
}
