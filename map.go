package firm

import (
	"errors"
	"fmt"
	"reflect"
)

// Keys returns a new KeysVldr, panics if there is an error
func Keys[T map[K]V, K comparable, V any](keyRules ...Rule) KeysVldr[T, K, V] {
	return mustNewValidator(func() (KeysVldr[T, K, V], error) { return KeysWithErr[T, K, V](keyRules...) })
}

// KeysWithErr returns a new KeysVldr
func KeysWithErr[T map[K]V, K comparable, V any](keyRules ...Rule) (KeysVldr[T, K, V], error) {
	var zero T
	keysAnyV, err := KeysAnyWithErr(reflect.TypeOf(zero), keyRules...)
	return KeysVldr[T, K, V]{KeysAnyVldr: keysAnyV}, err
}

// KeysAny returns a new KeysAnyVldr, panics if there is an error
func KeysAny(typ reflect.Type, keyRules ...Rule) KeysAnyVldr {
	return mustNewValidator(func() (KeysAnyVldr, error) { return KeysAnyWithErr(typ, keyRules...) })
}

// KeysAnyWithErr returns the KeysVldr validator without generics. Pointer types are indirected to their value type
func KeysAnyWithErr(typ reflect.Type, keyRules ...Rule) (KeysAnyVldr, error) {
	typ, err := mapType("Keys", typ)
	if err != nil {
		return KeysAnyVldr{}, err
	}

	for _, rule := range keyRules {
		if err := rule.TypeCheck(indirectType(typ.Key())); err != nil {
			return KeysAnyVldr{}, fmt.Errorf("Keys: key type: %w", err)
		}
	}
	return KeysAnyVldr{typ: typ, keyRules: stampRegistryBackers(keyRules, typ.Key())}, nil
}

// KeysVldr validates map keys
type KeysVldr[T map[K]V, K comparable, V any] struct{ KeysAnyVldr }

// Validate is firm.Validator(), but with a typed arg, so no type checking is done on runtime
func (s KeysVldr[T, K, V]) Validate(data T) ErrorMap { return validate(s, data) }

// KeysAnyVldr is a KeysVldr without generics
type KeysAnyVldr struct {
	typ      reflect.Type
	keyRules []Rule
}

// Type returns the Type the Validator handles
func (s KeysAnyVldr) Type() reflect.Type { return s.typ }

// ValidateAny validates the data
func (s KeysAnyVldr) ValidateAny(data any) ErrorMap { return validateAny(s, data) }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (s KeysAnyVldr) ValidateValue(value reflect.Value) ErrorMap { return validateValue(s, value) }

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (s KeysAnyVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	if value = safeValidateMergeValue(value); !value.IsValid() {
		return
	}
	for iter := value.MapRange(); iter.Next(); {
		// indirect to ensure passing a non-pointer down to a Rule
		validateMerge(indirect(iter.Key()), joinKeys(key, mapErrorKey(iter.Key())), errorMap, s.keyRules)
	}
}

// TypeCheck checks whether the type is valid for the Rule
func (s KeysAnyVldr) TypeCheck(typ reflect.Type) *RuleTypeError {
	return TypeCheck("KeysAnyVldr", typ, s.typ, "Map")
}

// KeyRules returns the rules for each key in the Map
func (s KeysAnyVldr) KeyRules() []Rule { return s.keyRules }

// Values returns a new ValuesVldr, panics if there is an error
func Values[T map[K]V, K comparable, V any](valueRules ...Rule) ValuesVldr[T, K, V] {
	return mustNewValidator(func() (ValuesVldr[T, K, V], error) { return ValuesWithErr[T, K, V](valueRules...) })
}

// ValuesWithErr returns a new ValuesVldr
func ValuesWithErr[T map[K]V, K comparable, V any](valueRules ...Rule) (ValuesVldr[T, K, V], error) {
	var zero T
	valuesAnyV, err := ValuesAnyWithErr(reflect.TypeOf(zero), valueRules...)
	return ValuesVldr[T, K, V]{ValuesAnyVldr: valuesAnyV}, err
}

