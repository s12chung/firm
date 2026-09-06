package rule

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/s12chung/firm"
)

func TestMatch_ValidateAll(t *testing.T) {
	rule := Match{Regexp: regexp.MustCompile(`\A[a-z]+\z`)}
	tcs := []struct {
		name     string
		data     string
		hasError bool
	}{
		{name: "match", data: "abc"},
		{name: "no_match", data: "ab1", hasError: true},
		{name: "empty", data: "", hasError: true},
		{name: "case_sensitive", data: "ABC", hasError: true},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var expected firm.ErrorMap
			if tc.hasError {
				expected = rule.ErrorMap()
			}
			require.Equal(t, expected, rule.Validate(tc.data))
			require.Equal(t, expected, rule.ValidateValue(reflect.ValueOf(tc.data)))
		})
	}
}

func TestMatch_ValidateAll_NoAnchors(t *testing.T) {
	// the Regexp is not auto-anchored--it's a substring match
	rule := Match{Regexp: regexp.MustCompile(`[a-z]+`)}
	require.Nil(t, rule.Validate("ab1"))
}

func TestMatch_TypeCheck(t *testing.T) {
	tcs := []struct {
		name         string
		data         any
		badCondition string
	}{
		{name: "string", data: ""},
		{name: "not string", data: 1, badCondition: "is not a String"},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			testTypeCheck(t, tc.data, matchName, tc.badCondition, Match{})
		})
	}
}

func TestMatch_ErrorMap(t *testing.T) {
	testErrorMap(t, Match{Regexp: regexp.MustCompile(`\A[a-z]+\z`)}, "Match: value does not match \\A[a-z]+\\z")
}
