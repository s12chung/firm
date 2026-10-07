// Package rule contains the default firm.Rule
package rule

import "github.com/s12chung/firm"

// rulesOrPanic returns rules, panicking when none--Or() and And() require at least one
func rulesOrPanic(rules []firm.RuleBasic, constructor string) []firm.RuleBasic {
	if len(rules) == 0 {
		panic(constructor + "() called without rules")
	}
	return rules
}
