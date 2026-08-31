package rule

import (
	"reflect"

	"github.com/s12chung/firm"
)

// Not is a rule that negates a firm.RuleBasic
type Not struct{ Rule firm.RuleBasic }

// ValidateValue negates the Rule's ValidateValue() (assumes TypeCheck is called)
func (n Not) ValidateValue(value reflect.Value) firm.ErrorMap {
	if n.Rule.ValidateValue(value).ToNil() == nil {
		return n.ErrorMap()
	}
	return nil
}

// TypeCheck checks whether the type is valid for the Rule
func (n Not) TypeCheck(typ reflect.Type) *firm.RuleTypeError { return n.Rule.TypeCheck(typ) }

// ErrorMap returns the ErrorMap returned from ValidateValue
func (n Not) ErrorMap() firm.ErrorMap {
	original := n.Rule.ErrorMap()
	if len(original) == 0 {
		return nil
	}

	errorMap := firm.ErrorMap{}
	for k, err := range original {
		err.Template += "--Not"
		errorMap["Not"+k] = err
	}
	return errorMap
}
