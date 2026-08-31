// Package firm contains functions to do validations
package firm

import (
	"reflect"
)

var (
	// MustRegisterType registers the TypeDefinition to the DefaultRegistry, panics if there is an error
	MustRegisterType = DefaultRegistry.MustRegisterType
	// RegisterType registers the TypeDefinition to the DefaultRegistry
	RegisterType = DefaultRegistry.RegisterType
	// ValidateAny validates the data with the DefaultRegistry
	ValidateAny = DefaultRegistry.ValidateAny
	// Backed returns a RegistryBacker for the DefaultRegistry
	Backed = DefaultRegistry.Backed

	// DefaultRegistry is the registry used for global functions.
	// Like any Registry, register all types upfront (e.g. in init()) and do not register while validating.
	DefaultRegistry = &Registry{}
)

// DefaultValidator is the validator used by registries for not found types when Registry.DefaultValidator is not defined
var DefaultValidator = RuleVldr{Rule: NotFoundRule{}}

// NotFoundRule is the rule used for not found types in the DefaultValidator
type NotFoundRule struct{}

// ValidateValue validates the value, naming the value's type as not found in the Registry
func (n NotFoundRule) ValidateValue(value reflect.Value) ErrorMap {
	// invalid values and nilType stand-ins have no type to name
	if !value.IsValid() || value.Type() == nilValueType {
		return n.ErrorMap()
	}
	return ErrorMap{"NotFound": TemplateError{
		Template:       "type, {{.ValueTypeName}}, not found in Registry",
		TemplateFields: map[string]string{"ValueTypeName": value.Type().String()},
	}}
}

// TypeCheck checks whether the type is valid for the Rule
func (n NotFoundRule) TypeCheck(_ reflect.Type) *RuleTypeError { return nil }

// ErrorMap returns the ErrorMap returned from ValidateValue, when the value is invalid
// and has no type to name
func (n NotFoundRule) ErrorMap() ErrorMap {
	return ErrorMap{"NotFound": TemplateError{Template: "type, {{.RootTypeName}}, not found in Registry"}}
}

// RuleMap is a map of struct fields to rules
type RuleMap map[string][]Rule

// Rule defines a rule for validation definitions and validators
type Rule interface {
	ValidateValue(value reflect.Value) ErrorMap
	TypeCheck(typ reflect.Type) *RuleTypeError
}

// RuleBasic is a Rule that is not composed of other rules
type RuleBasic interface {
	Rule
	ErrorMap() ErrorMap
}

// RuleTyped is a RuleBasic that's typed
type RuleTyped[T any] interface {
	RuleBasic
	Validate(data T) ErrorMap
}

// Validator validates the data
type Validator interface {
	Rule
	ValidateAny(data any) ErrorMap
	ValidateMerge(value reflect.Value, key string, errorMap ErrorMap)
}

// ValidatorTyped is a generic firm.Validator that has a typed Validate() function
type ValidatorTyped[T any] interface {
	Validator
	Validate(data T) ErrorMap
}
