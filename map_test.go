package firm

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

//
// Keys
//

var keysValidator = Keys[map[int]sliceValidatorElement](presentRule{})

var ptrKeysValidator = Keys[map[*int]sliceValidatorElement](presentRule{})

func TestKeysAny(t *testing.T) {
	typ := reflect.TypeFor[map[string]Child]()
	expected, err := KeysAnyWithErr(typ, presentRule{})
	require.NoError(t, err)
	require.Equal(t, expected, KeysAny(typ, presentRule{}))
	require.Equal(t, expected, KeysAny(reflect.TypeFor[*map[string]Child](), presentRule{}))

	require.Panics(t, func() { KeysAny(reflect.TypeFor[Child](), presentRule{}) })
}

func TestKeysAnyWithErr(t *testing.T) {
	noMatchingRule := onlyKindRule{kind: reflect.Bool}

	tcs := append(mapAnyWithErrTCs("Keys"), []anyWithErrTC{
		{name: "int_keys", data: map[int]Child{}, rules: []Rule{onlyKindRule{kind: reflect.Int}}},
		// key type, *string, is indirected before the rule's TypeCheck
		{name: "ptr_key_indirects_to_rule", data: map[*string]Child{}, rules: []Rule{onlyKindRule{kind: reflect.String}}},
		{name: "no_matching_rule", data: map[string]Child{}, rules: []Rule{noMatchingRule},
			err: fmt.Errorf("Keys: key type: %w", noMatchingRule.TypeCheck(reflect.TypeFor[string]()))},
	}...)
	testAnyWithErr(t, tcs, func(typ reflect.Type, rules ...Rule) (reflect.Type, []Rule, error) {
		validator, err := KeysAnyWithErr(typ, rules...)
		return validator.Type(), validator.KeyRules(), err
	})
}

func TestKeysVldr_Validate(t *testing.T) {
	errorKey := ErrorKey("map[int]firm.sliceValidatorElement.[0]." + presentRuleKey)
	tcs := []validateTC[map[int]sliceValidatorElement]{
		{name: "valid", data: map[int]sliceValidatorElement{1: {Int: 1}}},
		{name: "invalid", data: map[int]sliceValidatorElement{0: {Int: 1}}, result: ErrorMap{errorKey: *presentRuleError(errorKey)}},
	}
	testValidate(t, tcs, func() (ValidatorTyped[map[int]sliceValidatorElement], error) {
		return KeysWithErr[map[int]sliceValidatorElement](presentRule{})
	})
}

func TestKeysAnyVldr_ValidateAll(t *testing.T) {
	validator := keysValidator
	testMapValidateAllTypes(t, validator, (*map[int]sliceValidatorElement)(nil))

	testMapValidateAllCases(t, []mapValidatorTestCase{
		{name: "Full", errorKeys: nil, f: func() any {
			return map[int]sliceValidatorElement{1: {Int: 1}, 2: {Int: 2}}
		}},
		{name: "Empty", errorKeys: nil, f: func() any {
			return map[int]sliceValidatorElement{}
		}},
		{name: "Nil", errorKeys: nil, f: func() any {
			return map[int]sliceValidatorElement(nil)
		}},
		{name: "Key_Invalid", errorKeys: []string{"[0]"}, f: func() any {
			return map[int]sliceValidatorElement{0: {Int: 1}}
		}},
		{name: "Key_Invalid_mixed", errorKeys: []string{"[0]"}, f: func() any {
			return map[int]sliceValidatorElement{0: {Int: 1}, 1: {Int: 1}, 3: {Int: 2}}
		}},

		//
		// Pointer keys
		//
		{name: "Ptr_Key_valid", validator: ptrKeysValidator, errorKeys: nil, f: func() any {
			i := 1
			return map[*int]sliceValidatorElement{&i: {Int: 1}}
		}},
		{name: "Ptr_Key_nil", validator: ptrKeysValidator, invalidKeys: []string{"[<nil>]"}, f: func() any {
			return map[*int]sliceValidatorElement{nil: {Int: 1}}
		}},
		{name: "Ptr_Key_nil_mixed", validator: ptrKeysValidator, errorKeys: []string{"[0]"}, invalidKeys: []string{"[<nil>]"}, f: func() any {
			zero := 0
			return map[*int]sliceValidatorElement{nil: {Int: 1}, &zero: {Int: 1}}
		}},
	}, validator)
}

