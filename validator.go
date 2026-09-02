package firm

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
)

// Fields returns a new FieldsVldr, panics if there is an error
func Fields[T any](ruleMap RuleMap) FieldsVldr[T] {
	return mustNewValidator(func() (FieldsVldr[T], error) { return FieldsWithErr[T](ruleMap) })
}

// FieldsWithErr returns a new FieldsVldr
func FieldsWithErr[T any](ruleMap RuleMap) (FieldsVldr[T], error) {
	var zero T
	fieldsAnyV, err := FieldsAnyWithErr(reflect.TypeOf(zero), ruleMap)
	return FieldsVldr[T]{FieldsAnyVldr: fieldsAnyV}, err
}

// FieldsAny returns a new FieldsAnyVldr, panics if there is an error
func FieldsAny(typ reflect.Type, ruleMap RuleMap) FieldsAnyVldr {
	return mustNewValidator(func() (FieldsAnyVldr, error) { return FieldsAnyWithErr(typ, ruleMap) })
}

// FieldsAnyWithErr returns a new FieldsAnyVldr. Pointer types are indirected to their value type
func FieldsAnyWithErr(typ reflect.Type, ruleMap RuleMap) (FieldsAnyVldr, error) {
	typ, err := typeOfKind("Fields", typ, reflect.Struct)
	if err != nil {
		return FieldsAnyVldr{}, err
	}

	rm := map[string][]Rule{}
	fieldIndices := map[string][]int{}
	for fieldName, rules := range ruleMap {
		field, found := typ.FieldByName(fieldName)
		if !found {
			return FieldsAnyVldr{}, fmt.Errorf("Fields: field, %v, not found in type: %v", fieldName, typ.String())
		}
		if !field.IsExported() {
			return FieldsAnyVldr{}, fmt.Errorf("Fields: field, %v, is unexported in type: %v", fieldName, typ.String())
		}
		rules, err := TypeCheckAndBack(field.Type, rules, fmt.Sprintf("Fields: field, %v, in %v", fieldName, typ.String()))
		if err != nil {
			return FieldsAnyVldr{}, err
		}
		rm[fieldName] = rules
		fieldIndices[fieldName] = field.Index // set here to ensure promoted fields set (promoted fields is complicated)
	}
	return FieldsAnyVldr{typ: typ, ruleMap: rm, fieldIndices: fieldIndices}, nil
}

// FieldsVldr validates structs
type FieldsVldr[T any] struct{ FieldsAnyVldr }

// Validate is firm.Validator(), but with a typed arg, so no type checking is done on runtime
func (s FieldsVldr[T]) Validate(data T) ErrorMap { return ImplValidate(s, s.errOnNilSelf, data) }

// ErrOnNil is FieldsAnyVldr.ErrOnNil(), but typed
func (s FieldsVldr[T]) ErrOnNil(fields ...string) FieldsVldr[T] {
	s.FieldsAnyVldr = s.FieldsAnyVldr.ErrOnNil(fields...)
	return s
}

// ErrOnNilSelf is FieldsAnyVldr.ErrOnNilSelf(), but typed
func (s FieldsVldr[T]) ErrOnNilSelf() FieldsVldr[T] {
	s.FieldsAnyVldr = s.FieldsAnyVldr.ErrOnNilSelf()
	return s
}

// FieldsAnyVldr is a FieldsVldr without generics
type FieldsAnyVldr struct {
	typ     reflect.Type
	ruleMap map[string][]Rule
	// fieldIndices maps the StructField Name to the Index of each field in ruleMap and ErrOnNil(),
	// cached to avoid per-validation FieldByName() lookups
	fieldIndices map[string][]int
	// errOnNilFields flags fields to merge ErrInvalidValue() on, when the field's value is invalid (often from a nil pointer)
	errOnNilFields map[string]bool
	// errOnNilSelf flags to merge ErrInvalidValue(), when the value itself is invalid (often from a nil pointer)
	errOnNilSelf bool
}

// Type returns the Type the Validator handles
func (s FieldsAnyVldr) Type() reflect.Type { return s.typ }

// ValidateAny validates the data
func (s FieldsAnyVldr) ValidateAny(data any) ErrorMap {
	return ImplValidateAny(s, s.errOnNilSelf, data)
}

// ValidateValue validates the data value (assumes TypeCheck is called)
func (s FieldsAnyVldr) ValidateValue(value reflect.Value) ErrorMap {
	return ImplValidateValue(s, value)
}

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (s FieldsAnyVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	MustValidValue(value)
	for fieldName, rules := range s.ruleMap {
		ImplValidateMergeIndirected(fieldByIndex(value, s.fieldIndices[fieldName]), joinKeys(key, fieldName), errorMap, rules, s.errOnNilFields[fieldName])
	}
}

