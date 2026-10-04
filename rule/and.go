package rule

import (
	"reflect"

	"github.com/s12chung/firm"
)

// And checks if data passes all of .Rules--only for easier composition,
// as []firm.Rule is passed throughout in firm.Validator and firm.Definition as an AND
type And struct{ Rules []firm.RuleBasic }

// ValidateValue merges the ErrorMaps of every failing Rule (assumes TypeCheck is called)
func (a And) ValidateValue(value reflect.Value) firm.ErrorMap {
	return a.mergeErrorMaps(func(rule firm.RuleBasic) firm.ErrorMap { return rule.ValidateValue(value) })
}

// TypeCheck checks whether the type is valid for every Rule
func (a And) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	for _, rule := range a.Rules {
		if err := rule.TypeCheck(typ); err != nil {
			return err
		}
	}
	return nil
}

// ErrorMap returns the ErrorMap returned from ValidateValue
func (a And) ErrorMap() firm.ErrorMap { return a.mergeErrorMaps(firm.RuleBasic.ErrorMap) }

func (a And) mergeErrorMaps(errMapFor func(firm.RuleBasic) firm.ErrorMap) firm.ErrorMap {
	errorMap := firm.ErrorMap{}
	for _, rule := range a.Rules {
		errorMap.Merge("", errMapFor(rule))
	}
	return errorMap.ToNil()
}