func TestKeysAnyVldr_TypeCheck(t *testing.T) {
	validator := keysValidator
	badCondition := "is not matching Map of type map[int]firm.sliceValidatorElement"

	testMapTypeCheck(t, validator, validator.typ, "KeysAnyVldr", badCondition)
}

//
// Values
//

var valuesValidator = Values[map[string]sliceValidatorElement](
	presentRule{}, Fields[sliceValidatorElement](RuleMap{"Int": {presentRule{}}}))

var ptrValuesValidator = Values[map[string]*sliceValidatorElement](
	presentRule{}, Fields[sliceValidatorElement](RuleMap{"Int": {presentRule{}}}))

var ptrPtValuesValidator = ValuesAny(
	reflect.TypeFor[*map[string]**sliceValidatorElement](),
	presentRule{}, Fields[sliceValidatorElement](RuleMap{"Int": {presentRule{}}}))

func TestValuesAny(t *testing.T) {
	typ := reflect.TypeFor[map[string]Child]()
	expected, err := ValuesAnyWithErr(typ, presentRule{})
	require.NoError(t, err)
	require.Equal(t, expected, ValuesAny(typ, presentRule{}))
	require.Equal(t, expected, ValuesAny(reflect.TypeFor[*map[string]Child](), presentRule{}))

	require.Panics(t, func() { ValuesAny(reflect.TypeFor[Child](), presentRule{}) })
}

func TestValuesAnyWithErr(t *testing.T) {
	noMatchingRule := onlyKindRule{kind: reflect.Bool}

	tcs := append(mapAnyWithErrTCs("Values"), []anyWithErrTC{
		// value type, *Child, is indirected before the rule's TypeCheck
		{name: "ptr_value_indirects_to_rule", data: map[string]*Child{}, rules: []Rule{onlyKindRule{kind: reflect.Struct}}},
		{name: "no_matching_rule", data: map[string]Child{}, rules: []Rule{noMatchingRule},
			err: fmt.Errorf("Values: value type: %w", noMatchingRule.TypeCheck(reflect.TypeFor[Child]()))},
	}...)
	testAnyWithErr(t, tcs, func(typ reflect.Type, rules ...Rule) (reflect.Type, []Rule, error) {
		validator, err := ValuesAnyWithErr(typ, rules...)
		return validator.Type(), validator.ValueRules(), err
	})
}

func TestValuesVldr_Validate(t *testing.T) {
	errorKey := ErrorKey("map[string]firm.sliceValidatorElement.[a].Int." + presentRuleKey)
	tcs := []validateTC[map[string]sliceValidatorElement]{
		{name: "valid", data: map[string]sliceValidatorElement{"a": {Int: 1}}},
		{name: "invalid", data: map[string]sliceValidatorElement{"a": {UInt: 1}}, result: ErrorMap{errorKey: *presentRuleError(errorKey)}},
	}
	testValidate(t, tcs, func() (ValidatorTyped[map[string]sliceValidatorElement], error) {
		return ValuesWithErr[map[string]sliceValidatorElement](Fields[sliceValidatorElement](RuleMap{"Int": {presentRule{}}}))
	})
}

