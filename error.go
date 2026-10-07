package firm

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"unicode/utf8"
)

const (
	valueNamePrefix = "{{.ValueName}} " // default Error prefix--the error's subject (ex. "Str is not present")
	valueCap        = 50                // default {{value}} render cap
)

// ErrorMap is a map of TemplateError keys to their respective TemplateError
//
//nolint:errname
type ErrorMap map[ErrorKey]TemplateError

// Error returns the error string for the ErrorMap
func (e ErrorMap) Error() string { return e.ErrorWith(valueNamePrefix, "") }

// ErrorWith returns the error string for the ErrorMap, with prefix and suffix passed to each TemplateError.ErrorWith
func (e ErrorMap) ErrorWith(prefix, suffix string) string {
	errs := make([]string, len(e))
	for i, k := range e.sortedKeys() {
		errs[i] = string(k) + ": " + e[k].ErrorWith(prefix, suffix)
	}
	return strings.Join(errs, ", ")
}

func (e ErrorMap) sortedKeys() []ErrorKey {
	keys := make([]ErrorKey, len(e))
	i := 0
	for k := range e {
		keys[i] = k
		i++
	}
	slices.Sort(keys)
	return keys
}

// Merge merges src into e, given appending path to the src keys.
func (e ErrorMap) Merge(path string, src ErrorMap) {
	for k, v := range src {
		key := joinKeys(ErrorKey(path), k)
		v.ErrorKey = key // indicate path change
		e[key] = v
	}
}

// Clone returns a deep copy of e--new ErrorMap and TemplateFields, so mutating the copy does not affect e
func (e ErrorMap) Clone() ErrorMap {
	clone := make(ErrorMap, len(e))
	for key, templateError := range e {
		if templateError.TemplateFields != nil {
			templateError.TemplateFields = maps.Clone(templateError.TemplateFields)
		}
		clone[key] = templateError
	}
	return clone
}

// withValue returns a copy of e with value captured on each TemplateError--never
// mutating e, as Rules can return shared ErrorMaps
func (e ErrorMap) withValue(value reflect.Value) ErrorMap {
	withValue := make(ErrorMap, len(e))
	for key, templateError := range e {
		// validators nest and the innermost capture is the failing value
		if !templateError.value.IsValid() {
			templateError.value = value
		}
		withValue[key] = templateError
	}
	return withValue
}

// ToNil returns itself or nil if it's empty
func (e ErrorMap) ToNil() ErrorMap {
	if len(e) == 0 {
		return nil
	}
	return e
}

// TemplateError is an error that contains a key matching a field or "itself" as a Value, a golang template, and template fields
type TemplateError struct {
	Template       string
	TemplateFields map[string]string
	ErrorKey       ErrorKey
	// captured at validation time, rendered by the {{value}} template func so it's never serialized
	value reflect.Value
}

// Error returns a string for the error, prefixed with its ValueName
func (t TemplateError) Error() string { return t.ErrorWith(valueNamePrefix, "") }

// ErrorWith returns a string for the error, with prefix and suffix,
// each parsed as a template. {{value}} renders the captured failing value (ex. `{{value 20}}`)
func (t TemplateError) ErrorWith(prefix, suffix string) string {
	badTemplateString := t.Template + " (bad format)"
	temp, err := template.New("top").Funcs(template.FuncMap{
		"value": t.valueString,
	}).Parse(prefix + t.Template + suffix)
	if err != nil {
		return badTemplateString
	}

	templateDot := map[string]string{}
	if t.TemplateFields != nil {
		templateDot = maps.Clone(t.TemplateFields)
	}
	typeName := t.ErrorKey.RootTypeName()
	if typeName == "" {
		typeName = "NoType"
	}
	templateDot["RootTypeName"] = typeName
	valueName := t.ErrorKey.ValueName()
	if valueName == "" {
		valueName = "value"
	}
	templateDot["ValueName"] = valueName

	var sb strings.Builder
	if err = temp.Execute(&sb, templateDot); err != nil {
		return badTemplateString
	}
	return sb.String()
}

