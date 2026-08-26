package firm

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
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

// FieldsAnyWithErr returns a new FieldsAnyVldr
func FieldsAnyWithErr(typ reflect.Type, ruleMap RuleMap) (FieldsAnyVldr, error) {
	if typ == nil {
		return FieldsAnyVldr{}, errors.New("Fields: type, nil, is not a Struct")
	}
	if typ.Kind() != reflect.Struct {
		return FieldsAnyVldr{}, fmt.Errorf("Fields: type, %v, is not a Struct", typ.String())
	}

	rm := map[string]*[]Rule{}
	for fieldName, rules := range ruleMap {
		field, found := typ.FieldByName(fieldName)
		if !found {
			return FieldsAnyVldr{}, fmt.Errorf("Fields: field, %v, not found in type: %v", fieldName, typ.String())
		}
		for _, rule := range rules {
			if err := rule.TypeCheck(field.Type); err != nil {
				return FieldsAnyVldr{}, fmt.Errorf("Fields: field, %v, in %v: %w", fieldName, typ.String(), err)
			}
		}
		stamped := stampRegistryBackers(rules, field.Type)
		rm[fieldName] = &stamped
	}
	return FieldsAnyVldr{typ: typ, ruleMap: rm}, nil
}

// FieldsVldr validates structs
type FieldsVldr[T any] struct{ FieldsAnyVldr }

// Validate is firm.Validator(), but with a typed arg, so no type checking is done on runtime
func (s FieldsVldr[T]) Validate(data T) ErrorMap { return validate(s, data) }

// FieldsAnyVldr is a FieldsVldr without generics
type FieldsAnyVldr struct {
	typ     reflect.Type
	ruleMap map[string]*[]Rule
}

// Type returns the Type the Validator handles
func (s FieldsAnyVldr) Type() reflect.Type { return s.typ }

// ValidateAny validates the data
func (s FieldsAnyVldr) ValidateAny(data any) ErrorMap { return validateAny(s, data) }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (s FieldsAnyVldr) ValidateValue(value reflect.Value) ErrorMap { return validateValue(s, value) }

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (s FieldsAnyVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	if value = safeValidateMergeValue(value); !value.IsValid() {
		return
	}
	for fieldName, rules := range s.ruleMap {
		//nolint:godox // want the comment
		field, _ := value.Type().FieldByName(fieldName) // TODO: cache field indexes at FieldsAnyWithErr to avoid per-validation lookups?
		// indirect to ensure passing a non-pointer down to a Rule
		validateMerge(indirect(value.FieldByName(fieldName)), joinKeys(key, field.Name), errorMap, *rules)
	}
}

// TypeCheck checks whether the type is valid for the Rule
func (s FieldsAnyVldr) TypeCheck(typ reflect.Type) *RuleTypeError {
	return TypeCheck("FieldsAnyVldr", typ, s.typ, "Struct")
}

// RuleMap returns the rules mapped to each field
func (s FieldsAnyVldr) RuleMap() RuleMap {
	ruleMap := RuleMap{}
	for k, v := range s.ruleMap {
		ruleMap[k] = *v
	}
	return ruleMap
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

// ElemsAnyWithErr returns the ElemsVldr validator without generics
func ElemsAnyWithErr(typ reflect.Type, elementRules ...Rule) (ElemsAnyVldr, error) {
	if typ == nil {
		return ElemsAnyVldr{}, errors.New("Elems: type, nil, is not a Slice or Array")
	}
	// Ptr Validator types not allowed, Validator types just take pointers
	kind := indirectType(typ).Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return ElemsAnyVldr{}, fmt.Errorf("Elems: type, %v, is not a Slice or Array", typ.String())
	}

	for _, rule := range elementRules {
		if err := rule.TypeCheck(typ.Elem()); err != nil {
			return ElemsAnyVldr{}, fmt.Errorf("Elems: element type: %w", err)
		}
	}
	return ElemsAnyVldr{typ: typ, elementRules: stampRegistryBackers(elementRules, typ.Elem())}, nil
}

// ElemsVldr validates slices and arrays
type ElemsVldr[T []U, U any] struct{ ElemsAnyVldr }

// Validate is firm.Validator(), but with a typed arg, so no type checking is done on runtime
func (s ElemsVldr[T, U]) Validate(data T) ErrorMap { return validate(s, data) }

// ElemsAnyVldr is an ElemsVldr without generics
type ElemsAnyVldr struct {
	typ          reflect.Type
	elementRules []Rule
}

// Type returns the Type the Validator handles
func (s ElemsAnyVldr) Type() reflect.Type { return s.typ }

// ValidateAny validates the data
func (s ElemsAnyVldr) ValidateAny(data any) ErrorMap { return validateAny(s, data) }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (s ElemsAnyVldr) ValidateValue(value reflect.Value) ErrorMap { return validateValue(s, value) }

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (s ElemsAnyVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	if value = safeValidateMergeValue(value); !value.IsValid() {
		return
	}
	for i := range value.Len() {
		// indirect to ensure passing a non-pointer down to a Rule
		validateMerge(indirect(value.Index(i)), joinKeys(key, "["+strconv.Itoa(i)+"]"), errorMap, s.elementRules)
	}
}

