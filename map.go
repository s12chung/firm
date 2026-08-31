package firm

import (
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

	keyRules, err = TypeCheckRules(typ.Key(), keyRules, "Keys: key type")
	if err != nil {
		return KeysAnyVldr{}, err
	}
	return KeysAnyVldr{typ: typ, keyRules: keyRules}, nil
}

// KeysVldr validates map keys
type KeysVldr[T map[K]V, K comparable, V any] struct{ KeysAnyVldr }

// Validate is firm.Validator(), but with a typed arg, so no type checking is done on runtime
func (s KeysVldr[T, K, V]) Validate(data T) ErrorMap { return ImplValidate(s, data) }

// KeysAnyVldr is a KeysVldr without generics
type KeysAnyVldr struct {
	typ      reflect.Type
	keyRules []Rule
}

// Type returns the Type the Validator handles
func (s KeysAnyVldr) Type() reflect.Type { return s.typ }

// ValidateAny validates the data
func (s KeysAnyVldr) ValidateAny(data any) ErrorMap { return ImplValidateAny(s, data) }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (s KeysAnyVldr) ValidateValue(value reflect.Value) ErrorMap { return ImplValidateValue(s, value) }

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (s KeysAnyVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	for iter := value.MapRange(); iter.Next(); {
		// indirect to ensure passing a non-pointer down to a Rule
		keyValue := indirect(iter.Key())
		ImplValidateMerge(keyValue, joinKeys(key, mapErrorKey(keyValue)), errorMap, s.keyRules)
	}
}

// TypeCheck checks whether the type is valid for the Rule
func (s KeysAnyVldr) TypeCheck(typ reflect.Type) *RuleTypeError {
	return TypeCheck("KeysAnyVldr", typ, s.typ, "Map")
}

// KeyRules returns the rules for each key in the Map
func (s KeysAnyVldr) KeyRules() []Rule { return s.keyRules }

// AllRules returns all rules of the validator, the rules of each key
func (s KeysAnyVldr) AllRules() []Rule { return s.keyRules }

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

	valueRules, err = TypeCheckRules(typ.Elem(), valueRules, "Values: value type")
	if err != nil {
		return ValuesAnyVldr{}, err
	}
	return ValuesAnyVldr{typ: typ, valueRules: valueRules}, nil
}

// ValuesVldr validates map values
type ValuesVldr[T map[K]V, K comparable, V any] struct{ ValuesAnyVldr }

// Validate is firm.Validator(), but with a typed arg, so no type checking is done on runtime
func (s ValuesVldr[T, K, V]) Validate(data T) ErrorMap { return ImplValidate(s, data) }

// ValuesAnyVldr is a ValuesVldr without generics
type ValuesAnyVldr struct {
	typ        reflect.Type
	valueRules []Rule
}

// Type returns the Type the Validator handles
func (s ValuesAnyVldr) Type() reflect.Type { return s.typ }

// ValidateAny validates the data
func (s ValuesAnyVldr) ValidateAny(data any) ErrorMap { return ImplValidateAny(s, data) }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (s ValuesAnyVldr) ValidateValue(value reflect.Value) ErrorMap {
	return ImplValidateValue(s, value)
}

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (s ValuesAnyVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	for iter := value.MapRange(); iter.Next(); {
		// indirect to ensure passing a non-pointer down to a Rule
		keyValue := indirect(iter.Key())
		ImplValidateMerge(indirect(iter.Value()), joinKeys(key, mapErrorKey(keyValue)), errorMap, s.valueRules)
	}
}

// TypeCheck checks whether the type is valid for the Rule
func (s ValuesAnyVldr) TypeCheck(typ reflect.Type) *RuleTypeError {
	return TypeCheck("ValuesAnyVldr", typ, s.typ, "Map")
}

// ValueRules returns the rules for each value in the Map
func (s ValuesAnyVldr) ValueRules() []Rule { return s.valueRules }

// AllRules returns all rules of the validator, the rules of each value
func (s ValuesAnyVldr) AllRules() []Rule { return s.valueRules }

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

	keyValueRules, err = TypeCheckRules(typ, keyValueRules, "KeyValues: key-value pair type")
	if err != nil {
		return KeyValuesAnyVldr{}, err
	}
	return KeyValuesAnyVldr{typ: typ, keyValueRules: keyValueRules}, nil
}

// KeyValuesVldr validates map key-value pairs, passing each key-value pair down as a Map with only 1 key-value pair to validate
type KeyValuesVldr[T map[K]V, K comparable, V any] struct{ KeyValuesAnyVldr }

// Validate is firm.Validator(), but with a typed arg, so no type checking is done on runtime
func (s KeyValuesVldr[T, K, V]) Validate(data T) ErrorMap { return ImplValidate(s, data) }

// KeyValuesAnyVldr is a KeyValuesVldr without generics
type KeyValuesAnyVldr struct {
	typ           reflect.Type
	keyValueRules []Rule
}

// Type returns the Type the Validator handles
func (s KeyValuesAnyVldr) Type() reflect.Type { return s.typ }

// ValidateAny validates the data
func (s KeyValuesAnyVldr) ValidateAny(data any) ErrorMap { return ImplValidateAny(s, data) }

// ValidateValue validates the data value (assumes TypeCheck is called)
func (s KeyValuesAnyVldr) ValidateValue(value reflect.Value) ErrorMap {
	return ImplValidateValue(s, value)
}

// ValidateMerge validates the data value, also doing a merge with the errorMap (assumes TypeCheck is called)
func (s KeyValuesAnyVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	for iter := value.MapRange(); iter.Next(); {
		// passed down as a Map with only 1 key-value pair
		keyValue := reflect.MakeMapWithSize(s.typ, 1)
		keyValue.SetMapIndex(iter.Key(), iter.Value())
		ImplValidateMerge(keyValue, joinKeys(key, mapErrorKey(indirect(iter.Key()))), errorMap, s.keyValueRules)
	}
}

// TypeCheck checks whether the type is valid for the Rule
func (s KeyValuesAnyVldr) TypeCheck(typ reflect.Type) *RuleTypeError {
	return TypeCheck("KeyValuesAnyVldr", typ, s.typ, "Map")
}

// KeyValueRules returns the rules for each key-value pair in the Map
func (s KeyValuesAnyVldr) KeyValueRules() []Rule { return s.keyValueRules }

// AllRules returns all rules of the validator, the rules of each key-value pair
func (s KeyValuesAnyVldr) AllRules() []Rule { return s.keyValueRules }
