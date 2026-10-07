package rule

import (
	"reflect"

	"github.com/s12chung/firm"
)

// CustomizeErr returns a Rule that customizes the ErrorMap of rule through customErr,
// which receives a mutable Clone() of rule's ErrorMap
func CustomizeErr(rule firm.RuleBasic, customErr func(firm.ErrorMap) firm.ErrorMap) CustomizeErrWrap {
	return CustomizeErrWrap{rule, customErr}
}

// CustomizeErrWrap customizes the ErrorMap of its Rule through customErr,
// which receives a mutable Clone() of its Rule's ErrorMap
type CustomizeErrWrap struct {
	rule      firm.RuleBasic
	customErr func(firm.ErrorMap) firm.ErrorMap
}

// ValidateValue validates the data value with its Rule, customizing the ErrorMap (assumes TypeCheck is called)
func (c CustomizeErrWrap) ValidateValue(value reflect.Value) firm.ErrorMap {
	return c.customized(c.rule.ValidateValue(value))
}

// TypeCheck checks whether the type is valid for the Rule
func (c CustomizeErrWrap) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	return c.rule.TypeCheck(typ)
}

// ErrorMap returns the ErrorMap returned from ValidateValue
func (c CustomizeErrWrap) ErrorMap() firm.ErrorMap { return c.customized(c.rule.ErrorMap()) }

// customized returns customErr applied to a clone of errorMap--nil when its Rule passes (empty errorMap)
func (c CustomizeErrWrap) customized(errorMap firm.ErrorMap) firm.ErrorMap {
	if len(errorMap) == 0 {
		return nil
	}
	return c.customErr(errorMap.Clone()).ToNil()
}
