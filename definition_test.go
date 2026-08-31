package firm

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDefinition(t *testing.T) {
	require := require.New(t)

	require.NotPanics(func() { NewDefinition[Child]() })
	require.NotPanics(func() { NewDefinition[int]() })    // non-struct
	require.NotPanics(func() { NewDefinition[*Child]() }) // pointer, indirected
	require.Equal(reflect.TypeFor[Child](), NewDefinition[**Child]().Type())
	require.Panics(func() { NewDefinition[any]() }) // interface, no type
}

func TestDefinition_ValidatesSelf(t *testing.T) {
	require := require.New(t)
	rules := []Rule{presentRule{}}

	definition := NewDefinition[Child]().ValidatesSelf(rules...)
	require.Equal(rules, definition.SelfRules())
	require.Panics(func() {
		definition.ValidatesSelf()
	})
	// an empty call still counts as called
	require.Panics(func() { NewDefinition[Child]().ValidatesSelf().ValidatesSelf() })
}

func TestDefinition_Validates(t *testing.T) {
	require := require.New(t)

	ruleMap := RuleMap{"Validates": {presentRule{}}}
	definition := NewDefinition[Child]().Validates(ruleMap)
	require.Equal(ruleMap, definition.RuleMap())

	require.Panics(func() { definition.Validates(RuleMap{}) })
	// an empty RuleMap still counts as called
	require.Panics(func() { NewDefinition[Child]().Validates(RuleMap{}).Validates(RuleMap{}) })

	// field checks error at RegisterType() via FieldsAnyWithErr()
	registry := &Registry{}
	err := registry.RegisterType(NewDefinition[Child]().Validates(RuleMap{"DoesNotExist": {}}))
	require.EqualError(err, "RegisterType() with type firm.Child: Fields: field, DoesNotExist, not found in type: firm.Child")
	err = registry.RegisterType(NewDefinition[Child]().Validates(RuleMap{"private": {}}))
	require.EqualError(err, "RegisterType() with type firm.Child: Fields: field, private, is unexported in type: firm.Child")
	err = registry.RegisterType(NewDefinition[int]().Validates(RuleMap{"Str": {}})) // non-struct
	require.EqualError(err, "RegisterType() with type int: Fields: type, int, is not a Struct")
}

func TestDefinition_GettersReturnCopies(t *testing.T) {
	definition := NewDefinition[Child]().
		ValidatesSelf(presentRule{}).
		Validates(RuleMap{"Validates": {presentRule{}}})

	t.Run("self_rules", func(t *testing.T) {
		testRulesGetterIsolation(t, definition.SelfRules)
	})
	t.Run("rule_map", func(t *testing.T) {
		ruleMap := definition.RuleMap()
		require.Equal(t, []Rule{presentRule{}}, ruleMap["Validates"])
		ruleMap["Validates"][0] = mutantRule{}
		ruleMap["New"] = []Rule{mutantRule{}}
		require.Equal(t, RuleMap{"Validates": {presentRule{}}}, definition.RuleMap())
	})
}
