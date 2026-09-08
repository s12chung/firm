package rule

import (
	"fmt"
	"reflect"

	"github.com/s12chung/firm"
)

const oneNotNilName = "OneNotNil"

// OneNotNil requires exactly one of the named Fields (that are pointer types) to be non-nil
type OneNotNil struct{ Fields []string }

// ValidateValue returns an error, unless exactly one named field is non-nil (assumes TypeCheck is called)
func (o OneNotNil) ValidateValue(value reflect.Value) firm.ErrorMap {
	notNilFields := make([]string, 0, len(o.Fields))
	for _, f := range o.Fields {
		if !value.FieldByName(f).IsNil() {
			notNilFields = append(notNilFields, f)
		}
	}
	if len(notNilFields) == 1 {
		return nil
	}
	return firm.ErrorMap{oneNotNilName: firm.TemplateError{
		Template:       "must have exactly one of {{.Fields}} non-nil, got {{.NotNil}}",
		TemplateFields: map[string]string{"Fields": fmt.Sprintf("%v", o.Fields), "NotNil": fmt.Sprintf("%v", notNilFields)},
	}}
}

// TypeCheck requires a struct carrying all named Fields as unique struct fields that are pointers
func (o OneNotNil) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	if typ.Kind() != reflect.Struct {
		return firm.NewRuleTypeError(oneNotNilName, typ, "is not a Struct")
	}
	if len(o.Fields) == 0 {
		return firm.NewRuleTypeError(oneNotNilName, typ, "has no Fields")
	}
	seen := map[string]bool{}
	for _, f := range o.Fields {
		if seen[f] {
			return firm.NewRuleTypeError(oneNotNilName, typ, "has duplicate Fields, "+f)
		}
		seen[f] = true
		field, ok := typ.FieldByName(f)
		if !ok {
			return firm.NewRuleTypeError(oneNotNilName, typ, "has no field, "+f)
		}
		if field.Type.Kind() != reflect.Pointer {
			return firm.NewRuleTypeError(oneNotNilName, typ, "has a non-Pointer field, "+f)
		}
	}
	return nil
}

// ErrorMap returns the ErrorMap returned from ValidateValue, without the non-nil Fields named
func (o OneNotNil) ErrorMap() firm.ErrorMap {
	return firm.ErrorMap{oneNotNilName: firm.TemplateError{
		Template:       "must have exactly one of {{.Fields}} non-nil",
		TemplateFields: map[string]string{"Fields": fmt.Sprintf("%v", o.Fields)},
	}}
}
