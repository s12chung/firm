package rule

import (
	"reflect"
	"strings"

	"github.com/s12chung/firm"
)

const orName = "Or"

// Or returns a Rule that checks if data passes any one of rules, panics without rules
func Or(rules ...firm.RuleBasic) OrWrap { return OrWrap{rulesOrPanic(rules, "Or")} }

// OrWrap checks if data passes any one of its Rules
type OrWrap struct{ rules []firm.RuleBasic }

// ValidateValue passes if any one of its Rules pass (assumes TypeCheck is called)
func (o OrWrap) ValidateValue(value reflect.Value) firm.ErrorMap {
	for _, rule := range o.rules {
		if rule.ValidateValue(value).ToNil() == nil {
			return nil
		}
	}
	return o.ErrorMap()
}

// TypeCheck checks whether the type is valid for every Rule
func (o OrWrap) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	for _, rule := range o.rules {
		if err := rule.TypeCheck(typ); err != nil {
			return err
		}
	}
	return nil
}

// ErrorMap returns the ErrorMap returned from ValidateValue
func (o OrWrap) ErrorMap() firm.ErrorMap {
	return firm.ErrorMap{orName: firm.TemplateError{
		TemplateFields: map[string]string{"Errors": rulesStr(o.rules)},
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
