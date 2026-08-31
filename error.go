package firm

import (
	"maps"
	"reflect"
	"slices"
	"strings"
	"text/template"
)

// ErrorMap is a map of TemplateError keys to their respective TemplateError
//
//nolint:errname
type ErrorMap map[ErrorKey]TemplateError

func (e ErrorMap) Error() string {
	errors := make([]string, len(e))
	for i, k := range e.sortedKeys() {
		errors[i] = string(k) + ": " + e[k].Error()
	}
	return strings.Join(errors, ", ")
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

// MergeInto merges into dest, given appending path to the src keys.
func (e ErrorMap) MergeInto(path string, dest ErrorMap) {
	for k, v := range e {
		key := joinKeys(ErrorKey(path), k)
		v.ErrorKey = key // indicate path change
		dest[key] = v
	}
}

// ToNil returns itself or nil if it's empty
func (e ErrorMap) ToNil() ErrorMap {
	if len(e) == 0 {
		return nil
	}
	return e
}

// Finish finishes the ErrorMap for consumption by filling in the typeName and ValueName
func (e ErrorMap) Finish() ErrorMap {
	for k, v := range e {
		v.ErrorKey = k
		e[k] = v
	}
	return e.ToNil()
}

// TemplateError is an error that contains a key matching a field or "itself" as a Value, a golang template, and template fields
type TemplateError struct {
	Template       string
	TemplateFields map[string]string
	ErrorKey       ErrorKey
}

// Error returns a string for the error
func (t TemplateError) Error() string {
	badTemplateString := t.Template + " (bad format)"
	temp, err := template.New("top").Parse("{{.ValueName}} " + t.Template)
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

// RootTypeName returns the type name of the key
func (e ErrorKey) RootTypeName() string {
	segments := e.split()
	if len(segments) < 3 {
		return ""
	}
	return strings.Join(segments[:2], keySeparator)
}

// ValueName returns the value name of the key - the Struct field, array index or value type name
func (e ErrorKey) ValueName() string {
	segments := e.split()
	switch {
	case len(segments) < 3:
		return ""
	case len(segments) == 3:
		return strings.Join(segments[:2], keySeparator)
	}
	return segments[len(segments)-2]
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
