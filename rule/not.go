package rule

import (
	"reflect"

	"github.com/s12chung/firm"
)

// Not returns a Rule that negates rule
func Not(rule firm.RuleBasic) NotWrap { return NotWrap{rule} }

// NotWrap negates its Rule
type NotWrap struct{ rule firm.RuleBasic }

// ValidateValue negates its Rule's ValidateValue() (assumes TypeCheck is called)
func (n NotWrap) ValidateValue(value reflect.Value) firm.ErrorMap {
	if n.rule.ValidateValue(value).ToNil() == nil {
		return n.ErrorMap()
	}
	return nil
}

// TypeCheck checks whether the type is valid for the Rule
func (n NotWrap) TypeCheck(typ reflect.Type) *firm.RuleTypeError { return n.rule.TypeCheck(typ) }

// ErrorMap returns the ErrorMap returned from ValidateValue
func (n NotWrap) ErrorMap() firm.ErrorMap {
	original := n.rule.ErrorMap()
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
