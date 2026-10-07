package rule

import (
	"reflect"

	"github.com/s12chung/firm"
)

// And returns a Rule that checks if data passes all of rules--only for easier composition,
// as []firm.Rule is passed throughout in firm.Validator and firm.Definition as an AND. Panics without rules
func And(rules ...firm.RuleBasic) AndWrap { return AndWrap{rulesOrPanic(rules, "And")} }

// AndWrap checks if data passes all of its Rules--only for easier composition,
// as []firm.Rule is passed throughout in firm.Validator and firm.Definition as an AND
type AndWrap struct{ rules []firm.RuleBasic }

// ValidateValue merges the ErrorMaps of every failing Rule (assumes TypeCheck is called)
func (a AndWrap) ValidateValue(value reflect.Value) firm.ErrorMap {
	return a.mergeErrorMaps(func(rule firm.RuleBasic) firm.ErrorMap { return rule.ValidateValue(value) })
}

// TypeCheck checks whether the type is valid for every Rule
func (a AndWrap) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	for _, rule := range a.rules {
		if err := rule.TypeCheck(typ); err != nil {
			return err
		}
	}
	return nil
}

// ErrorMap returns the ErrorMap returned from ValidateValue
func (a AndWrap) ErrorMap() firm.ErrorMap { return a.mergeErrorMaps(firm.RuleBasic.ErrorMap) }

func (a AndWrap) mergeErrorMaps(errMapFor func(firm.RuleBasic) firm.ErrorMap) firm.ErrorMap {
	errorMap := firm.ErrorMap{}
	for _, rule := range a.rules {
		errorMap.Merge("", errMapFor(rule))
	}
	return errorMap.ToNil()
}