// ValuesAny returns a new ValuesAnyVldr, panics if there is an error
func ValuesAny(typ reflect.Type, valueRules ...Rule) ValuesAnyVldr {
	return mustNewValidator(func() (ValuesAnyVldr, error) { return ValuesAnyWithErr(typ, valueRules...) })
}

// ValuesAnyWithErr returns the ValuesVldr validator without generics. Pointer types are indirected to their value type
func ValuesAnyWithErr(typ reflect.Type, valueRules ...Rule) (ValuesAnyVldr, error) {
	typ, err := mapType("Values", typ)
	if err != nil {
		return ValuesAnyVldr{}, err
	}

	for _, rule := range valueRules {
		if err := rule.TypeCheck(indirectType(typ.Elem())); err != nil {
			return ValuesAnyVldr{}, fmt.Errorf("Values: value type: %w", err)
		}
	}
	return ValuesAnyVldr{typ: typ, valueRules: stampRegistryBackers(valueRules, typ.Elem())}, nil
}

// ValuesVldr validates map values
type ValuesVldr[T map[K]V, K comparable, V any] struct{ ValuesAnyVldr }

// Validate is firm.Validator(), but with a typed arg, so no type checking is done on runtime
func (s ValuesVldr[T, K, V]) Validate(data T) ErrorMap { return validate(s, data) }

// ValuesAnyVldr is a ValuesVldr without generics
type ValuesAnyVldr struct {
	typ        reflect.Type
	valueRules []Rule
}

// Type returns the Type the Validator handles
func (s ValuesAnyVldr) Type() reflect.Type { return s.typ }

// ValidateAny validates the data
func (s ValuesAnyVldr) ValidateAny(data any) ErrorMap { return validateAny(s, data) }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (s ValuesAnyVldr) ValidateValue(value reflect.Value) ErrorMap { return validateValue(s, value) }

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (s ValuesAnyVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	if value = safeValidateMergeValue(value); !value.IsValid() {
		return
	}
	for iter := value.MapRange(); iter.Next(); {
		// indirect to ensure passing a non-pointer down to a Rule
		validateMerge(indirect(iter.Value()), joinKeys(key, mapErrorKey(iter.Key())), errorMap, s.valueRules)
	}
}

// TypeCheck checks whether the type is valid for the Rule
func (s ValuesAnyVldr) TypeCheck(typ reflect.Type) *RuleTypeError {
	return TypeCheck("ValuesAnyVldr", typ, s.typ, "Map")
}

// ValueRules returns the rules for each value in the Map
func (s ValuesAnyVldr) ValueRules() []Rule { return s.valueRules }

// KeyValue represents a key/value-pair of a map to validate
type KeyValue[K comparable, V any] struct {
	Key   K
	Value V
}

// KeyValues returns a new KeyValuesVldr, panics if there is an error
func KeyValues[T map[K]V, K comparable, V any](entryRules ...Rule) KeyValuesVldr[T, K, V] {
	return mustNewValidator(func() (KeyValuesVldr[T, K, V], error) { return KeyValuesWithErr[T, K, V](entryRules...) })
}

// KeyValuesWithErr returns a new KeyValuesVldr
func KeyValuesWithErr[T map[K]V, K comparable, V any](entryRules ...Rule) (KeyValuesVldr[T, K, V], error) {
	var zero T
	// Not calling KeyValuesAnyWithErr instead due to anonymous KeyValue struct instead of named type
	kvAnyV, err := keyValuesAnyWithErr(reflect.TypeOf(zero), reflect.TypeFor[KeyValue[K, V]](), entryRules...)
	return KeyValuesVldr[T, K, V]{KeyValuesAnyVldr: kvAnyV}, err
}

// KeyValuesAny returns a new KeyValuesAnyVldr, panics if there is an error
func KeyValuesAny(typ reflect.Type, entryRules ...Rule) KeyValuesAnyVldr {
	return mustNewValidator(func() (KeyValuesAnyVldr, error) { return KeyValuesAnyWithErr(typ, entryRules...) })
}

