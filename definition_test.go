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
}

func TestDefinition_ValidatesSelf(t *testing.T) {
	require := require.New(t)
	rules := []Rule{presentRule{}}

	definition := NewDefinition[Child]().ValidatesSelf(rules...)
	require.Equal(rules, definition.SelfRules())
	require.Panics(func() {
		definition.ValidatesSelf()
	})
}

func TestDefinition_Validates(t *testing.T) {
	require := require.New(t)

	ruleMap := RuleMap{"Validates": {presentRule{}}}
	definition := NewDefinition[Child]().Validates(ruleMap)
	require.Equal(ruleMap, definition.RuleMap())

	require.Panics(func() { definition.Validates(RuleMap{}) })
	require.Panics(func() { NewDefinition[Child]().Validates(RuleMap{"DoesNotExist": {}}) })
	require.Panics(func() { NewDefinition[Child]().Validates(RuleMap{"private": {}}) })
}