// TypeCheck checks whether the type is valid for the Rule
func (s ElemsAnyVldr) TypeCheck(typ reflect.Type) *RuleTypeError {
	return TypeCheck("ElemsAnyVldr", typ, s.typ, "Slice or Array")
}

// ElementRules returns the rules each element in the Slice or Array
func (s ElemsAnyVldr) ElementRules() []Rule { return s.elementRules }

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

// ValueAnyWithErr returns a ValueAnyVldr
func ValueAnyWithErr(typ reflect.Type, rules ...Rule) (ValueAnyVldr, error) {
	if typ == nil {
		return ValueAnyVldr{}, errors.New("Value: type is nil, not recommended")
	}
	if typ.Kind() == reflect.Pointer {
		return ValueAnyVldr{}, fmt.Errorf("Value: type, %v, is a Pointer, not recommended", typ.String())
	}

	for _, rule := range rules {
		if err := rule.TypeCheck(typ); err != nil {
			return ValueAnyVldr{}, err
		}
	}
	return ValueAnyVldr{typ: typ, rules: rules}, nil
}

// ValueVldr validates a simple value
type ValueVldr[T any] struct{ ValueAnyVldr }

// Validate is firm.Validator(), but with a typed arg, so no type checking is done on runtime
func (v ValueVldr[T]) Validate(data T) ErrorMap { return validate(v, data) }

// ValueAnyVldr is a ValueVldr without generics
type ValueAnyVldr struct {
	typ   reflect.Type
	rules []Rule
}

// Type returns the Type the Validator handles
func (v ValueAnyVldr) Type() reflect.Type { return v.typ }

// ValidateAny validates the data
func (v ValueAnyVldr) ValidateAny(data any) ErrorMap { return validateAny(v, data) }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (v ValueAnyVldr) ValidateValue(value reflect.Value) ErrorMap {
	errorMap := ErrorMap{}
	v.ValidateMerge(value, "", errorMap)
	return errorMap.ToNil()
}

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (v ValueAnyVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	validateMerge(value, key, errorMap, v.rules)
}

// TypeCheck checks whether the type is valid for the Rule
func (v ValueAnyVldr) TypeCheck(typ reflect.Type) *RuleTypeError {
	return TypeCheck("ValueAnyVldr", typ, v.typ, "")
}

// Rules returns the rules for ValueAnyVldr
func (v ValueAnyVldr) Rules() []Rule { return v.rules }

// RuleVldr is a Validator wrapper around Rule
type RuleVldr struct{ Rule }

// ValidateAny validates the data
func (r RuleVldr) ValidateAny(data any) ErrorMap { return validateAny(r, data) }

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (r RuleVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	validateMerge(value, key, errorMap, []Rule{r.Rule})
}

func mustNewValidator[T any](f func() (T, error)) T {
	validator, err := f()
	if err != nil {
		panic(err.Error())
	}
	return validator
}

var errInvalidValue = ErrorMap{"ValidateAny": TemplateError{Template: "value is not valid"}}

func validateAny(validator Validator, data any) ErrorMap {
	value := reflect.ValueOf(data)
	if !value.IsValid() {
		return errInvalidValue
	}
	return validateValueResult(validator, value)
}

func validateValueResult(validator Validator, value reflect.Value) ErrorMap {
	// Users often don't have control over whether any is a pointer, so we're generous via indirect
	value = indirect(value)
	if !value.IsValid() {
		return errInvalidValue
	}
	typ := value.Type()
	if err := validator.TypeCheck(typ); err != nil {
		return ErrorMap{"TypeCheck": err.TemplateError()}
	}

	errorMap := ErrorMap{}
	validator.ValidateMerge(value, typ.String(), errorMap)
	return errorMap.Finish()
}

func validateValue(validator Validator, value reflect.Value) ErrorMap {
	errorMap := ErrorMap{}
	validator.ValidateMerge(value, "", errorMap)
	return errorMap.ToNil()
}

func safeValidateMergeValue(value reflect.Value) reflect.Value {
	if !value.IsValid() {
		return reflect.Value{}
	}
	if value.Kind() == reflect.Pointer && value.IsNil() {
		return reflect.Value{}
	}
	return indirect(value)
}
func validateMerge(value reflect.Value, key string, errorMap ErrorMap, rules []Rule) {
	for _, rule := range rules {
		rule.ValidateValue(value).MergeInto(key, errorMap)
	}
}

func validate(validator Validator, data any) ErrorMap {
	// Users often don't have control over whether any is a pointer, so we're generous via indirect
	value := indirect(reflect.ValueOf(data))
	if !value.IsValid() {
		return errInvalidValue
	}
	errorMap := ErrorMap{}
	validator.ValidateMerge(value, value.Type().String(), errorMap)
	return errorMap.Finish()
}

// stampRegistryBackers returns rules with each RegistryBacker set to typ
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