func TestValuesAnyVldr_ValidateAll(t *testing.T) {
	validator := valuesValidator
	testMapValidateAllTypes(t, validator, (*map[string]sliceValidatorElement)(nil))

	testMapValidateAllCases(t, []mapValidatorTestCase{
		{name: "Full", errorKeys: nil, f: func() any {
			return map[string]sliceValidatorElement{"a": {Int: 1, UInt: 1}, "b": {Int: 2, UInt: 2}, "c": {Int: 3, UInt: 3}}
		}},
		{name: "Empty", errorKeys: nil, f: func() any {
			return map[string]sliceValidatorElement{}
		}},
		{name: "Nil", errorKeys: nil, f: func() any {
			return map[string]sliceValidatorElement(nil)
		}},
		{name: "Value_Not_Full", errorKeys: nil, f: func() any {
			return map[string]sliceValidatorElement{"a": {Int: 1}, "b": {Int: 2}}
		}},
		{name: "Value_Invalid", errorKeys: []string{"[a].Int", "[b].Int"}, f: func() any {
			return map[string]sliceValidatorElement{"a": {UInt: 1}, "b": {UInt: 2}}
		}},
		{name: "Value_Empty", errorKeys: []string{"[a]", "[b]", "[a].Int", "[b].Int"}, f: func() any {
			return map[string]sliceValidatorElement{"a": {}, "b": {}}
		}},

		//
		// Pointer values
		//
		{name: "Ptr_Value_valid", validator: ptrValuesValidator, errorKeys: nil, f: func() any {
			return map[string]*sliceValidatorElement{"a": {Int: 1}, "b": {Int: 2}}
		}},
		{name: "Ptr_Value_nil", validator: ptrValuesValidator,
			// the invalid value errors and never reaches the rules
			invalidKeys: []string{"[a]"}, f: func() any {
				return map[string]*sliceValidatorElement{"a": nil}
			}},
		{name: "Ptr_Value_nil_mixed", validator: ptrValuesValidator,
			errorKeys: []string{"[c]", "[c].Int"}, invalidKeys: []string{"[b]"}, f: func() any {
				return map[string]*sliceValidatorElement{"a": {Int: 1}, "b": nil, "c": {}}
			}},

		//
		// Pointer map, double-pointer values
		//
		{name: "Double_Ptr_Value_valid", validator: ptrPtValuesValidator, errorKeys: nil, f: func() any {
			return &map[string]**sliceValidatorElement{"a": toElemPtrPtr(sliceValidatorElement{Int: 1})}
		}},
		{name: "Double_Ptr_Value_nil_mixed", validator: ptrPtValuesValidator,
			errorKeys: []string{"[c]", "[c].Int"}, invalidKeys: []string{"[b]"}, f: func() any {
				return &map[string]**sliceValidatorElement{
					"a": toElemPtrPtr(sliceValidatorElement{Int: 1}), "b": nil, "c": toElemPtrPtr(sliceValidatorElement{})}
			}},
	}, validator)
}

func TestValuesAnyVldr_TypeCheck(t *testing.T) {
	validator := valuesValidator
	badCondition := "is not matching Map of type map[string]firm.sliceValidatorElement"

	testMapTypeCheck(t, validator, validator.typ, "ValuesAnyVldr", badCondition)
}

//
// KeyValues
//

type keyValueRule[K comparable, V any] struct {
	keyRules   []Rule
	valueRules []Rule
}

func newKeyValueRule[K comparable, V any](keyRules, valueRules []Rule) keyValueRule[K, V] {
	return keyValueRule[K, V]{
		keyRules:   stampRegistryBackers(keyRules, reflect.TypeFor[K]()),
		valueRules: stampRegistryBackers(valueRules, reflect.TypeFor[V]()),
	}
}

func (e keyValueRule[K, V]) ValidateValue(value reflect.Value) ErrorMap {
	errorMap := ErrorMap{}
	for iter := value.MapRange(); iter.Next(); {
		ImplValidateMerge(indirect(iter.Key()), "Key", errorMap, e.keyRules)
		ImplValidateMerge(indirect(iter.Value()), "Value", errorMap, e.valueRules)
	}
	return errorMap.ToNil()
}

func (e keyValueRule[K, V]) TypeCheck(typ reflect.Type) *RuleTypeError {
	return TypeCheck("keyValueRule", typ, reflect.TypeFor[map[K]V](), "Map")
}

var keyValuesValidator = KeyValues[map[int]sliceValidatorElement](newKeyValueRule[int, sliceValidatorElement](
	[]Rule{presentRule{}},
	[]Rule{presentRule{}, Fields[sliceValidatorElement](RuleMap{"Int": {presentRule{}}})},
))

var ptrKeyValuesValidator = KeyValues[map[string]*sliceValidatorElement](newKeyValueRule[string, *sliceValidatorElement](
	nil,
	[]Rule{presentRule{}},
))