// ErrOnNil flags fields to merge ErrInvalidValue() on, when the field's value is invalid
// (often from a nil pointer), instead of skipping it. Fields not in the RuleMap are nil-checked
// with no rules. Fields must be exported; promoted fields are allowed. Panics on no fields
// given, a field not found, or unexported in the type
func (s FieldsAnyVldr) ErrOnNil(fields ...string) FieldsAnyVldr {
	v, err := s.ErrOnNilWithErr(fields...)
	if err != nil {
		panic(err)
	}
	return v
}

// ErrOnNilWithErr is ErrOnNil(), but returns an error instead of panicking
func (s FieldsAnyVldr) ErrOnNilWithErr(fields ...string) (FieldsAnyVldr, error) {
	if len(fields) == 0 {
		return FieldsAnyVldr{}, errors.New("ErrOnNil: no fields given")
	}
	errOnNilFields := map[string]bool{}
	for _, fieldName := range fields {
		field, found := s.typ.FieldByName(fieldName)
		if !found {
			return FieldsAnyVldr{}, fmt.Errorf("ErrOnNil: field, %v, not found in type: %v", fieldName, s.typ.String())
		}
		if !field.IsExported() {
			return FieldsAnyVldr{}, fmt.Errorf("ErrOnNil: field, %v, is unexported in type: %v", fieldName, s.typ.String())
		}
		if _, found := s.ruleMap[fieldName]; !found {
			s.ruleMap[fieldName] = nil
			s.fieldIndices[fieldName] = field.Index // set here to ensure promoted fields set (promoted fields is complicated)
		}
		errOnNilFields[fieldName] = true
	}
	s.errOnNilFields = errOnNilFields
	return s, nil
}

// ErrOnNilSelf flags to merge ErrInvalidValue(), when the value itself is nil
// (often from a nil pointer), instead of skipping it
func (s FieldsAnyVldr) ErrOnNilSelf() FieldsAnyVldr { s.errOnNilSelf = true; return s }

// TypeCheck checks whether the type is valid for the Rule
func (s FieldsAnyVldr) TypeCheck(typ reflect.Type) *RuleTypeError {
	return TypeCheck("FieldsAnyVldr", typ, s.typ, "Struct")
}

// RuleMap returns a copy of the rules mapped to each field
func (s FieldsAnyVldr) RuleMap() RuleMap { return copyRuleMap(s.ruleMap) }

// AllRules returns all rules of the validator, the rules of every field
func (s FieldsAnyVldr) AllRules() []Rule {
	var rules []Rule
	for _, fieldRules := range s.ruleMap {
		rules = append(rules, fieldRules...)
	}
	return rules
}

// Elems returns a new ElemsVldr, panics if there is an error
func Elems[T []U, U any](elementRules ...Rule) ElemsVldr[T, U] {
	return mustNewValidator(func() (ElemsVldr[T, U], error) { return ElemsWithErr[T, U](elementRules...) })
}

// ElemsWithErr returns a new ElemsVldr
func ElemsWithErr[T []U, U any](elementRules ...Rule) (ElemsVldr[T, U], error) {
	var zero T
	elemsAnyV, err := ElemsAnyWithErr(reflect.TypeOf(zero), elementRules...)
	return ElemsVldr[T, U]{ElemsAnyVldr: elemsAnyV}, err
}

// ElemsAny returns a new ElemsAnyVldr, panics if there is an error
func ElemsAny(typ reflect.Type, elementRules ...Rule) ElemsAnyVldr {
	return mustNewValidator(func() (ElemsAnyVldr, error) { return ElemsAnyWithErr(typ, elementRules...) })
}

// ElemsAnyWithErr returns the ElemsVldr validator without generics. Pointer types are indirected to their value type
func ElemsAnyWithErr(typ reflect.Type, elementRules ...Rule) (ElemsAnyVldr, error) {
	typ, err := typeOfKind("Elems", typ, reflect.Slice, reflect.Array)
	if err != nil {
		return ElemsAnyVldr{}, err
	}

	elementRules, err = TypeCheckAndBack(typ.Elem(), elementRules, "Elems: element type")
	if err != nil {
		return ElemsAnyVldr{}, err
	}
	return ElemsAnyVldr{typ: typ, elementRules: elementRules}, nil
}

