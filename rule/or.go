package rule

import (
	"reflect"
	"strings"

	"github.com/s12chung/firm"
)

const orName = "Or"

// Or checks if data passes any one of .Rules
type Or struct{ Rules []firm.RuleBasic }

// ValidateValue passes if any one of the Rules pass (assumes TypeCheck is called)
func (o Or) ValidateValue(value reflect.Value) firm.ErrorMap {
	for _, rule := range o.Rules {
		if rule.ValidateValue(value).ToNil() == nil {
			return nil
		}
	}
	return o.ErrorMap()
}

// TypeCheck checks whether the type is valid for every Rule
func (o Or) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	for _, rule := range o.Rules {
		if err := rule.TypeCheck(typ); err != nil {
			return err
		}
	}
	return nil
}

// ErrorMap returns the ErrorMap returned from ValidateValue
func (o Or) ErrorMap() firm.ErrorMap {
	return firm.ErrorMap{orName: firm.TemplateError{
		TemplateFields: map[string]string{"Errors": rulesStr(o.Rules)},
		Template:       "is not any of {{.Errors}}",
	}}
}

func rulesStr(rules []firm.RuleBasic) string {
	strs := make([]string, len(rules))
	for i, rule := range rules {
		strs[i] = rule.ErrorMap().Error()
	}
	return "[" + strings.Join(strs, "; ") + "]"
}
