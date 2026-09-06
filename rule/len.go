package rule

import (
	"reflect"
	"strconv"

	"github.com/s12chung/firm"
)

const lenName = "Len"

// Len checks if the length of data is .Is or between .Min and .Max (0 = unset)
type Len struct {
	Is       int
	Min, Max int
}

// ValidateValue validates the data value (assumes TypeCheck is called)
func (l Len) ValidateValue(value reflect.Value) firm.ErrorMap {
	if l.valid(value.Len()) {
		return nil
	}
	return l.ErrorMap()
}

// TypeCheck checks whether the type is valid for the Rule
func (l Len) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	//nolint:exhaustive // these are the only types that return nil
	switch typ.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan, reflect.String:
		return nil
	}
	return firm.NewRuleTypeError(lenName, typ, "does not have a length (not a Slice, Array, Channel, Map or String)")
}

// ErrorMap returns the ErrorMap returned from ValidateValue
func (l Len) ErrorMap() firm.ErrorMap {
	switch {
	case l.Is != 0:
		return firm.ErrorMap{lenName: firm.TemplateError{
			TemplateFields: map[string]string{"Is": strconv.Itoa(l.Is)},
			Template:       "does not have a length of {{.Is}}",
		}}
	case l.Min != 0 && l.Max != 0:
		return firm.ErrorMap{lenName: firm.TemplateError{
			TemplateFields: map[string]string{"Min": strconv.Itoa(l.Min), "Max": strconv.Itoa(l.Max)},
			Template:       "does not have a length between {{.Min}} and {{.Max}}",
		}}
	case l.Min != 0:
		return firm.ErrorMap{lenName: firm.TemplateError{
			TemplateFields: map[string]string{"Min": strconv.Itoa(l.Min)},
			Template:       "is too short (minimum length is {{.Min}})",
		}}
	default:
		return firm.ErrorMap{lenName: firm.TemplateError{
			TemplateFields: map[string]string{"Max": strconv.Itoa(l.Max)},
			Template:       "is too long (maximum length is {{.Max}})",
		}}
	}
}

func (l Len) valid(length int) bool {
	if l.Is != 0 {
		return length == l.Is
	}
	if l.Min != 0 && length < l.Min {
		return false
	}
	if l.Max != 0 && length > l.Max {
		return false
	}
	return true
}
