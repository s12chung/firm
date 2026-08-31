package firm

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
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
	fieldIndexes := map[string][]int{}
	for fieldName, rules := range ruleMap {
		field, found := typ.FieldByName(fieldName)
		if !found {
			return FieldsAnyVldr{}, fmt.Errorf("Fields: field, %v, not found in type: %v", fieldName, typ.String())
		}
		if !field.IsExported() {
			return FieldsAnyVldr{}, fmt.Errorf("Fields: field, %v, is unexported in type: %v", fieldName, typ.String())
		}
		rules, err := typeCheckRules(field.Type, rules, fmt.Sprintf("Fields: field, %v, in %v", fieldName, typ.String()))
		if err != nil {
			return FieldsAnyVldr{}, err
		}
		rm[fieldName] = rules
		fieldIndexes[fieldName] = field.Index
	}
	return FieldsAnyVldr{typ: typ, ruleMap: rm, fieldIndexes: fieldIndexes}, nil
}

// FieldsVldr validates structs
type FieldsVldr[T any] struct{ FieldsAnyVldr }

// Validate is firm.Validator(), but with a typed arg, so no type checking is done on runtime
func (s FieldsVldr[T]) Validate(data T) ErrorMap { return validate(s, data) }

// FieldsAnyVldr is a FieldsVldr without generics
type FieldsAnyVldr struct {
	typ     reflect.Type
	ruleMap map[string][]Rule
	// fieldIndexes maps the StructField Name to the Index of each field in ruleMap, cached to avoid per-validation FieldByName() lookups
	fieldIndexes map[string][]int
}

// Type returns the Type the Validator handles
func (s FieldsAnyVldr) Type() reflect.Type { return s.typ }

// ValidateAny validates the data
func (s FieldsAnyVldr) ValidateAny(data any) ErrorMap { return validateAny(s, data) }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (s FieldsAnyVldr) ValidateValue(value reflect.Value) ErrorMap { return validateValue(s, value) }

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (s FieldsAnyVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	for fieldName, rules := range s.ruleMap {
		// indirect to ensure passing a non-pointer down to a Rule
		validateMerge(indirect(fieldByIndex(value, s.fieldIndexes[fieldName])), joinKeys(key, fieldName), errorMap, rules)
	}
}

// TypeCheck checks whether the type is valid for the Rule
func (s FieldsAnyVldr) TypeCheck(typ reflect.Type) *RuleTypeError {
	return TypeCheck("FieldsAnyVldr", typ, s.typ, "Struct")
}

// RuleMap returns the rules mapped to each field
func (s FieldsAnyVldr) RuleMap() RuleMap {
	ruleMap := RuleMap{}
	maps.Copy(ruleMap, s.ruleMap)
	return ruleMap
}

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

	elementRules, err = typeCheckRules(typ.Elem(), elementRules, "Elems: element type")
	if err != nil {
		return ElemsAnyVldr{}, err
	}
	return ElemsAnyVldr{typ: typ, elementRules: elementRules}, nil
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
	for i := range value.Len() {
		// indirect to ensure passing a non-pointer down to a Rule
		validateMerge(indirect(value.Index(i)), joinKeys(key, sliceErrorKey(i)), errorMap, s.elementRules)
	}
}

// TypeCheck checks whether the type is valid for the Rule
func (s ElemsAnyVldr) TypeCheck(typ reflect.Type) *RuleTypeError {
	return TypeCheck("ElemsAnyVldr", typ, s.typ, "Slice or Array")
}

// ElementRules returns the rules each element in the Slice or Array
func (s ElemsAnyVldr) ElementRules() []Rule { return s.elementRules }

// AllRules returns all rules of the validator, the rules of each element
func (s ElemsAnyVldr) AllRules() []Rule { return s.elementRules }

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
	rules, err := typeCheckRules(typ, rules, "")
	if err != nil {
		return ValueAnyVldr{}, err
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
func (v ValueAnyVldr) ValidateValue(value reflect.Value) ErrorMap { return validateValue(v, value) }

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

// AllRules returns all rules of the validator
func (v ValueAnyVldr) AllRules() []Rule { return v.rules }

// RuleVldr is a Validator wrapper around Rule
type RuleVldr struct{ Rule }

// ValidateAny validates the data
func (r RuleVldr) ValidateAny(data any) ErrorMap { return validateAny(r, data) }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (r RuleVldr) ValidateValue(value reflect.Value) ErrorMap { return validateValue(r, value) }

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (r RuleVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	validateMerge(value, key, errorMap, []Rule{r.Rule})
}

// AllRules returns all rules of the validator, the wrapped Rule
func (r RuleVldr) AllRules() []Rule { return []Rule{r.Rule} }

func mustNewValidator[T any](f func() (T, error)) T {
	validator, err := f()
	if err != nil {
		panic(err.Error())
	}
	return validator
}

// ErrInvalidValue returns the ErrorMap that answers unsafe values, keyed at "Invalid"
func ErrInvalidValue() ErrorMap                       { return ErrorMap{"Invalid": TemplateError{Template: "is not valid"}} }
func mergeInvalidValue(key string, errorMap ErrorMap) { ErrInvalidValue().MergeInto(key, errorMap) }

func validateAny(validator Validator, data any) ErrorMap {
	value := reflect.ValueOf(data)
	if !value.IsValid() {
		return ErrInvalidValue()
	}
	return validateValueResult(validator, value)
}

func validateValueResult(validator Validator, value reflect.Value) ErrorMap {
	// Users often don't have control over whether any is a pointer, so we're generous via indirect
	value = indirect(value)
	if !value.IsValid() {
		return ErrInvalidValue()
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
	return errorMap.Finish()
}

func validateValue(validator Validator, value reflect.Value) ErrorMap {
	errorMap := ErrorMap{}
	validator.ValidateMerge(value, "", errorMap)
	return errorMap.Finish()
}

func validateMerge(value reflect.Value, key string, errorMap ErrorMap, rules []Rule) {
	// invalid values (e.g. from indirecting a nil pointer) never reach Rules
	if !value.IsValid() {
		mergeInvalidValue(key, errorMap)
		return
	}
	for _, rule := range rules {
		rule.ValidateValue(value).MergeInto(key, errorMap)
	}
}

func validate(validator Validator, data any) ErrorMap {
	// Users often don't have control over whether any is a pointer, so we're generous via indirect
	value := indirect(reflect.ValueOf(data))
	if !value.IsValid() {
		return ErrInvalidValue()
	}
	errorMap := ErrorMap{}
	validator.ValidateMerge(value, value.Type().String(), errorMap)
	return errorMap.Finish()
}

// typeCheckRules TypeChecks each rule against the indirected typ, wrapping any error with errContext,
// and returns rules with each RegistryBacker set to typ
func typeCheckRules(typ reflect.Type, rules []Rule, errContext string) ([]Rule, error) {
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
// Not stamped for custom validators due to typeCheckRules() never being called.
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
