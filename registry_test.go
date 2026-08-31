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

type cycleParent struct {
	Child *cycleChild
}

type cycleChild struct {
	Parent *cycleParent
}

type cycleSelfRef struct {
	Child *cycleSelfRef
}

type cycleSliceParent struct {
	Children []cycleSliceChild
}

type cycleSliceChild struct {
	Parent *cycleSliceParent
}

// cycleCustomVldr wraps a Rule like RuleVldr, exposing its rules via AllRules()
type cycleCustomVldr struct{ Rule }

func (c cycleCustomVldr) ValidateAny(data any) ErrorMap { return ImplValidateAny(c, data) }

func (c cycleCustomVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	ImplValidateMerge(value, key, errorMap, []Rule{c.Rule})
}

func (c cycleCustomVldr) AllRules() []Rule { return []Rule{c.Rule} }

// cycleOpaqueVldr is like cycleCustomVldr, but without AllRules()
type cycleOpaqueVldr struct{ Rule }

func (c cycleOpaqueVldr) ValidateAny(data any) ErrorMap { return ImplValidateAny(c, data) }

func (c cycleOpaqueVldr) ValidateMerge(value reflect.Value, key string, errorMap ErrorMap) {
	ImplValidateMerge(value, key, errorMap, []Rule{c.Rule})
}

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
				FieldsAnyVldr{typ: registryParentType, ruleMap: map[string][]Rule{
					// stamped at registration with the field's type
					"Child": {RegistryBacker{Registry: registry, typ: registryChildType}},
				}, fieldIndexes: map[string][]int{"Child": {1}}},
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
	value := indirect(reflect.ValueOf(data))
	errorMap := ErrorMap{}
	key := typeName(value)
	if value.Type() == nilValueType {
		key = "" // nilType stand-ins are not named
	}
	DefaultValidator.ValidateMerge(value, key, errorMap)
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
				require.Equal(ErrInvalidValue(), registry.ValidateAny(data))
				return
			}
			if strings.HasPrefix(tc.name, "not_found") {
				data := registryNotFoundTest{}
				require.Equal(notFoundError(data), registry.ValidateAny(data))
				require.Equal(notFoundError(&data), registry.ValidateAny(&data))

				notFoundTemplateError := &TemplateError{
					Template:       "type, {{.ValueTypeName}}, not found in Registry",
					TemplateFields: map[string]string{"ValueTypeName": typeName(reflect.ValueOf(data))},
				}
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
		keySuffixes []string
		invalidKeys []string
	}{
		{name: "Child___valid", data: &ppGood},
		{name: "Child___empty", data: &ppEmpty,
			keySuffixes: []string{"Validates"}},

		{name: "Parent___valid", data: multiPtrParent{Ptr: &ppGood, Ptrs: &[]**Child{ppGood}}},
		{name: "Parent___empty", data: multiPtrParent{},
			invalidKeys: []string{"Ptr", "Ptrs"}},
		{name: "Parent___Ptr_empty", data: multiPtrParent{Ptr: &ppEmpty, Ptrs: &[]**Child{ppGood}},
			keySuffixes: []string{"Ptr", "Ptr.Validates"}},
		{name: "Parent__Ptr_nil", data: multiPtrParent{Ptr: ppNil, Ptrs: &[]**Child{ppGood}},
			// the nil Child behind the pointers surfaces an Invalid error
			invalidKeys: []string{"Ptr"}},
		{name: "Parent___Ptrs_mixed", data: multiPtrParent{Ptr: &ppGood, Ptrs: &[]**Child{nil, ppEmpty}},
			// the nil element surfaces an Invalid error, while the empty child surfaces errors
			keySuffixes: []string{"Ptrs.[1].Validates"},
			invalidKeys: []string{"Ptrs.[0]"}},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			testValidateAllKeys(t, registry, tc.data, joinAll(tc.keySuffixes, presentRuleKey), joinAll(tc.invalidKeys, invalidKey))
		})
	}
}

