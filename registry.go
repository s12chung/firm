package firm

import (
	"fmt"
	"reflect"
)

// Registry registers types to find the right validator to validate with.
//
// RegisterType must not be called while validating; register all types upfront,
// e.g. in init(). Once registration completes, ValidateAny/ValidateValue/TypeCheck
// are safe for concurrent use.
type Registry struct {
	typeToValidator  map[reflect.Type]*ValueAnyVldr
	DefaultValidator Validator
}

// MustRegisterType registers the Definition to validate the type, panics if there is an error
func (r *Registry) MustRegisterType(definition *Definition) {
	if err := r.RegisterType(definition); err != nil {
		panic(err.Error())
	}
}

// RegisterType registers the Definition to validate the type
func (r *Registry) RegisterType(definition *Definition) error {
	if r.typeToValidator == nil {
		r.typeToValidator = map[reflect.Type]*ValueAnyVldr{}
	}

	typ := definition.typ
	if _, exists := r.typeToValidator[typ]; exists {
		return fmt.Errorf("RegisterType() with type %v already exists", typ.String())
	}

	r.typeToValidator[typ] = r.toValidator(definition)
	return nil
}

func (r *Registry) toValidator(definition *Definition) *ValueAnyVldr {
	typ := definition.Type()
	selfRules := definition.SelfRules()
	if len(definition.RuleMap()) > 0 {
		selfRules = append(selfRules, mustNewValidator(func() (FieldsAnyVldr, error) { return FieldsAnyWithErr(definition.typ, definition.RuleMap()) }))
	}
	v := mustNewValidator(func() (ValueAnyVldr, error) { return ValueAnyWithErr(typ, selfRules...) })
	return &v
}

// nilType stands-in for nil values--it is private, so it can't be registered outside of this package
// and validation returns a "not found in Registry" error
type nilType struct{}

// nilValue is used when the data is invalid (e.g. nil), since the type can't be inferred
var nilValue = reflect.ValueOf(nilType{})

// ValidateAny validates the data with the correct validator
func (r *Registry) ValidateAny(data any) ErrorMap {
	value := reflect.ValueOf(data)
	if !value.IsValid() {
		value = nilValue
	}
	// value is used here, so can't use validateValue to save reflect.TypeOf call
	return validateValueResult(r.DefaultedValidator(value.Type()), value)
}

// ValidateValue validates the data value with the correct validator (assumes TypeCheck is called)
func (r *Registry) ValidateValue(value reflect.Value) ErrorMap {
	return r.DefaultedValidator(value.Type()).ValidateValue(value)
}

// TypeCheck checks whether the type is valid for the Rule
func (r *Registry) TypeCheck(typ reflect.Type) *RuleTypeError {
	return r.DefaultedValidator(typ).TypeCheck(typ)
}

// ValidateMerge validates the data value with the correct validator, also doing a merge with the errorMap (assumes TypeCheck is called)
func (r *Registry) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	r.DefaultedValidator(value.Type()).ValidateMerge(value, key, errorMap)
}

// DefaultedValidator returns the validator for the value, defaulted by r.DefaultValidator, then DefaultValidator
func (r *Registry) DefaultedValidator(typ reflect.Type) Validator {
	validator := r.Validator(typ)
	if validator != nil {
		return validator
	}
	if r.DefaultValidator != nil {
		return r.DefaultValidator
	}
	return DefaultValidator
}

// Validator returns the validator for the type (not defaulted)
func (r *Registry) Validator(typ reflect.Type) Validator {
	if typ == nil || r.typeToValidator == nil {
		return nil
	}
	// Registry only contains indirect types, make the function safe
	typ = indirectType(typ)
	validator := r.typeToValidator[typ]
	if validator == nil {
		return nil
	}
	return validator
}

// Backed returns a RegistryBacker for the Registry
func (r *Registry) Backed() RegistryBacker { return RegistryBacker{Registry: r} }

// RegistryBacker safely validates against a Registry.
// typ allows handling `nil` values, as the reflect.Type can't be inferred.
// typ is set via stampRegistryBackers() through FieldsAnyWithErr().
// Intended to be used with Registry.Backed() for easy reading.
type RegistryBacker struct {
	Registry *Registry
	typ      reflect.Type
}

// ValidateAny validates the data
func (b RegistryBacker) ValidateAny(data any) ErrorMap {
	return b.Registry.DefaultedValidator(b.typ).ValidateAny(data)
}

// ValidateValue validates the data value (assumes TypeCheck is called)
func (b RegistryBacker) ValidateValue(value reflect.Value) ErrorMap {
	if !value.IsValid() {
		return nil
	}
	errorMap := ErrorMap{}
	b.ValidateMerge(value, "", errorMap)
	return errorMap.ToNil()
}

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (b RegistryBacker) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	if !value.IsValid() {
		return
	}
	b.Registry.DefaultedValidator(b.typ).ValidateMerge(value, key, errorMap)
}

// TypeCheck checks whether the type is valid for the Rule
func (b RegistryBacker) TypeCheck(typ reflect.Type) *RuleTypeError { return b.Registry.TypeCheck(typ) }
