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
		panic(err)
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

	validator, err := r.toValidator(definition)
	if err != nil {
		return fmt.Errorf("RegisterType() with type %v: %w", typ.String(), err)
	}
	r.typeToValidator[typ] = validator
	return nil
}

func (r *Registry) toValidator(definition *Definition) (*ValueAnyVldr, error) {
	typ := definition.Type()
	mergedRules, err := r.mergedRules(definition)
	if err != nil {
		return nil, err
	}
	v, err := ValueAnyWithErr(typ, mergedRules...)
	if err != nil {
		return nil, err
	}
	if definition.errOnNilSelf {
		v = v.ErrOnNilSelf()
	}
	if err := checkRecursion(r, typ, v.rules); err != nil {
		return nil, err
	}
	return &v, nil
}

// mergedRules returns the Definition's self rules, with the fields validator appended to it
func (r *Registry) mergedRules(definition *Definition) ([]Rule, error) {
	selfRules := definition.SelfRules()
	if len(definition.ruleMap) == 0 && definition.errOnNilFields == nil {
		return selfRules, nil
	}
	fieldsV, err := FieldsAnyWithErr(definition.typ, definition.RuleMap())
	if err != nil {
		return nil, err
	}
	if definition.errOnNilFields != nil {
		fieldsV, err = fieldsV.ErrOnNilWithErr(definition.errOnNilFields...)
		if err != nil {
			return nil, err
		}
	}
	return append(selfRules, fieldsV), nil
}

// nilType stands-in for nil values--it is private, so it can't be registered outside of this package
// and validation returns a "not found in Registry" error
type nilType struct{}

// nilValue is used when the data is nil, since the type can't be inferred
var nilValue = reflect.ValueOf(nilType{})

// nilValueType is used to stand-in values, without naming nilType in errors
var nilValueType = reflect.TypeFor[nilType]()

// ValidateAny validates the data with the correct validator
func (r *Registry) ValidateAny(data any) ErrorMap {
	value := reflect.ValueOf(data)
	if !value.IsValid() {
		// the type can't be inferred from nil, and nilType stand-ins can't be registered,
		// so a "not found in Registry" error is returned
		value = nilValue
	}
	return r.DefaultedValidator(value.Type()).ValidateAny(value.Interface())
}

// ValidateValue validates the data value with the correct validator (assumes TypeCheck is called)
func (r *Registry) ValidateValue(value reflect.Value) ErrorMap { return ImplValidateValue(r, value) }

// TypeCheck checks whether the type is valid for the Rule
func (r *Registry) TypeCheck(typ reflect.Type) *RuleTypeError {
	return r.DefaultedValidator(typ).TypeCheck(typ)
}

// ValidateMerge validates the data value with the correct validator, also doing a merge with the errorMap (assumes TypeCheck is called)
func (r *Registry) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	MustValidValue(value)
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
// typ is set via stampRegistryBackers() through the *WithErr constructors.
// When typ is not set, the data's own type routes the validation via the Registry's logic.
// Intended to be used with Registry.Backed() for easy reading.
type RegistryBacker struct {
	Registry *Registry
	typ      reflect.Type
}

// Type returns the Type the RegistryBacker handles, nil when not stamped
func (b RegistryBacker) Type() reflect.Type { return b.typ }

// ValidateAny validates the data
func (b RegistryBacker) ValidateAny(data any) ErrorMap {
	if b.typ == nil {
		return b.Registry.ValidateAny(data)
	}
	return b.Registry.DefaultedValidator(b.typ).ValidateAny(data)
}

// ValidateValue validates the data value (assumes TypeCheck is called)
func (b RegistryBacker) ValidateValue(value reflect.Value) ErrorMap {
	return ImplValidateValue(b, value)
}

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (b RegistryBacker) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	if b.typ == nil {
		b.Registry.ValidateMerge(value, key, errorMap)
		return
	}
	b.Registry.DefaultedValidator(b.typ).ValidateMerge(value, key, errorMap)
}

// TypeCheck checks whether the type is valid for the Rule
func (b RegistryBacker) TypeCheck(typ reflect.Type) *RuleTypeError {
	if b.typ == nil {
		return b.Registry.TypeCheck(typ)
	}
	return TypeCheck("RegistryBacker", typ, b.typ, "")
}

// RuleLister is traversed by checkRecursion--implement AllRules() on a custom Validator to join the cycle check
type RuleLister interface{ AllRules() []Rule }

// backerTarget is the validation edge a RegistryBacker traverses: the type validated in a registry
type backerTarget struct {
	registry *Registry
	typ      reflect.Type
}

const maxRecursionDepth = 100

// checkRecursion errors if backing from rules of the registering typ reaches the typ again,
// as validation would infinitely recurse
func checkRecursion(r *Registry, typ reflect.Type, rules []Rule) error {
	check := recursionCheck{origin: backerTarget{r, typ}, visited: map[backerTarget]bool{}}
	return check.walk(rules, 0)
}

type recursionCheck struct {
	origin  backerTarget
	visited map[backerTarget]bool
}

func (c recursionCheck) walk(rules []Rule, depth int) error {
	if depth > maxRecursionDepth {
		return fmt.Errorf("Registry: recursion check exceeds depth %v on type, %v--check for a cycle in a custom RuleLister's AllRules()",
			maxRecursionDepth, c.origin.typ.String())
	}
	for _, rule := range rules {
		if backer, ok := rule.(RegistryBacker); ok {
			if err := c.walkBacker(backer, depth); err != nil {
				return err
			}
			continue
		}
		if lister, ok := rule.(RuleLister); ok {
			if err := c.walk(lister.AllRules(), depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c recursionCheck) walkBacker(backer RegistryBacker, depth int) error {
	target := backerTarget{backer.Registry, backer.typ}
	if target == c.origin {
		return fmt.Errorf("Registry: type, %v, recurses back to itself via RegistryBacker", c.origin.typ.String())
	}
	if c.visited[target] {
		return nil
	}
	c.visited[target] = true
	lister, ok := backer.Registry.Validator(backer.typ).(RuleLister)
	if !ok {
		return nil
	}
	return c.walk(lister.AllRules(), depth+1)
}