func TestRegistryBacker(t *testing.T) {
	require := require.New(t)

	registry := &Registry{}
	require.NoError(registry.RegisterType(NewDefinition[registryParent]().ValidatesSelf(presentRule{})))

	// unstamped backers have no type, so the data's own type routes the validation
	testValidateAll(t, registry.Backed(), registryParent{}, presentRuleError(""), presentRuleKey)
	// wrapping in RuleVldr leaves the backer unstamped
	testValidateAll(t, RuleVldr{Rule: registry.Backed()}, registryParent{}, presentRuleError(""), presentRuleKey)

	require.Equal(notFoundError(registryNotFoundTest{}), registry.Backed().ValidateAny(registryNotFoundTest{}))
	require.Equal(notFoundError(nilType{}), registry.Backed().ValidateAny(nil))

	// stamped backers route by their stamped type
	parentType := reflect.TypeFor[registryParent]()
	stamped := RegistryBacker{Registry: registry, typ: parentType}
	testValidateAll(t, stamped, registryParent{}, presentRuleError(""), presentRuleKey)
	// registryNotFoundTest is unregistered, so the stamped type routes, not the data's own type
	require.Equal(ErrorMap{"TypeCheck": NewRuleTypeError(
		"ValueAnyVldr", reflect.TypeFor[registryNotFoundTest](), "is not matching of type "+parentType.String(),
	).TemplateError()}, stamped.ValidateAny(registryNotFoundTest{}))
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

func TestRegistry_RegisterType_Recursion(t *testing.T) {
	cycleError := func(typ string) string {
		return "RegisterType() with type firm." + typ + ": " +
			"Registry: type, firm." + typ + ", recurses back to itself via RegistryBacker"
	}

	t.Run("fields", func(t *testing.T) {
		require := require.New(t)
		registry := &Registry{}
		// the edge to the not-yet-registered cycleParent terminates the walk, so allow
		require.NoError(registry.RegisterType(NewDefinition[cycleChild]().
			Validates(RuleMap{"Parent": {registry.Backed()}})))
		err := registry.RegisterType(NewDefinition[cycleParent]().
			Validates(RuleMap{"Child": {registry.Backed()}}))
		require.EqualError(err, cycleError("cycleParent"))
	})

	t.Run("self_reference", func(t *testing.T) {
		require := require.New(t)
		registry := &Registry{}
		err := registry.RegisterType(NewDefinition[cycleSelfRef]().
			Validates(RuleMap{"Child": {registry.Backed()}}))
		require.EqualError(err, cycleError("cycleSelfRef"))
	})

	t.Run("elems", func(t *testing.T) {
		require := require.New(t)
		registry := &Registry{}
		require.NoError(registry.RegisterType(NewDefinition[cycleSliceChild]().
			Validates(RuleMap{"Parent": {registry.Backed()}})))
		err := registry.RegisterType(NewDefinition[cycleSliceParent]().
			Validates(RuleMap{"Children": {Elems[[]cycleSliceChild](registry.Backed())}}))
		require.EqualError(err, cycleError("cycleSliceParent"))
	})

	t.Run("custom_validator", func(t *testing.T) {
		require := require.New(t)
		registry := &Registry{}
		require.NoError(registry.RegisterType(NewDefinition[cycleChild]().
			Validates(RuleMap{"Parent": {registry.Backed()}})))
		// AllRules() exposes the wrapped Fields validator's stamped backer to the cycle check
		err := registry.RegisterType(NewDefinition[cycleParent]().
			Validates(RuleMap{"Child": {cycleCustomVldr{Rule: Fields[cycleChild](RuleMap{"Parent": {registry.Backed()}})}}}))
		require.EqualError(err, cycleError("cycleParent"))
	})

	t.Run("custom_validator_without_all_rules", func(t *testing.T) {
		require := require.New(t)
		registry := &Registry{}
		require.NoError(registry.RegisterType(NewDefinition[cycleChild]().
			Validates(RuleMap{"Parent": {registry.Backed()}})))
		// without AllRules(), the wrapped validator is opaque to the cycle check
		require.NoError(registry.RegisterType(NewDefinition[cycleParent]().
			Validates(RuleMap{"Child": {cycleOpaqueVldr{Rule: Fields[cycleChild](RuleMap{"Parent": {registry.Backed()}})}}})))
	})
}

// nolint:funlen // a bunch of test cases
func TestRegistryBacker_Stamping(t *testing.T) {
	registry := &Registry{}
	require.NoError(t, registry.RegisterType(NewDefinition[registryChild]().ValidatesSelf(presentRule{})))
	// KeyValuesAnyWithErr stamps key-value pair rules as the Map type itself, so it is registered by the Map type.
	// Keys() is used, as a key-value pair is a non-nil Map with only 1 key-value pair, so presentRule would always pass on "itself"
	require.NoError(t, registry.RegisterType(NewDefinition[map[registryChild]registryChild]().
		ValidatesSelf(Keys[map[registryChild]registryChild](presentRule{}))))

	childBacker := []Rule{RegistryBacker{Registry: registry, typ: reflect.TypeFor[registryChild]()}}
	mapBacker := []Rule{RegistryBacker{Registry: registry, typ: reflect.TypeFor[map[registryChild]registryChild]()}}

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
				return v, v.ruleMap["Child"], err
			}},
		{name: "fields_ptr", data: registryParent{}, keySuffixes: []string{joinKeys("Child", presentRuleKey)},
			expectedRules: childBacker,
			newValidator: func() (Validator, []Rule, error) {
				v, err := FieldsAnyWithErr(reflect.TypeFor[*registryParent](), RuleMap{"Child": {registry.Backed()}})
				return v, v.ruleMap["Child"], err
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
		{name: "key_values", data: map[registryChild]registryChild{{}: {}}, keySuffixes: []string{joinKeys("[{}].[{}]", presentRuleKey)},
			expectedRules: mapBacker,
			newValidator: func() (Validator, []Rule, error) {
				v, err := KeyValuesAnyWithErr(reflect.TypeFor[map[registryChild]registryChild](), registry.Backed())
				return v, v.keyValueRules, err
			}},
		{name: "key_values_ptr", data: map[registryChild]registryChild{{}: {}}, keySuffixes: []string{joinKeys("[{}].[{}]", presentRuleKey)},
			expectedRules: mapBacker,
			newValidator: func() (Validator, []Rule, error) {
				v, err := KeyValuesAnyWithErr(reflect.TypeFor[*map[registryChild]registryChild](), registry.Backed())
				return v, v.keyValueRules, err
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