// valueString renders the captured value: quoted capped strings (ex. `"Nou...`), numbers/bools
// as-is, composites as type names--nil / never-captured values render `<nil>` / `<no value>`
func (t TemplateError) valueString(limits ...int) string {
	// limit is variadic, not a plain arg, because text/template dispatches on exact arity--
	// the optional cap lets one FuncMap name serve both `{{value}}` and `{{value 20}}` (like printf)
	limit := valueCap
	if len(limits) > 0 {
		limit = limits[0]
	}

	value := t.value
	if !value.IsValid() {
		return "<no value>"
	}
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return "<nil>"
		}
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.String:
		return quoteCapped(value.String(), limit)
	case reflect.Slice, reflect.Array, reflect.Map:
		// composite fields/elements produce their own errors, so name the type only
		return fmt.Sprintf("%v (len %d)", value.Type(), value.Len())
	case reflect.Struct, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return value.Type().String()
	default:
		return fmt.Sprintf("%v", value.Interface())
	}
}

// quoteCapped returns s strconv.Quote'd, truncated to limit bytes on a rune boundary
func quoteCapped(s string, limit int) string {
	truncated := false
	if len(s) > limit {
		for limit > 0 && !utf8.RuneStart(s[limit]) {
			limit--
		}
		s, truncated = s[:limit], true
	}
	quoted := strconv.Quote(s)
	if truncated {
		quoted += "..."
	}
	return quoted
}

// ErrorKey is a string that has helper functions relating to error keys
type ErrorKey string

// split returns the key's segments, split on keySeparator outside of brackets
func (e ErrorKey) split() []string {
	s := string(e)

	var segments []string
	depth, start := 0, 0
	for i, r := range s {
		switch r {
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		case '.':
			if depth == 0 {
				segments = append(segments, s[start:i])
				start = i + 1
			}
		}
	}
	return append(segments, s[start:])
}

const typeSegmentCount = 2

// RootTypeName returns the type name of the key
func (e ErrorKey) RootTypeName() string {
	segments := e.split()
	if len(segments) <= typeSegmentCount {
		return ""
	}
	return strings.Join(segments[:2], keySeparator)
}

// ValueName returns the value name of the key - the Struct field, array index or value type name
func (e ErrorKey) ValueName() string {
	segments := e.split()
	switch {
	case len(segments) <= typeSegmentCount:
		return ""
	case len(segments) == typeSegmentCount+1:
		return strings.Join(segments[:typeSegmentCount], keySeparator)
	}
	// fixes: "[0] does not match" --> "Domains[0]"
	name := segments[len(segments)-typeSegmentCount]
	// not a [0]
	if !strings.HasPrefix(name, "[") {
		return name
	}
	start := len(segments) - typeSegmentCount
	for start > typeSegmentCount && strings.HasPrefix(segments[start-1], "[") {
		start--
	}
	if start == typeSegmentCount {
		return name
	}
	return segments[start-1] + strings.Join(segments[start:len(segments)-1], "")
}

// ErrorName returns the error name of the key
func (e ErrorKey) ErrorName() string {
	segments := e.split()
	if len(segments) < 2 {
		return ""
	}
	return segments[len(segments)-1]
}

// NewRuleTypeError returns a new RuleTypeError
func NewRuleTypeError(ruleName string, typ reflect.Type, badCondition string) *RuleTypeError {
	return &RuleTypeError{RuleName: ruleName, Type: typ, BadCondition: badCondition}
}

// RuleTypeError is an error returned by Rule.TypeCheck
type RuleTypeError struct {
	RuleName     string
	Type         reflect.Type
	BadCondition string
}

// TemplateError returns the TemplateError represented by the RuleTypeError
func (r RuleTypeError) TemplateError() TemplateError {
	valueTypeName := "nil"
	if r.Type != nil {
		valueTypeName = r.Type.String()
	}
	return TemplateError{
		TemplateFields: map[string]string{"ValueTypeName": valueTypeName},
		Template:       r.BadCondition + ", got {{.ValueTypeName}}",
	}
}

// Error returns the error string for the error
func (r RuleTypeError) Error() string { return r.RuleName + ": " + r.TemplateError().Error() }

// TypeCheck is a basic implementation for TypeCheck
func TypeCheck(ruleName string, typ, expectedType reflect.Type, kindString string) *RuleTypeError {
	// Validator types are stored indirected, so indirect the incoming type--data may be a pointer
	typ = indirectType(typ)
	if typ == expectedType {
		return nil
	}
	if kindString != "" {
		kindString = " " + kindString + " of"
	}
	return NewRuleTypeError(ruleName, typ, "is not matching"+kindString+" type "+expectedType.String())
}
