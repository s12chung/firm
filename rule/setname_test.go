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

func TestSetName_ValidateAll(t *testing.T) {
	oneOf := OneOf[string]{Values: []string{"a"}}
	setName := SetName("MyValues", oneOf)

	require.Nil(t, setName.ValidateValue(reflect.ValueOf("a")))

	expected := firm.ErrorMap{"MyValues": oneOf.ErrorMap()[oneOfName]}
	require.Equal(t, expected, setName.ValidateValue(reflect.ValueOf("c")))
}

func TestSetName_MultipleErrorKeys(t *testing.T) {
	setName := SetName("My", multiErrorRule{})
	inner := multiErrorRule{}.ErrorMap()

	expected := firm.ErrorMap{"My-A": inner["A"], "My-B": inner["B"]}
	require.Equal(t, expected, setName.ValidateValue(reflect.ValueOf(0)))
	require.Equal(t, expected, setName.ErrorMap())
}

func TestSetName_TypeCheck(t *testing.T) {
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
			testTypeCheck(t, tc.data, "MyValues", tc.badCondition, SetName("MyValues", OneOf[int]{}))
		})
	}
}

func TestSetName_ErrorMap(t *testing.T) {
	testErrorMap(t, SetName("MyValues", OneOf[string]{Values: []string{"a", "b"}}), "MyValues: value is not one of [\"a\" \"b\"]")
}
