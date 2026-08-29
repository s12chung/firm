package firm

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type registryParent struct {
	Primitive int
	Child     registryChild
}

type registryChild struct{}

type registryNotFoundTest struct{}

func TestRegistry_RegisterType(t *testing.T) {
	require := require.New(t)

	registry := &Registry{}
	require.NoError(registry.RegisterType(NewDefinition[registryParent]().
		Validates(RuleMap{"Child": {registry.Backed()}}).
		ValidatesSelf(presentRule{})))

	registryParentType := reflect.TypeFor[registryParent]()
	registryChildType := reflect.TypeFor[registryChild]()
	typeToValidator := map[reflect.Type]*ValueAnyVldr{
		registryParentType: {
			typ: registryParentType,
			rules: []Rule{presentRule{},
				FieldsAnyVldr{typ: registryParentType, ruleMap: map[string]*[]Rule{
					// stamped at registration with the field's type
					"Child": {RegistryBacker{Registry: registry, typ: registryChildType}},
				}},
			}},
	}
	require.Equal(typeToValidator, registry.typeToValidator)

	require.NoError(registry.RegisterType(NewDefinition[registryChild]()))
	typeToValidator[registryChildType] = &ValueAnyVldr{typ: registryChildType, rules: []Rule{}}
	require.Equal(typeToValidator, registry.typeToValidator)

	require.Equal(errors.New("RegisterType() with type firm.registryParent already exists"),
		registry.RegisterType(NewDefinition[registryParent]().ValidatesSelf(presentRule{})))
}

type registryInt int

func TestRegistry_RegisterType_NonStruct(t *testing.T) {
	require := require.New(t)

	registry := &Registry{}
	require.NoError(registry.RegisterType(NewDefinition[registryInt]().ValidatesSelf(presentRule{})))

	valid := registryInt(1)
	testValidateAll(t, registry, valid, nil)
	testValidateAll(t, registry, &valid, nil)

	invalid := registryInt(0)
	testValidateAll(t, registry, invalid, presentRuleError(""), presentRuleKey)
	testValidateAll(t, registry, &invalid, presentRuleError(""), presentRuleKey)
}

func notFoundError(data any) ErrorMap {
	value := reflect.ValueOf(data)
	errorMap := ErrorMap{}
	DefaultValidator.ValidateMerge(value, typeName(value), errorMap)
	return errorMap.Finish()
}

// nolint:funlen // a bunch of test cases
func TestRegistry_ValidateAll(t *testing.T) {
	type testCase struct {
		name       string
		definition *Definition
		data       func() registryParent

		expectedKeySuffix string
		err               *TemplateError
	}
	tcs := []testCase{
		{
			name:              "self",
			definition:        NewDefinition[registryParent]().ValidatesSelf(presentRule{}),
			data:              func() registryParent { return registryParent{} },
			expectedKeySuffix: presentRuleKey,
			err:               presentRuleError(""),
		},
		{
			name: "field_Primitive",
			definition: NewDefinition[registryParent]().Validates(RuleMap{
				"Primitive": {presentRule{}},
			}),
			data:              func() registryParent { return registryParent{} },
			expectedKeySuffix: "Primitive.presentRule",
			err:               presentRuleError(""),
		},
		{
			name:              "not_found",
			definition:        NewDefinition[registryParent]().ValidatesSelf(presentRule{}),
			expectedKeySuffix: "NotFound",
		},
		{
			name: "not_found_field_Primitive",
			definition: NewDefinition[registryParent]().Validates(RuleMap{
				"Primitive": {presentRule{}},
			}),
			expectedKeySuffix: "NotFound",
		},
		{
			name:       "invalid",
			definition: NewDefinition[registryParent]().ValidatesSelf(presentRule{}),
			data:       nil,
		},
		{
			name:       "typed_nil",
			definition: NewDefinition[registryParent]().ValidatesSelf(presentRule{}),
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			registry := &Registry{}
			require.NoError(registry.RegisterType(tc.definition))

			if tc.name == "invalid" {
				require.Equal(notFoundError(nilType{}), registry.ValidateAny(nil))
				var data any
				require.Equal(notFoundError(&data), registry.ValidateAny(&data))
				return
			}
			if tc.name == "typed_nil" {
				var data *registryParent
				require.Equal(errInvalidValue, registry.ValidateAny(data))
				return
			}
			if strings.HasPrefix(tc.name, "not_found") {
				data := registryNotFoundTest{}
				require.Equal(notFoundError(data), registry.ValidateAny(data))
				require.Equal(notFoundError(&data), registry.ValidateAny(&data))

				notFoundTemplateError := &TemplateError{Template: "type, {{.RootTypeName}}, not found in Registry"}
				testValidateAllFull(t, true, registry, data, notFoundTemplateError, tc.expectedKeySuffix)
				testValidateAllFull(t, true, registry, &data, notFoundTemplateError, tc.expectedKeySuffix)
				return
			}
			data := tc.data()
			testValidateAll(t, registry, data, tc.err, tc.expectedKeySuffix)
			testValidateAll(t, registry, &data, tc.err, tc.expectedKeySuffix)
		})
	}
}

