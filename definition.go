package firm

import (
	"fmt"
	"reflect"
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

// Validates defines rules for fields
// Field checks (struct type, field exists and exported) happen at RegisterType() via FieldsAnyWithErr()
func (s *Definition) Validates(ruleMap RuleMap) *Definition {
	if s.ruleMapSet {
		panic(fmt.Sprintf("Validates() called twice in type: %v", s.typ.String()))
	}
	s.ruleMapSet = true
	s.ruleMap = ruleMap
	return s
}

// Type returns the type for the definition
func (s *Definition) Type() reflect.Type { return s.typ }

// SelfRules return the rules that apply to "itself" as a Value
func (s *Definition) SelfRules() []Rule { return s.selfRules }

// RuleMap returns the map of rules for the structure
func (s *Definition) RuleMap() RuleMap { return s.ruleMap }
