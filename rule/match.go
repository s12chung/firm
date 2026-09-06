package rule

import (
	"reflect"
	"regexp"

	"github.com/s12chung/firm"
)

const matchName = "Match"

// Match checks if data matches .Regexp
type Match struct{ Regexp *regexp.Regexp }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (m Match) ValidateValue(value reflect.Value) firm.ErrorMap {
	return m.Validate(value.String())
}

// Validate validates the data value
func (m Match) Validate(data string) firm.ErrorMap {
	if m.Regexp.MatchString(data) {
		return nil
	}
	return m.ErrorMap()
}

// TypeCheck checks whether the type is valid for the Rule
func (m Match) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	if typ.Kind() != reflect.String {
		return firm.NewRuleTypeError(matchName, typ, "is not a String")
	}
	return nil
}

// ErrorMap returns the ErrorMap returned from ValidateValue
func (m Match) ErrorMap() firm.ErrorMap {
	return firm.ErrorMap{matchName: firm.TemplateError{
		TemplateFields: map[string]string{"Regexp": m.Regexp.String()},
		Template:       "does not match {{.Regexp}}",
	}}
}