func TestRegistry_DefaultedValidator(t *testing.T) {
	require := require.New(t)

	registry := &Registry{}
	require.NoError(registry.RegisterType(NewDefinition[registryParent]().ValidatesSelf(presentRule{})))
	expected, err := ValueAnyWithErr(reflect.TypeFor[registryParent](), presentRule{})
	require.NoError(err)
	require.Equal(&expected, registry.DefaultedValidator(reflect.TypeFor[registryParent]()))

	notFoundType := reflect.TypeOf(nil)
	require.Equal(DefaultValidator, registry.DefaultedValidator(notFoundType))

	type anyType struct{}
	registry.DefaultValidator = ValueVldr[anyType]{}
	require.Equal(registry.DefaultValidator, registry.DefaultedValidator(notFoundType))
}

func TestRegistry_Validator(t *testing.T) {
	registry := &Registry{}
	require.NoError(t, registry.RegisterType(NewDefinition[registryParent]().ValidatesSelf(presentRule{})))
	testParentValidator, err := ValueAnyWithErr(reflect.TypeFor[registryParent](), presentRule{})
	require.NoError(t, err)

	tcs := []struct {
		name     string
		data     any
		expected Validator
	}{
		{name: "normal", data: registryParent{}, expected: &testParentValidator},
		{name: "pointer", data: &registryParent{}, expected: &testParentValidator},
		{name: "not_found", data: registryNotFoundTest{}},
		{name: "not_found_pointer", data: &registryNotFoundTest{}},
		{name: "nil_validator", data: Validator(nil)},
		{name: "pure_nil", data: nil},
		{name: "zero", data: 0},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expected, registry.Validator(reflect.TypeOf(tc.data)))
		})
	}
}

type multiPtrParent struct {
	Ptr  ***Child
	Ptrs *[]**Child
}

// nolint:funlen // a bunch of test cases
func TestMultiPtr_ValidateAll(t *testing.T) {
	registry := &Registry{}
	require.NoError(t, registry.RegisterType(NewDefinition[multiPtrParent]().Validates(RuleMap{
		"Ptr":  {presentRule{}, registry.Backed()},
		"Ptrs": {presentRule{}, Elems[[]**Child](registry.Backed())},
	})))
	require.NoError(t, registry.RegisterType(NewDefinition[Child]().Validates(RuleMap{
		"Validates": {presentRule{}},
	})))

	// valid
	good := Child{Validates: "ok"}
	pGood := &good
	ppGood := &pGood

	// empty
	empty := Child{}
	pEmpty := &empty
	ppEmpty := &pEmpty

	var nilChild *Child
	pNil := &nilChild
	ppNil := &pNil

	tcs := []struct {
		name        string
		data        any
		err         *TemplateError
		keySuffixes []string
	}{
		{name: "Child___valid", data: &ppGood},
		{name: "Child___empty", data: &ppEmpty, err: presentRuleError(""),
			keySuffixes: []string{joinKeys("Validates", presentRuleKey)}},

		{name: "Parent___valid", data: multiPtrParent{Ptr: &ppGood, Ptrs: &[]**Child{ppGood}}},
		{name: "Parent___empty", data: multiPtrParent{},
			err: presentRuleError(""), keySuffixes: []string{
				joinKeys("Ptr", presentRuleKey),
				joinKeys("Ptrs", presentRuleKey),
			}},
		{name: "Parent___Ptr_empty", data: multiPtrParent{Ptr: &ppEmpty, Ptrs: &[]**Child{ppGood}},
			err: presentRuleError(""), keySuffixes: []string{
				joinKeys("Ptr", presentRuleKey),
				joinKeys("Ptr.Validates", presentRuleKey),
			}},
		{name: "Parent__Ptr_nil", data: multiPtrParent{Ptr: ppNil, Ptrs: &[]**Child{ppGood}},
			err: presentRuleError(""), keySuffixes: []string{
				joinKeys("Ptr", presentRuleKey),
			}},
		{name: "Parent___Ptrs_mixed", data: multiPtrParent{Ptr: &ppGood, Ptrs: &[]**Child{nil, ppEmpty}},
			// nil elements are indirected into invalid values and silently skipped by the RegistryBacker,
			// so only the empty child surfaces errors
			err: presentRuleError(""), keySuffixes: []string{
				joinKeys("Ptrs.[1].Validates", presentRuleKey),
			}},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			testValidateAll(t, registry, tc.data, tc.err, tc.keySuffixes...)
		})
	}
}

