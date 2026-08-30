package firm

import (
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
	typ, err := typeOfKind("Keys", typ, reflect.Map)
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
	typ, err := typeOfKind("Values", typ, reflect.Map)
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

// KeyValues returns a new KeyValuesVldr, panics if there is an error
func KeyValues[T map[K]V, K comparable, V any](keyValueRules ...Rule) KeyValuesVldr[T, K, V] {
	return mustNewValidator(func() (KeyValuesVldr[T, K, V], error) { return KeyValuesWithErr[T, K, V](keyValueRules...) })
}

// KeyValuesWithErr returns a new KeyValuesVldr
func KeyValuesWithErr[T map[K]V, K comparable, V any](keyValueRules ...Rule) (KeyValuesVldr[T, K, V], error) {
	var zero T
	kvAnyV, err := KeyValuesAnyWithErr(reflect.TypeOf(zero), keyValueRules...)
	return KeyValuesVldr[T, K, V]{KeyValuesAnyVldr: kvAnyV}, err
}

// KeyValuesAny returns a new KeyValuesAnyVldr, panics if there is an error
func KeyValuesAny(typ reflect.Type, keyValueRules ...Rule) KeyValuesAnyVldr {
	return mustNewValidator(func() (KeyValuesAnyVldr, error) { return KeyValuesAnyWithErr(typ, keyValueRules...) })
}

// KeyValuesAnyWithErr returns the KeyValuesVldr validator without generics. Pointer types are indirected to their value type
func KeyValuesAnyWithErr(typ reflect.Type, keyValueRules ...Rule) (KeyValuesAnyVldr, error) {
	typ, err := typeOfKind("KeyValues", typ, reflect.Map)
	if err != nil {
		return KeyValuesAnyVldr{}, err
	}

	for _, rule := range keyValueRules {
		if err := rule.TypeCheck(typ); err != nil {
			return KeyValuesAnyVldr{}, fmt.Errorf("KeyValues: key-value pair type: %w", err)
		}
	}
	return KeyValuesAnyVldr{typ: typ, keyValueRules: stampRegistryBackers(keyValueRules, typ)}, nil
}

// KeyValuesVldr validates map key-value pairs, passing each key-value pair down as a Map with only 1 key-value pair to validate
type KeyValuesVldr[T map[K]V, K comparable, V any] struct{ KeyValuesAnyVldr }

// Validate is firm.Validator(), but with a typed arg, so no type checking is done on runtime
func (s KeyValuesVldr[T, K, V]) Validate(data T) ErrorMap { return validate(s, data) }

// KeyValuesAnyVldr is a KeyValuesVldr without generics
type KeyValuesAnyVldr struct {
	typ           reflect.Type
	keyValueRules []Rule
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
		// passed down as a Map with only 1 key-value pair
		keyValue := reflect.MakeMapWithSize(s.typ, 1)
		keyValue.SetMapIndex(iter.Key(), iter.Value())
		validateMerge(keyValue, joinKeys(key, mapErrorKey(iter.Key())), errorMap, s.keyValueRules)
	}
}

// TypeCheck checks whether the type is valid for the Rule
func (s KeyValuesAnyVldr) TypeCheck(typ reflect.Type) *RuleTypeError {
	return TypeCheck("KeyValuesAnyVldr", typ, s.typ, "Map")
}

// KeyValueRules returns the rules for each key-value pair in the Map
func (s KeyValuesAnyVldr) KeyValueRules() []Rule { return s.keyValueRules }

// mapErrorKey returns the error path key of a map key, e.g. [key], mirroring a slice's [i].
// Pointer keys are indirected, as addresses are unstable across runs and unreadable
func mapErrorKey(key reflect.Value) string {
	key = indirect(key)
	if !key.IsValid() {
		return "[<nil>]"
	}
	return "[" + fmt.Sprintf("%v", key.Interface()) + "]"
}
