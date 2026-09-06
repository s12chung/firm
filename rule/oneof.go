package rule

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/s12chung/firm"
)

const oneOfName = "OneOf"

// OneOf checks if data is one of .Values
type OneOf[T comparable] struct{ Values []T }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (o OneOf[T]) ValidateValue(value reflect.Value) firm.ErrorMap {
	data, ok := reflect.TypeAssert[T](value)
	if !ok {
		panic("OneOf ValidateValue type not matching type--called before TypeCheck?")
	}
	return o.Validate(data)
}

// Validate validates the data value
func (o OneOf[T]) Validate(data T) firm.ErrorMap {
	if slices.Contains(o.Values, data) {
		return nil
	}
	return o.ErrorMap()
}

// TypeCheck checks whether the type is valid for the Rule
func (o OneOf[T]) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	typFor := reflect.TypeFor[T]()
	if typFor == typ {
		return nil
	}
	return firm.NewRuleTypeError(oneOfName, typ, "is not a "+typFor.String())
}

// ErrorMap returns the ErrorMap returned from ValidateValue
func (o OneOf[T]) ErrorMap() firm.ErrorMap {
	return firm.ErrorMap{oneOfName: firm.TemplateError{
		TemplateFields: map[string]string{"Values": fmt.Sprintf("%v", o.Values)},
		Template:       "is not one of {{.Values}}",
	}}
}
