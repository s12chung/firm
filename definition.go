package firm

import (
	"fmt"
	"reflect"
)

// NewDefinition returns a new Definition. Pointer types are indirected to their value type
func NewDefinition[T any]() *Definition {
	var zero T
	typ := indirectType(reflect.TypeOf(zero))
	validator := &Definition{
		typ:       typ,
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
}

// ValidatesSelf defines rules of "itself" as a Value
func (s *Definition) ValidatesSelf(rules ...Rule) *Definition {
	if len(s.selfRules) != 0 {
		panic(fmt.Sprintf("ValidatesSelf() called twice in type: %v", s.typ.String()))
	}
	s.selfRules = rules
	return s
}

// Validates defines rules for fields
func (s *Definition) Validates(ruleMap RuleMap) *Definition {
	if len(s.ruleMap) != 0 {
		panic(fmt.Sprintf("Validates() called twice in type: %v", s.typ.String()))
	}
	for fieldName := range ruleMap {
		_, exists := s.typ.FieldByName(fieldName)
		if !exists {
			panic(fmt.Sprintf("Validates() called with fieldName, %v, not in type: %v", fieldName, s.typ.String()))
		}
	}
	s.ruleMap = ruleMap
	return s
}

// Type returns the type for the definition
func (s *Definition) Type() reflect.Type { return s.typ }

// SelfRules return the rules that apply to "itself" as a Value
func (s *Definition) SelfRules() []Rule { return s.selfRules }

// RuleMap returns the map of rules for the structure
func (s *Definition) RuleMap() RuleMap { return s.ruleMap }