func TestKeyValuesAny(t *testing.T) {
	typ := reflect.TypeFor[map[string]Child]()
	expected, err := KeyValuesAnyWithErr(typ, onlyKindRule{kind: reflect.Map})
	require.NoError(t, err)
	require.Equal(t, expected, KeyValuesAny(typ, onlyKindRule{kind: reflect.Map}))
	require.Equal(t, expected, KeyValuesAny(reflect.TypeFor[*map[string]Child](), onlyKindRule{kind: reflect.Map}))

	require.Panics(t, func() { KeyValuesAny(reflect.TypeFor[Child](), onlyKindRule{kind: reflect.Map}) })
}

func TestKeyValuesAnyWithErr(t *testing.T) {
	noMatchingRule := onlyKindRule{kind: reflect.Int}

	tcs := append(mapAnyWithErrTCs("KeyValues"), []anyWithErrTC{
		// the key-value pair rule TypeChecks against the Map type itself, as each key-value pair is passed down as a Map with only 1 key-value pair
		{name: "keyvalue_rule", data: map[string]Child{}, rules: []Rule{newKeyValueRule[string, Child](nil, nil)}},
		{name: "no_matching_rule", data: map[string]Child{}, rules: []Rule{noMatchingRule},
			err: fmt.Errorf("KeyValues: key-value pair type: %w", noMatchingRule.TypeCheck(reflect.TypeFor[map[string]Child]()))},
	}...)
	testAnyWithErr(t, tcs, func(typ reflect.Type, rules ...Rule) (reflect.Type, []Rule, error) {
		validator, err := KeyValuesAnyWithErr(typ, rules...)
		return validator.Type(), validator.KeyValueRules(), err
	})
}

func TestKeyValuesVldr_Validate(t *testing.T) {
	errorKey := ErrorKey("map[int]firm.sliceValidatorElement.[0].Key." + presentRuleKey)
	tcs := []validateTC[map[int]sliceValidatorElement]{
		{name: "valid", data: map[int]sliceValidatorElement{1: {Int: 1}}},
		{name: "invalid", data: map[int]sliceValidatorElement{0: {Int: 1}}, result: ErrorMap{errorKey: *presentRuleError(errorKey)}},
	}
	testValidate(t, tcs, func() (ValidatorTyped[map[int]sliceValidatorElement], error) {
		return KeyValuesWithErr[map[int]sliceValidatorElement](newKeyValueRule[int, sliceValidatorElement]([]Rule{presentRule{}}, nil))
	})
}

func TestKeyValuesAnyVldr_ValidateAll(t *testing.T) {
	validator := keyValuesValidator
	testMapValidateAllTypes(t, validator, (*map[int]sliceValidatorElement)(nil))

	testMapValidateAllCases(t, []mapValidatorTestCase{
		{name: "Full", errorKeys: nil, f: func() any {
			return map[int]sliceValidatorElement{1: {Int: 1, UInt: 1}, 2: {Int: 2, UInt: 2}}
		}},
		{name: "Empty", errorKeys: nil, f: func() any {
			return map[int]sliceValidatorElement{}
		}},
		{name: "Nil", errorKeys: nil, f: func() any {
			return map[int]sliceValidatorElement(nil)
		}},
		{name: "Key_Invalid", errorKeys: []string{"[0].Key"}, f: func() any {
			return map[int]sliceValidatorElement{0: {Int: 1}}
		}},
		{name: "Value_Invalid", errorKeys: []string{"[1].Value.Int"}, f: func() any {
			return map[int]sliceValidatorElement{1: {UInt: 1}}
		}},
		{name: "KeyValue_Invalid_mixed", errorKeys: []string{"[0].Key", "[0].Value.Int"}, f: func() any {
			return map[int]sliceValidatorElement{0: {UInt: 1}, 1: {Int: 1}}
		}},

		//
		// Pointer values
		//
		{name: "Ptr_Value_valid", validator: ptrKeyValuesValidator, errorKeys: nil, f: func() any {
			return map[string]*sliceValidatorElement{"a": {Int: 1}}
		}},
		{name: "Ptr_Value_nil", validator: ptrKeyValuesValidator, invalidKeys: []string{"[a].Value"}, f: func() any {
			return map[string]*sliceValidatorElement{"a": nil}
		}},
	}, validator)
}

