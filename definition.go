package firm

import (
	"fmt"
	"reflect"
	"slices"
)

// NewDefinition returns a new Definition. Pointer types are indirected to their value type.
// The type parameter cannot be an interface type, since its zero value has no type
func NewDefinition[T any]() *Definition {
	var zero T
	typ := reflect.TypeOf(zero)
	if typ == nil {
		panic("NewDefinition() called with an interface type--the type of its zero value is nil")
	}
	validator := &Definition{
		typ:       indirectType(typ),
		selfRules: []Rule{},
		ruleMap:   RuleMap{},
	}
	return validator
}

// Definition is a definition of a validation for a type; non-structs support ValidatesSelf rules only
type Definition struct {
	typ       reflect.Type
	selfRules []Rule
	ruleMap   RuleMap

	selfRulesSet bool
	ruleMapSet   bool

	// errOnNilFields flags fields to merge ErrInvalidValue() on, when the field's value is
	// invalid (often from a nil pointer); nil is unset
	errOnNilFields []string

	// errOnNilSelf flags to merge ErrInvalidValue(), when the value itself is invalid
	// (often from a nil pointer) in ValidateAny()/Validate()
	errOnNilSelf bool
}

// ValidatesSelf defines rules of "itself" as a Value
func (s *Definition) ValidatesSelf(rules ...Rule) *Definition {
	if s.selfRulesSet {
		panic(fmt.Sprintf("ValidatesSelf() called twice in type: %v", s.typ.String()))
	}
	s.selfRulesSet = true
	s.selfRules = rules
	return s
}

// Validates defines rules for fields. Panics on a non-struct type. Field checks happen at RegisterType()
func (s *Definition) Validates(ruleMap RuleMap) *Definition {
	if s.typ.Kind() != reflect.Struct {
		panic(fmt.Sprintf("Validates() called on a non-struct type: %v", s.typ.String()))
	}
	if s.ruleMapSet {
		panic(fmt.Sprintf("Validates() called twice in type: %v", s.typ.String()))
	}
	s.ruleMapSet = true
	s.ruleMap = ruleMap
	return s
}

// ErrOnNil flags fields to merge firm.ErrInvalidValue() on, when the field's value is
// invalid (often from a nil pointer), instead of skipping it. Fields must be exported.
// Panics on a non-struct type, no fields given, or called twice. Field checks happen at RegisterType()
func (s *Definition) ErrOnNil(fields ...string) *Definition {
	if s.typ.Kind() != reflect.Struct {
		panic(fmt.Sprintf("ErrOnNil() called on a non-struct type: %v", s.typ.String()))
	}
	if s.errOnNilFields != nil {
		panic(fmt.Sprintf("ErrOnNil() called twice in type: %v", s.typ.String()))
	}
	if len(fields) == 0 {
		panic(fmt.Sprintf("ErrOnNil() called with no fields in type: %v", s.typ.String()))
	}
	s.errOnNilFields = fields
	return s
}

// ErrOnNilSelf flags to merge firm.ErrInvalidValue(), when the value itself is nil
// (often from a nil pointer) in firm.Registry.ValidateAny()/Validate(), instead of skipping it.
// Panics when called twice
func (s *Definition) ErrOnNilSelf() *Definition {
	if s.errOnNilSelf {
		panic(fmt.Sprintf("ErrOnNilSelf() called twice in type: %v", s.typ.String()))
	}
	s.errOnNilSelf = true
	return s
}

// Type returns the type for the definition
func (s *Definition) Type() reflect.Type { return s.typ }

// SelfRules returns a copy of the rules that apply to "itself" as a Value
func (s *Definition) SelfRules() []Rule { return slices.Clone(s.selfRules) }

// RuleMap returns a copy of the rules mapped to each field
func (s *Definition) RuleMap() RuleMap { return copyRuleMap(s.ruleMap) }