// KeyValuesAnyWithErr returns the KeyValuesVldr validator without generics. Pointer types are indirected to their value type.
func KeyValuesAnyWithErr(typ reflect.Type, entryRules ...Rule) (KeyValuesAnyVldr, error) {
	typ, err := mapType("KeyValues", typ)
	if err != nil {
		return KeyValuesAnyVldr{}, err
	}

	// use anonymous KeyValue struct, struct { Key K; Value V },
	// since generic type can't be structed via reflection
	// declined: golang/go#45591 — proposal: spec: generics: Anonymous generic aggregate types
	// declined: golang/go#54393 — proposal: reflect: add generic type arg info to reflect.Type
	keyValueType := reflect.StructOf([]reflect.StructField{
		{Name: "Key", Type: typ.Key()},
		{Name: "Value", Type: typ.Elem()},
	})
	return keyValuesAnyWithErr(typ, keyValueType, entryRules...)
}

func keyValuesAnyWithErr(typ, keyValueType reflect.Type, entryRules ...Rule) (KeyValuesAnyVldr, error) {
	for _, rule := range entryRules {
		if err := rule.TypeCheck(keyValueType); err != nil {
			return KeyValuesAnyVldr{}, fmt.Errorf("KeyValues: entry type: %w", err)
		}
	}
	return KeyValuesAnyVldr{typ: typ, keyValueType: keyValueType, entryRules: stampRegistryBackers(entryRules, keyValueType)}, nil
}

// KeyValuesVldr validates map entries, passing each entry as a KeyValue to validate
type KeyValuesVldr[T map[K]V, K comparable, V any] struct{ KeyValuesAnyVldr }

// Validate is firm.Validator(), but with a typed arg, so no type checking is done on runtime
func (s KeyValuesVldr[T, K, V]) Validate(data T) ErrorMap { return validate(s, data) }

// KeyValuesAnyVldr is a KeyValuesVldr without generics
type KeyValuesAnyVldr struct {
	typ          reflect.Type
	keyValueType reflect.Type
	entryRules   []Rule
}

// Type returns the Type the Validator handles
func (s KeyValuesAnyVldr) Type() reflect.Type { return s.typ }

// ValidateAny validates the data
func (s KeyValuesAnyVldr) ValidateAny(data any) ErrorMap { return validateAny(s, data) }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (s KeyValuesAnyVldr) ValidateValue(value reflect.Value) ErrorMap { return validateValue(s, value) }

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (s KeyValuesAnyVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	if value = safeValidateMergeValue(value); !value.IsValid() {
		return
	}
	for iter := value.MapRange(); iter.Next(); {
		kv := reflect.New(s.keyValueType).Elem()
		kv.FieldByName("Key").Set(iter.Key())
		kv.FieldByName("Value").Set(iter.Value())
		validateMerge(kv, joinKeys(key, mapErrorKey(iter.Key())), errorMap, s.entryRules)
	}
}

// TypeCheck checks whether the type is valid for the Rule
func (s KeyValuesAnyVldr) TypeCheck(typ reflect.Type) *RuleTypeError {
	return TypeCheck("KeyValuesAnyVldr", typ, s.typ, "Map")
}

// KeyValueRules returns the rules for each entry in the Map
func (s KeyValuesAnyVldr) KeyValueRules() []Rule { return s.entryRules }

// KeyValueType returns the KeyValue type each entry in the Map is validated as
func (s KeyValuesAnyVldr) KeyValueType() reflect.Type { return s.keyValueType }

// mapType returns the indirected type, ensuring the type is a Map
func mapType(vdlrClass string, typ reflect.Type) (reflect.Type, error) {
	if typ == nil {
		return nil, errors.New(vdlrClass + ": type, nil, is not a Map")
	}
	typ = indirectType(typ)
	if typ.Kind() != reflect.Map {
		return nil, fmt.Errorf("%s: type, %v, is not a Map", vdlrClass, typ.String())
	}
	return typ, nil
}

// mapErrorKey returns the error path key of a map key, e.g. [key], mirroring a slice's [i].
// Pointer keys are indirected, as addresses are unstable across runs and unreadable
func mapErrorKey(key reflect.Value) string {
	key = indirect(key)
	if !key.IsValid() {
		return "[<nil>]"
	}
	return "[" + fmt.Sprintf("%v", key.Interface()) + "]"
}