func TestKeyValuesAnyVldr_TypeCheck(t *testing.T) {
	validator := keyValuesValidator
	badCondition := "is not matching Map of type map[int]firm.sliceValidatorElement"

	testMapTypeCheck(t, validator, validator.typ, "KeyValuesAnyVldr", badCondition)
}

//
// map test helpers
//

type mapValidatorTestCase struct {
	name string
	f    func() any
	// defaults to the test's default validator
	validator Validator
	errorKeys []string
	// invalidKeys keys Invalid errors for invalid values (e.g. nil pointer keys/values)
	invalidKeys []string
}

type anyWithErrTC struct {
	name  string
	data  any
	rules []Rule
	err   error
}

func mapAnyWithErrTCs(prefix string) []anyWithErrTC {
	mapPt := &map[string]Child{}
	return []anyWithErrTC{
		{name: "normal", data: map[string]Child{}, rules: []Rule{presentRule{}}},
		{name: "map_pointer", data: mapPt, rules: []Rule{presentRule{}}},
		{name: "double_pointer", data: &mapPt, rules: []Rule{presentRule{}}},
		{name: "nil_type", data: nil, err: errors.New(prefix + ": type, nil, is not a Map")},
		{name: "not_map", data: Child{}, err: fmt.Errorf("%s: type, firm.Child, is not a Map", prefix)},
		{name: "slice", data: []Child{}, err: fmt.Errorf("%s: type, []firm.Child, is not a Map", prefix)},
		{name: "pointer_to_not_map", data: &Child{}, err: fmt.Errorf("%s: type, firm.Child, is not a Map", prefix)},
	}
}

func testAnyWithErr(t *testing.T, tcs []anyWithErrTC, newValidator func(typ reflect.Type, rules ...Rule) (reflect.Type, []Rule, error)) {
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			validatorTyp, rules, err := newValidator(reflect.TypeOf(tc.data), tc.rules...)
			if tc.err != nil {
				require.Equal(tc.err, err)
				return
			}

			require.NoError(err)
			require.Equal(indirectType(reflect.TypeOf(tc.data)), validatorTyp)
			require.Equal(tc.rules, rules)
		})
	}
}

// testMapValidateAllTypes asserts the error results of ValidateAny for non-map, nil and nil-pointer data
func testMapValidateAllTypes(t *testing.T, validator Validator, nilPtrData any) {
	tcs := []struct {
		name   string
		data   any
		result ErrorMap
	}{
		{name: "not_map", data: 1, result: typeCheckErrorResult(validator, 1)},
		{name: "invalid", data: nil, result: ErrInvalidValue()},
		{name: "nil_pointer", data: nilPtrData, result: ErrInvalidValue()},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) { require.Equal(t, tc.result, validator.ValidateAny(tc.data)) })
	}
}

// testMapValidateAllCases asserts ValidateAny/ValidateValue/ValidateMerge for each case's data, both raw and pointer-boxed
func testMapValidateAllCases(t *testing.T, cases []mapValidatorTestCase, defaultValidator Validator) {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			validator := tc.validator
			if validator == nil {
				validator = defaultValidator
			}
			rawData := tc.f()
			// rawData comes boxed in an any, so &rawData would be a *any; build a typed pointer instead
			ptrData := reflect.New(reflect.TypeOf(rawData))
			ptrData.Elem().Set(reflect.ValueOf(rawData))
			testValidateAllKeys(t, validator, rawData,
				joinAll(tc.errorKeys, presentRuleKey), joinAll(tc.invalidKeys, invalidKey))
			testValidateAllKeys(t, validator, ptrData.Interface(),
				joinAll(tc.errorKeys, presentRuleKey), joinAll(tc.invalidKeys, invalidKey))
		})
	}
}

func testMapTypeCheck(t *testing.T, validator Validator, typ reflect.Type, ruleName, badCondition string) {
	tcs := []struct {
		name         string
		data         any
		badCondition string
	}{
		{name: "matching map", data: reflect.Zero(typ).Interface()},
		{name: "matching map pointer", data: reflect.New(typ).Interface()},
		{name: "other map", data: map[string]int{}, badCondition: badCondition},
		{name: "not map", data: 1, badCondition: badCondition},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			testTypeCheck(t, tc.data, ruleName, tc.badCondition, func() (Rule, error) {
				return validator, nil
			})
		})
	}
}
