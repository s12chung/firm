package rule

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"

	"github.com/s12chung/firm"
)

const oneOfName = "OneOf"

// OneOf checks if data is one of .Values or the result of .ValuesFunc; both must not be set
type OneOf[T comparable] struct {
	Values     []T
	ValuesFunc func() []T
}

// ValidateValue validates the data value (assumes TypeCheck is called)
func (o OneOf[T]) ValidateValue(value reflect.Value) firm.ErrorMap {
	data, ok := reflect.TypeAssert[T](value)
	if !ok {
		panic("OneOf ValidateValue type not matching type--called before TypeCheck?")
	}
	return o.Validate(data)
}

func (o OneOf[T]) values() []T {
	if o.ValuesFunc == nil {
		return o.Values
	}
	return o.ValuesFunc()
}

// Validate validates the data value
func (o OneOf[T]) Validate(data T) firm.ErrorMap {
	if slices.Contains(o.values(), data) {
		return nil
	}
	return o.ErrorMap()
}

// TypeCheck checks whether the type is valid for the Rule
func (o OneOf[T]) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	if o.Values != nil && o.ValuesFunc != nil {
		return firm.NewRuleTypeError(oneOfName, typ, "Values and ValuesFunc must not both be set")
	}
	typFor := reflect.TypeFor[T]()
	if typFor == typ {
		return nil
	}
	return firm.NewRuleTypeError(oneOfName, typ, "is not a "+typFor.String())
}

// ErrorMap returns the ErrorMap returned from ValidateValue
func (o OneOf[T]) ErrorMap() firm.ErrorMap {
	return firm.ErrorMap{oneOfName: firm.TemplateError{
		TemplateFields: map[string]string{"Values": valuesStr(o.values())},
		Template:       "is not one of {{.Values}}",
	}}
}

// valuesStr formats values for the error message; string values are quoted,
// otherwise an empty string won't show up
func valuesStr[T comparable](values []T) string {
	strs := make([]string, len(values))
	for i, v := range values {
		if s, ok := any(v).(string); ok {
			strs[i] = strconv.Quote(s)
		} else {
			strs[i] = fmt.Sprintf("%v", v)
		}
	}
	return fmt.Sprintf("%v", strs)
}