// ElemsVldr validates slices and arrays
type ElemsVldr[T []U, U any] struct{ ElemsAnyVldr }

// Validate is firm.Validator(), but with a typed arg, so no type checking is done on runtime
func (s ElemsVldr[T, U]) Validate(data T) ErrorMap { return ImplValidate(s, s.errOnNilSelf, data) }

// ErrOnNil is ElemsAnyVldr.ErrOnNil(), but typed
func (s ElemsVldr[T, U]) ErrOnNil() ElemsVldr[T, U] {
	s.ElemsAnyVldr = s.ElemsAnyVldr.ErrOnNil()
	return s
}

// ErrOnNilSelf is ElemsAnyVldr.ErrOnNilSelf(), but typed
func (s ElemsVldr[T, U]) ErrOnNilSelf() ElemsVldr[T, U] {
	s.ElemsAnyVldr = s.ElemsAnyVldr.ErrOnNilSelf()
	return s
}

// ElemsAnyVldr is an ElemsVldr without generics
type ElemsAnyVldr struct {
	typ          reflect.Type
	elementRules []Rule
	// errOnNil flags to merge ErrInvalidValue(), when an element's value is invalid (often from a nil pointer)
	errOnNil bool
	// errOnNilSelf flags to merge ErrInvalidValue(), when the value itself is invalid (often from a nil pointer)
	errOnNilSelf bool
}

// Type returns the Type the Validator handles
func (s ElemsAnyVldr) Type() reflect.Type { return s.typ }

// ValidateAny validates the data
func (s ElemsAnyVldr) ValidateAny(data any) ErrorMap { return ImplValidateAny(s, s.errOnNilSelf, data) }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (s ElemsAnyVldr) ValidateValue(value reflect.Value) ErrorMap { return ImplValidateValue(s, value) }

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (s ElemsAnyVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	MustValidValue(value)
	for i := range value.Len() {
		ImplValidateMergeIndirected(value.Index(i), joinKeys(key, sliceErrorKey(i)), errorMap, s.elementRules, s.errOnNil)
	}
}

// ErrOnNil flags to merge ErrInvalidValue(), when an element's value is invalid
// (often from a nil pointer), instead of skipping it
func (s ElemsAnyVldr) ErrOnNil() ElemsAnyVldr { s.errOnNil = true; return s }

// ErrOnNilSelf flags to merge ErrInvalidValue(), when the value itself is nil
// (often from a nil pointer), instead of skipping it
func (s ElemsAnyVldr) ErrOnNilSelf() ElemsAnyVldr { s.errOnNilSelf = true; return s }

// TypeCheck checks whether the type is valid for the Rule
func (s ElemsAnyVldr) TypeCheck(typ reflect.Type) *RuleTypeError {
	return TypeCheck("ElemsAnyVldr", typ, s.typ, "Slice or Array")
}

// ElementRules returns a copy of the rules each element in the Slice or Array
func (s ElemsAnyVldr) ElementRules() []Rule { return slices.Clone(s.elementRules) }

// AllRules returns all rules of the validator, the rules of each element
func (s ElemsAnyVldr) AllRules() []Rule { return s.ElementRules() }

// Value returns a new ValueVldr, panics if there is an error
func Value[T any](rules ...Rule) ValueVldr[T] {
	return mustNewValidator(func() (ValueVldr[T], error) { return ValueWithErr[T](rules...) })
}

// ValueWithErr returns a new ValueVldr
func ValueWithErr[T any](rules ...Rule) (ValueVldr[T], error) {
	var zero T
	valueAnyV, err := ValueAnyWithErr(reflect.TypeOf(zero), rules...)
	return ValueVldr[T]{ValueAnyVldr: valueAnyV}, err
}

// ValueAny returns a new ValueAnyVldr, panics if there is an error
func ValueAny(typ reflect.Type, rules ...Rule) ValueAnyVldr {
	return mustNewValidator(func() (ValueAnyVldr, error) { return ValueAnyWithErr(typ, rules...) })
}

// ValueAnyWithErr returns a ValueAnyVldr. Pointer types are indirected to their value type
func ValueAnyWithErr(typ reflect.Type, rules ...Rule) (ValueAnyVldr, error) {
	if typ == nil {
		return ValueAnyVldr{}, errors.New("Value: type is nil")
	}
	typ = indirectType(typ)
	rules, err := TypeCheckAndBack(typ, rules, "")
	if err != nil {
		return ValueAnyVldr{}, err
	}
	return ValueAnyVldr{typ: typ, rules: rules}, nil
}

