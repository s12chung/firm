package rule

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/s12chung/firm"
)

type multiErrorRule struct{}

func (m multiErrorRule) ValidateValue(value reflect.Value) firm.ErrorMap {
	if value.Int() != 0 {
		return nil
	}
	return m.ErrorMap()
}

func (m multiErrorRule) ErrorMap() firm.ErrorMap {
	return firm.ErrorMap{
		"A": firm.TemplateError{Template: "is a"},
		"B": firm.TemplateError{Template: "is b"},
	}
}

func (m multiErrorRule) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	if typ.Kind() == reflect.Int {
		return nil
	}
	return firm.NewRuleTypeError("multiErrorRule", typ, "is not a int")
}

func TestNamed_ValidateAll(t *testing.T) {
	oneOf := OneOf[string]{Values: []string{"a"}}
	named := Named{Name: "MyValues", Rule: oneOf}

	require.Nil(t, named.ValidateValue(reflect.ValueOf("a")))

	expected := firm.ErrorMap{"MyValues": oneOf.ErrorMap()[oneOfName]}
	require.Equal(t, expected, named.ValidateValue(reflect.ValueOf("c")))
}

func TestNamed_MultipleErrorKeys(t *testing.T) {
	named := Named{Name: "My", Rule: multiErrorRule{}}
	inner := multiErrorRule{}.ErrorMap()

	expected := firm.ErrorMap{"My-A": inner["A"], "My-B": inner["B"]}
	require.Equal(t, expected, named.ValidateValue(reflect.ValueOf(0)))
	require.Equal(t, expected, named.ErrorMap())
}

func TestNamed_TypeCheck(t *testing.T) {
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
			testTypeCheck(t, tc.data, "MyValues", tc.badCondition, Named{Name: "MyValues", Rule: OneOf[int]{}})
		})
	}
}

func TestNamed_ErrorMap(t *testing.T) {
	testErrorMap(t, Named{Name: "MyValues", Rule: OneOf[string]{Values: []string{"a", "b"}}}, "MyValues: value is not one of [a b]")
}
