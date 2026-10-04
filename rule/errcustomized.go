package rule

import (
	"reflect"

	"github.com/s12chung/firm"
)

// ErrCustomized customizes the ErrorMap of .Rule through .CustomErr,
// which receives a mutable Clone() of .Rule's ErrorMap
type ErrCustomized struct {
	Rule      firm.RuleBasic
	CustomErr func(firm.ErrorMap) firm.ErrorMap
}

// ValidateValue validates the data value with .Rule, customizing the ErrorMap (assumes TypeCheck is called)
func (e ErrCustomized) ValidateValue(value reflect.Value) firm.ErrorMap {
	return e.customized(e.Rule.ValidateValue(value))
}

// TypeCheck checks whether the type is valid for the Rule
func (e ErrCustomized) TypeCheck(typ reflect.Type) *firm.RuleTypeError { return e.Rule.TypeCheck(typ) }

// ErrorMap returns the ErrorMap returned from ValidateValue
func (e ErrCustomized) ErrorMap() firm.ErrorMap { return e.customized(e.Rule.ErrorMap()) }

// customized returns .CustomErr applied to a clone of errorMap--nil when .Rule passes (empty errorMap)
func (e ErrCustomized) customized(errorMap firm.ErrorMap) firm.ErrorMap {
	if len(errorMap) == 0 {
		return nil
	}
	return e.CustomErr(errorMap.Clone()).ToNil()
}
