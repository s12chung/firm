package rule

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/s12chung/firm"
)

const includedName = "Included"

// Included checks if data is one of .In
type Included[T comparable] struct{ In []T }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (i Included[T]) ValidateValue(value reflect.Value) firm.ErrorMap {
	data, ok := reflect.TypeAssert[T](value)
	if !ok {
		panic("Included ValidateValue type not matching type--called before TypeCheck?")
	}
	return i.Validate(data)
}

// Validate validates the data value
func (i Included[T]) Validate(data T) firm.ErrorMap {
	if slices.Contains(i.In, data) {
		return nil
	}
	return i.ErrorMap()
}

// TypeCheck checks whether the type is valid for the Rule
func (i Included[T]) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	typFor := reflect.TypeFor[T]()
	if typFor == typ {
		return nil
	}
	return firm.NewRuleTypeError(includedName, typ, "is not a "+typFor.String())
}

// ErrorMap returns the ErrorMap returned from ValidateValue
func (i Included[T]) ErrorMap() firm.ErrorMap {
	return firm.ErrorMap{includedName: firm.TemplateError{
		TemplateFields: map[string]string{"Values": fmt.Sprintf("%v", i.In)},
		Template:       "is not one of {{.Values}}",
	}}
}
