package rule

import (
	"reflect"

	"github.com/s12chung/firm"
)

// Named renames the error keys of .Rule to .Name
type Named struct {
	Name string
	Rule firm.RuleBasic
}

// ValidateValue validates the data value with .Rule, renaming the error keys (assumes TypeCheck is called)
func (n Named) ValidateValue(value reflect.Value) firm.ErrorMap {
	return n.renamed(n.Rule.ValidateValue(value))
}

// TypeCheck checks whether the type is valid for the Rule
func (n Named) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	ruleTypeError := n.Rule.TypeCheck(typ)
	if ruleTypeError != nil {
		ruleTypeError.RuleName = n.Name
	}
	return ruleTypeError
}

// ErrorMap returns the ErrorMap returned from ValidateValue
func (n Named) ErrorMap() firm.ErrorMap {
	return n.renamed(n.Rule.ErrorMap())
}

// renamed renames the error keys
func (n Named) renamed(errorMap firm.ErrorMap) firm.ErrorMap {
	if len(errorMap) == 0 {
		return nil
	}
	renamed := firm.ErrorMap{}
	for k, err := range errorMap {
		name := n.Name
		if len(errorMap) > 1 {
			name += "-" + string(k)
		}
		renamed[firm.ErrorKey(name)] = err
	}
	return renamed
}
