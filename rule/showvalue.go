package rule

import (
	"reflect"

	"github.com/s12chung/firm"
)

// ShowValue appends the failing value to .Rule's error messages,
// the CLI standard: `Str is not one of Noun, Verb: "Noun"`
type ShowValue struct{ Rule firm.RuleBasic }

// ValidateValue validates the data value with .Rule, appending the failing value to the errors (assumes TypeCheck is called)
func (s ShowValue) ValidateValue(value reflect.Value) firm.ErrorMap {
	return s.valueToErrTemplate(s.Rule.ValidateValue(value))
}

// TypeCheck checks whether the type is valid for the Rule
func (s ShowValue) TypeCheck(typ reflect.Type) *firm.RuleTypeError { return s.Rule.TypeCheck(typ) }

// ErrorMap returns the ErrorMap returned from ValidateValue
func (s ShowValue) ErrorMap() firm.ErrorMap { return s.valueToErrTemplate(s.Rule.ErrorMap()) }

func (s ShowValue) valueToErrTemplate(errorMap firm.ErrorMap) firm.ErrorMap {
	if len(errorMap) == 0 {
		return nil
	}
	shown := firm.ErrorMap{}
	for key, err := range errorMap {
		err.Template += ": {{value}}"
		shown[key] = err
	}
	return shown
}
