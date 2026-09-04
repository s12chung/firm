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

	// notNilFields flags fields to merge ErrNilPointer() on, when the field's value is
	// a nil pointer; nil is unset
	notNilFields []string

	// notNilSelf flags to merge ErrNilPointer(), when the value itself is a nil pointer in ValidateAny()/Validate()
	notNilSelf bool
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

// NotNil flags fields to merge firm.ErrNilPointer() on,
// when the field's value is a nil pointer, instead of skipping it. Fields must be exported.
// Panics on a non-struct type, no fields given, or called twice. Field checks happen at RegisterType()
func (s *Definition) NotNil(fields ...string) *Definition {
	if s.typ.Kind() != reflect.Struct {
		panic(fmt.Sprintf("NotNil() called on a non-struct type: %v", s.typ.String()))
	}
	if s.notNilFields != nil {
		panic(fmt.Sprintf("NotNil() called twice in type: %v", s.typ.String()))
	}
	if len(fields) == 0 {
		panic(fmt.Sprintf("NotNil() called with no fields in type: %v", s.typ.String()))
	}
	s.notNilFields = fields
	return s
}

// NotNilSelf flags to merge firm.ErrNilPointer(),
// when the value itself is a nil pointer in firm.Registry.ValidateAny()/Validate(), instead of skipping it.
// Panics when called twice
func (s *Definition) NotNilSelf() *Definition {
	if s.notNilSelf {
		panic(fmt.Sprintf("NotNilSelf() called twice in type: %v", s.typ.String()))
	}
	s.notNilSelf = true
	return s
}

// Type returns the type for the definition
func (s *Definition) Type() reflect.Type { return s.typ }

// SelfRules returns a copy of the rules that apply to "itself" as a Value
func (s *Definition) SelfRules() []Rule { return slices.Clone(s.selfRules) }

// RuleMap returns a copy of the rules mapped to each field
func (s *Definition) RuleMap() RuleMap { return copyRuleMap(s.ruleMap) }
