package firm

import (
	"maps"
	"reflect"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

const nilName = "nil"
const presentRuleKey = "presentRule"
const invalidKey = "Invalid"

func typeName(value reflect.Value) string {
	if !value.IsValid() {
		return nilName
	}
	return indirect(value).Type().String()
}

type onlyKindRule struct{ kind reflect.Kind }

func (o onlyKindRule) ValidateValue(_ reflect.Value) ErrorMap { return nil }
func (o onlyKindRule) TypeCheck(typ reflect.Type) *RuleTypeError {
	if typ.Kind() != o.kind {
		return NewRuleTypeError("onlyKindRule", typ, "is not "+o.kind.String())
	}
	return nil
}

type mutantRule struct{}

func (mutantRule) ValidateValue(_ reflect.Value) ErrorMap {
	return ErrorMap{"mutantRule": TemplateError{Template: "mutant rule ran"}}
}
func (mutantRule) TypeCheck(_ reflect.Type) *RuleTypeError { return nil }

// testRulesGetterIsolation asserts mutating a getter's returned rules, via index assignment or
// reslice-append (both write through shared backing arrays), does not affect the validator's rules
func testRulesGetterIsolation(t *testing.T, getter func() []Rule) {
	t.Helper()
	require := require.New(t)
	expected := slices.Clone(getter())
	require.NotEmpty(expected)

	rules := getter()
	rules[0] = mutantRule{}
	_ = append(rules[:0], mutantRule{})

	require.Equal(expected, getter())
}

type presentRule struct{}

func (p presentRule) ValidateValue(value reflect.Value) ErrorMap {
	if !value.IsValid() || value.IsZero() {
		return ErrorMap{presentRuleKey: *presentRuleError("")}
	}
	return nil
}
func (p presentRule) TypeCheck(_ reflect.Type) *RuleTypeError { return nil }

func errOnNilValidator(validator Validator) Validator {
	switch validator := validator.(type) {
	case ElemsVldr[[]*sliceValidatorElement, *sliceValidatorElement]:
		return validator.ErrOnNil()
	case ElemsAnyVldr:
		return validator.ErrOnNil()
	case KeysVldr[map[*int]sliceValidatorElement, *int, sliceValidatorElement]:
		return validator.ErrOnNil()
	case ValuesVldr[map[string]*sliceValidatorElement, string, *sliceValidatorElement]:
		return validator.ErrOnNil()
	case ValuesAnyVldr:
		return validator.ErrOnNil()
	default:
		return nil
	}
}

// errOnNilSelfer is a Validator whose ErrOnNilSelf() returns its own type--setters return copies
type errOnNilSelfer[V Validator] interface {
	Validator
	ErrOnNilSelf() V
}

// testErrOnNilSelf asserts nil values are skipped by default, and error with the ErrOnNilSelf()
// variant--whose setter returns a copy, leaving the validator unchanged
func testErrOnNilSelf(t *testing.T, validator any) {
	require := require.New(t)

	v, ok := validator.(Validator)
	require.True(ok)
	require.Nil(v.ValidateAny(nil))

	// ErrOnNilSelf() returns the validator's own type--setters return copies
	method := reflect.ValueOf(v).MethodByName("ErrOnNilSelf")
	require.True(method.IsValid())
	errOnNilSelfValidator, ok := reflect.TypeAssert[Validator](method.Call(nil)[0])
	require.True(ok)
	require.Equal(ErrInvalidValue(), errOnNilSelfValidator.ValidateAny(nil))

	// the setter returns a copy--the unset validator still skips
	require.Nil(v.ValidateAny(nil))
}

//nolint:unparam // leave it for tests
func presentRuleError(errorKey ErrorKey) *TemplateError {
	return &TemplateError{ErrorKey: errorKey, Template: presentRuleKey + " template"}
}

// invalidError returns the Invalid error the validators use for invalid values
func invalidError(errorKey ErrorKey) *TemplateError {
	templateError := ErrInvalidValue()["Invalid"]
	templateError.ErrorKey = errorKey
	return &templateError
}

func typeCheckErrorResult(rule Rule, data any) ErrorMap {
	return ErrorMap{"TypeCheck": rule.TypeCheck(reflect.TypeOf(data)).TemplateError()}
}

// joinAll joins each key with the suffix
func joinAll(keys []string, suffix string) []string {
	joined := make([]string, len(keys))
	for i, key := range keys {
		joined[i] = joinKeys(key, suffix)
	}
	return joined
}

// testValidateAll asserts ValidateAny/ValidateValue/ValidateMerge, expecting err keyed at every keySuffix
func testValidateAll(t *testing.T, validator Validator, data any, err *TemplateError, keySuffixes ...string) {
	testValidateAllExpected(t, false, validator, data, suffixErrorMap(err, keySuffixes))
}

// testValidateAllKeys asserts with presentRule errors at keySuffixes and Invalid errors at invalidKeySuffixes
func testValidateAllKeys(t *testing.T, validator Validator, data any, keySuffixes, invalidKeySuffixes []string) {
	expected := suffixErrorMap(presentRuleError(""), keySuffixes)
	for key, err := range suffixErrorMap(invalidError(""), invalidKeySuffixes) {
		if expected == nil {
			expected = ErrorMap{}
		}
		expected[key] = err
	}
	testValidateAllExpected(t, false, validator, data, expected)
}

func testValidateAllFull(t *testing.T, skipValidate bool, validator Validator, data any, err *TemplateError, keySuffixes ...string) {
	testValidateAllExpected(t, skipValidate, validator, data, suffixErrorMap(err, keySuffixes))
}

func suffixErrorMap(err *TemplateError, keySuffixes []string) ErrorMap {
	var expected ErrorMap
	if err != nil && len(keySuffixes) > 0 {
		expected = ErrorMap{}
		for _, key := range keySuffixes {
			keyedErr := *err
			keyedErr.ErrorKey = ErrorKey(key)
			expected[ErrorKey(key)] = keyedErr
		}
	}
	return expected
}

func testValidateAllExpected(t *testing.T, skipValidate bool, validator Validator, data any, validateValueExpected ErrorMap) {
	require := require.New(t)

	validateExpected := ErrorMap{}
	validateValueExpected.MergeInto(typeName(reflect.ValueOf(data)), validateExpected)
	validateExpected = validateExpected.ToNil()

	if !skipValidate {
		require.Equal(validateExpected, validator.ValidateAny(data))
	}
	indirectValue := indirect(reflect.ValueOf(data))
	require.Equal(validateValueExpected, validator.ValidateValue(indirectValue))

	errorKey := "pkger.Mover.Parent"
	errorMap := ErrorMap{"Existing": TemplateError{}}
	expectedErrorMap := maps.Clone(errorMap)
	for keySuffix, err := range validateValueExpected {
		key := ErrorKey(joinKeys(errorKey, string(keySuffix)))
		mergedErr := err
		mergedErr.ErrorKey = key
		expectedErrorMap[key] = mergedErr
	}
	indirectValue = indirect(reflect.ValueOf(data))
	validator.ValidateMerge(indirectValue, errorKey, errorMap)
	require.Equal(expectedErrorMap, errorMap)
}

type validateTC[T any] struct {
	name   string
	data   T
	result ErrorMap
}

func testValidate[T any](t *testing.T, tcs []validateTC[T], newValidator func() (ValidatorTyped[T], error)) {
	validator, err := newValidator()
	require.NoError(t, err)

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.result, validator.Validate(tc.data))
			require.Equal(tc.result, validator.ValidateAny(tc.data))
		})
	}
}

func testTypeCheck(t *testing.T, data any, ruleName, badCondition string, newValidator func() (Rule, error)) {
	require := require.New(t)

	validator, err := newValidator()
	require.NoError(err)

	typ := reflect.TypeOf(data)

	var ruleTypeError *RuleTypeError
	if badCondition != "" {
		ruleTypeError = NewRuleTypeError(ruleName, typ, badCondition)
	}
	require.Equal(ruleTypeError, validator.TypeCheck(typ))
}
