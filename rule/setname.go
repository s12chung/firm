package rule

import (
	"reflect"

	"github.com/s12chung/firm"
)

// SetName returns a Rule that renames the error keys of rule to name
func SetName(name string, rule firm.RuleBasic) SetNameWrap { return SetNameWrap{name, rule} }

// SetNameWrap renames the error keys of its Rule to name
type SetNameWrap struct {
	name string
	rule firm.RuleBasic
}

// ValidateValue validates the data value with its Rule, renaming the error keys (assumes TypeCheck is called)
func (s SetNameWrap) ValidateValue(value reflect.Value) firm.ErrorMap {
	return s.renamed(s.rule.ValidateValue(value))
}

// TypeCheck checks whether the type is valid for the Rule
func (s SetNameWrap) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	ruleTypeError := s.rule.TypeCheck(typ)
	if ruleTypeError != nil {
		ruleTypeError.RuleName = s.name
	}
	return ruleTypeError
}

// ErrorMap returns the ErrorMap returned from ValidateValue
func (s SetNameWrap) ErrorMap() firm.ErrorMap {
	return s.renamed(s.rule.ErrorMap())
}

// renamed renames the error keys
func (s SetNameWrap) renamed(errorMap firm.ErrorMap) firm.ErrorMap {
	if len(errorMap) == 0 {
		return nil
	}
	renamed := firm.ErrorMap{}
	for k, err := range errorMap {
		name := s.name
		if len(errorMap) > 1 {
			name += "-" + string(k)
		}
		renamed[firm.ErrorKey(name)] = err
	}
	return renamed
}