// ValueVldr validates a simple value
type ValueVldr[T any] struct{ ValueAnyVldr }

// Validate is firm.Validator(), but with a typed arg, so no type checking is done on runtime
func (v ValueVldr[T]) Validate(data T) ErrorMap { return ImplValidate(v, v.errOnNilSelf, data) }

// ErrOnNilSelf is ValueAnyVldr.ErrOnNilSelf(), but typed
func (v ValueVldr[T]) ErrOnNilSelf() ValueVldr[T] {
	v.ValueAnyVldr = v.ValueAnyVldr.ErrOnNilSelf()
	return v
}

// ValueAnyVldr is a ValueVldr without generics
type ValueAnyVldr struct {
	typ   reflect.Type
	rules []Rule
	// errOnNilSelf flags to merge ErrInvalidValue(), when the value itself is invalid (often from a nil pointer)
	errOnNilSelf bool
}

// Type returns the Type the Validator handles
func (v ValueAnyVldr) Type() reflect.Type { return v.typ }

// ValidateAny validates the data
func (v ValueAnyVldr) ValidateAny(data any) ErrorMap { return ImplValidateAny(v, v.errOnNilSelf, data) }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (v ValueAnyVldr) ValidateValue(value reflect.Value) ErrorMap { return ImplValidateValue(v, value) }

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (v ValueAnyVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	ImplValidateMerge(value, key, errorMap, v.rules)
}

// ErrOnNilSelf flags to merge ErrInvalidValue(), when the value itself is nil
// (often from a nil pointer), instead of skipping it
func (v ValueAnyVldr) ErrOnNilSelf() ValueAnyVldr { v.errOnNilSelf = true; return v }

// TypeCheck checks whether the type is valid for the Rule
func (v ValueAnyVldr) TypeCheck(typ reflect.Type) *RuleTypeError {
	return TypeCheck("ValueAnyVldr", typ, v.typ, "")
}

// Rules returns a copy of the rules for ValueAnyVldr
func (v ValueAnyVldr) Rules() []Rule { return slices.Clone(v.rules) }

// AllRules returns all rules of the validator
func (v ValueAnyVldr) AllRules() []Rule { return v.Rules() }

// RuleVldr is a Validator wrapper around Rule
type RuleVldr struct {
	Rule

	// errOnNilSelf flags to merge ErrInvalidValue(), when the value itself is invalid (often from a nil pointer)
	errOnNilSelf bool
}

// ValidateAny validates the data
func (r RuleVldr) ValidateAny(data any) ErrorMap { return ImplValidateAny(r, r.errOnNilSelf, data) }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (r RuleVldr) ValidateValue(value reflect.Value) ErrorMap { return ImplValidateValue(r, value) }

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (r RuleVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	ImplValidateMerge(value, key, errorMap, []Rule{r.Rule})
}

// ErrOnNilSelf flags to merge ErrInvalidValue(), when the value itself is nil
// (often from a nil pointer), instead of skipping it
func (r RuleVldr) ErrOnNilSelf() RuleVldr { r.errOnNilSelf = true; return r }

// AllRules returns all rules of the validator, the wrapped Rule
func (r RuleVldr) AllRules() []Rule { return []Rule{r.Rule} }

func mustNewValidator[T any](f func() (T, error)) T {
	validator, err := f()
	if err != nil {
		panic(err)
	}
	return validator
}

// ErrInvalidValue returns the ErrorMap that answers unsafe values, keyed at "Invalid"
func ErrInvalidValue() ErrorMap                       { return ErrorMap{"Invalid": TemplateError{Template: "is not valid"}} }
func mergeInvalidValue(key string, errorMap ErrorMap) { errorMap.Merge(key, ErrInvalidValue()) }

// ImplValidateAny validates the data with the validator, the implementation of Validator.ValidateAny().
// Invalid values (often from nil pointers) are skipped, unless errOnNilSelf is set--then ErrInvalidValue() is returned
func ImplValidateAny(validator Validator, errOnNilSelf bool, data any) ErrorMap {
	value := reflect.ValueOf(data)
	if !value.IsValid() {
		if errOnNilSelf {
			return ErrInvalidValue()
		}
		return nil
	}
	return validateIndirected(validator, errOnNilSelf, value)
}

// ImplValidateValue validates the data value with the validator (assumes TypeCheck is called),
// the implementation of Rule.ValidateValue(). Panics on an invalid value--safe values are expected
func ImplValidateValue(validator Validator, value reflect.Value) ErrorMap {
	MustValidValue(value)
	errorMap := ErrorMap{}
	validator.ValidateMerge(value, "", errorMap)
	return errorMap.ToNil()
}

