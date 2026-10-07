package rule

import (
	"reflect"

	"github.com/s12chung/firm"
)

// ShowValue returns a Rule that appends the failing value to rule's error messages,
// the CLI standard: `Str is not one of Noun, Verb: "Noun"`
func ShowValue(rule firm.RuleBasic) ShowValueWrap { return ShowValueWrap{rule} }

// ShowValueWrap appends the failing value to its Rule's error messages,
// the CLI standard: `Str is not one of Noun, Verb: "Noun"`
type ShowValueWrap struct{ rule firm.RuleBasic }

// ValidateValue validates the data value with its Rule, appending the failing value to the errors (assumes TypeCheck is called)
func (s ShowValueWrap) ValidateValue(value reflect.Value) firm.ErrorMap {
	return s.valueToErrTemplate(s.rule.ValidateValue(value))
}

// TypeCheck checks whether the type is valid for the Rule
func (s ShowValueWrap) TypeCheck(typ reflect.Type) *firm.RuleTypeError { return s.rule.TypeCheck(typ) }

// ErrorMap returns the ErrorMap returned from ValidateValue
func (s ShowValueWrap) ErrorMap() firm.ErrorMap { return s.valueToErrTemplate(s.rule.ErrorMap()) }

func (s ShowValueWrap) valueToErrTemplate(errorMap firm.ErrorMap) firm.ErrorMap {
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