func TestRegistryBacker(t *testing.T) {
	require := require.New(t)

	registry := &Registry{}
	require.NoError(RegisterType(NewDefinition[registryChild]()))

	// unstamped backers have no type, so they fall back to the DefaultValidator's NotFound error
	require.Equal(notFoundError(registryParent{}), registry.Backed().ValidateAny(registryParent{}))
}

func TestRegistryBacker_SelfRecursion(t *testing.T) {
	require := require.New(t)

	registry := &Registry{}
	selfRecursionErr := "RegisterType() with type firm.registryChild: " +
		"Registry: self recursion with RegistryBacker.Registry pointing to RegisterType()'s Registry"

	// recursive, so reject
	require.EqualError(registry.RegisterType(NewDefinition[registryChild]().ValidatesSelf(registry.Backed())), selfRecursionErr)

	// a different Registry terminates, so allow
	otherRegistry := &Registry{}
	require.NoError(registry.RegisterType(NewDefinition[registryChild]().ValidatesSelf(otherRegistry.Backed())))
}

// nolint:funlen // a bunch of test cases
func TestRegistryBacker_Stamping(t *testing.T) {
	registry := &Registry{}
	require.NoError(t, registry.RegisterType(NewDefinition[registryChild]().ValidatesSelf(presentRule{})))
	// KeyValuesAnyWithErr stamps entries as an anonymous KeyValue struct, so it is
	// registered by the anonymous type
	require.NoError(t, registry.RegisterType(NewDefinition[struct {
		Key   registryChild
		Value registryChild
	}]().ValidatesSelf(presentRule{})))

	childBacker := []Rule{RegistryBacker{Registry: registry, typ: reflect.TypeFor[registryChild]()}}
	keyValueBacker := []Rule{RegistryBacker{Registry: registry, typ: reflect.StructOf([]reflect.StructField{
		{Name: "Key", Type: reflect.TypeFor[registryChild]()},
		{Name: "Value", Type: reflect.TypeFor[registryChild]()},
	})}}

	tcs := []struct {
		name string
		// the Validator and the rules holding the backer, per constructor
		newValidator  func() (Validator, []Rule, error)
		expectedRules []Rule
		data          any
		keySuffixes   []string
	}{
		{name: "fields", data: registryParent{}, keySuffixes: []string{joinKeys("Child", presentRuleKey)},
			expectedRules: childBacker,
			newValidator: func() (Validator, []Rule, error) {
				v, err := FieldsAnyWithErr(reflect.TypeFor[registryParent](), RuleMap{"Child": {registry.Backed()}})
				return v, *v.ruleMap["Child"], err
			}},
		{name: "fields_ptr", data: registryParent{}, keySuffixes: []string{joinKeys("Child", presentRuleKey)},
			expectedRules: childBacker,
			newValidator: func() (Validator, []Rule, error) {
				v, err := FieldsAnyWithErr(reflect.TypeFor[*registryParent](), RuleMap{"Child": {registry.Backed()}})
				return v, *v.ruleMap["Child"], err
			}},
		{name: "elems", data: []registryChild{{}}, keySuffixes: []string{joinKeys("[0]", presentRuleKey)},
			expectedRules: childBacker,
			newValidator: func() (Validator, []Rule, error) {
				v, err := ElemsAnyWithErr(reflect.TypeFor[[]registryChild](), registry.Backed())
				return v, v.elementRules, err
			}},
		{name: "elems_ptr", data: []registryChild{{}}, keySuffixes: []string{joinKeys("[0]", presentRuleKey)},
			expectedRules: childBacker,
			newValidator: func() (Validator, []Rule, error) {
				v, err := ElemsAnyWithErr(reflect.TypeFor[*[]registryChild](), registry.Backed())
				return v, v.elementRules, err
			}},
		{name: "value", data: registryChild{}, keySuffixes: []string{presentRuleKey},
			expectedRules: childBacker,
			newValidator: func() (Validator, []Rule, error) {
				v, err := ValueAnyWithErr(reflect.TypeFor[registryChild](), registry.Backed())
				return v, v.rules, err
			}},
		{name: "value_ptr", data: registryChild{}, keySuffixes: []string{presentRuleKey},
			expectedRules: childBacker,
			newValidator: func() (Validator, []Rule, error) {
				v, err := ValueAnyWithErr(reflect.TypeFor[*registryChild](), registry.Backed())
				return v, v.rules, err
			}},
		{name: "keys", data: map[registryChild]registryChild{{}: {}}, keySuffixes: []string{joinKeys("[{}]", presentRuleKey)},
			expectedRules: childBacker,
			newValidator: func() (Validator, []Rule, error) {
				v, err := KeysAnyWithErr(reflect.TypeFor[map[registryChild]registryChild](), registry.Backed())
				return v, v.keyRules, err
			}},
		{name: "keys_ptr", data: map[registryChild]registryChild{{}: {}}, keySuffixes: []string{joinKeys("[{}]", presentRuleKey)},
			expectedRules: childBacker,
			newValidator: func() (Validator, []Rule, error) {
				v, err := KeysAnyWithErr(reflect.TypeFor[*map[registryChild]registryChild](), registry.Backed())
				return v, v.keyRules, err
			}},
		{name: "values", data: map[registryChild]registryChild{{}: {}}, keySuffixes: []string{joinKeys("[{}]", presentRuleKey)},
			expectedRules: childBacker,
			newValidator: func() (Validator, []Rule, error) {
				v, err := ValuesAnyWithErr(reflect.TypeFor[map[registryChild]registryChild](), registry.Backed())
				return v, v.valueRules, err
			}},
		{name: "values_ptr", data: map[registryChild]registryChild{{}: {}}, keySuffixes: []string{joinKeys("[{}]", presentRuleKey)},
			expectedRules: childBacker,
			newValidator: func() (Validator, []Rule, error) {
				v, err := ValuesAnyWithErr(reflect.TypeFor[*map[registryChild]registryChild](), registry.Backed())
				return v, v.valueRules, err
			}},
		{name: "key_values", data: map[registryChild]registryChild{{}: {}}, keySuffixes: []string{joinKeys("[{}]", presentRuleKey)},
			expectedRules: keyValueBacker,
			newValidator: func() (Validator, []Rule, error) {
				v, err := KeyValuesAnyWithErr(reflect.TypeFor[map[registryChild]registryChild](), registry.Backed())
				return v, v.entryRules, err
			}},
		{name: "key_values_ptr", data: map[registryChild]registryChild{{}: {}}, keySuffixes: []string{joinKeys("[{}]", presentRuleKey)},
			expectedRules: keyValueBacker,
			newValidator: func() (Validator, []Rule, error) {
				v, err := KeyValuesAnyWithErr(reflect.TypeFor[*map[registryChild]registryChild](), registry.Backed())
				return v, v.entryRules, err
			}},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			validator, rules, err := tc.newValidator()
			require.NoError(err)
			require.Equal(tc.expectedRules, rules)

			// the stamped backer resolves through the Registry at validation time
			testValidateAll(t, validator, tc.data, presentRuleError(""), tc.keySuffixes...)
		})
	}
}