// ImplValidateMerge validates the data value with the rules, also doing a merge with the errorMap
// (assumes TypeCheck is called), the implementation of Validator.ValidateMerge(). Panics on an invalid value.
func ImplValidateMerge(value reflect.Value, key string, errorMap ErrorMap, rules []Rule) {
	MustValidValue(value)
	for _, rule := range rules {
		errorMap.Merge(key, rule.ValidateValue(value))
	}
}

const safeValuePanic = "invalid value--safe values are expected, see the Types, Pointers, and Safe Values section in the README"

// MustValidValue panics on an invalid value--a safe values contract violation. Implementations that
// recurse in ValidateMerge(), using the value before delegating (e.g. over fields or elements),
// call it on entry; implementations calling ImplValidateMerge() directly don't need to
func MustValidValue(value reflect.Value) {
	if !value.IsValid() {
		panic(safeValuePanic)
	}
}

// ImplValidateMergeIndirected calls ImplValidateMerge() after indirecting the value. On an invalid
// value (often from a nil pointer), ErrInvalidValue() is merged when errOnNil is set; skipped otherwise
func ImplValidateMergeIndirected(value reflect.Value, key string, errorMap ErrorMap, rules []Rule, errOnNil bool) {
	value = indirect(value)
	if !value.IsValid() {
		if errOnNil {
			mergeInvalidValue(key, errorMap)
		}
		return
	}
	ImplValidateMerge(value, key, errorMap, rules)
}

// ImplValidate validates the data with the validator, the implementation of ValidatorTyped.Validate()--
// no type checking is done on runtime. Invalid values (often from nil pointers) are skipped,
// unless errOnNilSelf is set--then ErrInvalidValue() is returned
func ImplValidate(validator Validator, errOnNilSelf bool, data any) ErrorMap {
	// Users often don't have control over whether any is a pointer, so we're generous via indirect
	value := indirect(reflect.ValueOf(data))
	if !value.IsValid() {
		if errOnNilSelf {
			return ErrInvalidValue()
		}
		return nil
	}
	errorMap := ErrorMap{}
	validator.ValidateMerge(value, value.Type().String(), errorMap)
	return errorMap.ToNil()
}

// validateIndirected indirects and TypeChecks the validator on the value, then merges with a key of
// the type's name. Used by validators that pick the validator by the value's type, e.g. Registry.ValidateAny().
// Invalid values (often from nil pointers) are skipped, unless errOnNilSelf is set--then ErrInvalidValue() is returned
func validateIndirected(validator Validator, errOnNilSelf bool, value reflect.Value) ErrorMap {
	// Users often don't have control over whether any is a pointer, so we're generous via indirect
	value = indirect(value)
	if !value.IsValid() {
		if errOnNilSelf {
			return ErrInvalidValue()
		}
		return nil
	}
	typ := value.Type()
	if err := validator.TypeCheck(typ); err != nil {
		return ErrorMap{"TypeCheck": err.TemplateError()}
	}

	// nilType stand-ins have no user-facing type, so they are not named in the key
	key := typ.String()
	if typ == nilValueType {
		key = ""
	}
	errorMap := ErrorMap{}
	validator.ValidateMerge(value, key, errorMap)
	return errorMap.ToNil()
}

// TypeCheckAndBack TypeChecks each rule against the indirected typ, wrapping any error with errContext,
// and returns rules with each RegistryBacker set to typ
func TypeCheckAndBack(typ reflect.Type, rules []Rule, errContext string) ([]Rule, error) {
	for _, rule := range rules {
		if err := rule.TypeCheck(indirectType(typ)); err != nil {
			if errContext == "" {
				return nil, err
			}
			return nil, fmt.Errorf("%s: %w", errContext, err)
		}
	}
	return stampRegistryBackers(rules, typ), nil
}

// stampRegistryBackers returns rules with each RegistryBacker set to typ
//
// Not stamped for custom validators due to TypeCheckAndBack() never being called.
// Custom validators may not have a `typ` restriction, so it makes no sense.
func stampRegistryBackers(rules []Rule, typ reflect.Type) []Rule {
	if len(rules) == 0 {
		return rules
	}
	stamped := make([]Rule, len(rules))
	for i, rule := range rules {
		if backer, ok := rule.(RegistryBacker); ok {
			backer.typ = indirectType(typ)
			rule = backer
		}
		stamped[i] = rule
	}
	return stamped
}
