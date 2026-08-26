// Command firm-try runs the README quickstart against a JSON string:
//
//	go run github.com/s12chung/firm/cmd/firm-try@latest '{"queries":[{"str":""},{"pos":"Noun"}]}'
//
// nolint:forbidigo
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/s12chung/firm"
	"github.com/s12chung/firm/rule"
)

type Config struct {
	Queries []Query `json:"queries,omitempty"`
}
type Query struct {
	Str string `json:"str,omitempty"`
	POS string `json:"pos,omitempty"`
}

func init() {
	//
	// Define validations (Step 1 of 2)
	// Defined in `init()` to avoid concurrent `map` changes
	//
	firm.MustRegisterType(
		// For the `Config` struct
		firm.NewDefinition[Config]().
			// On the `Config` struct "itself", NOT the `Config`'s fields,
			// validate whether the struct is present (a non-empty value)
			//
			// Can be used to register non-Structs (not recommended though)
			ValidatesSelf(rule.Present{}).
			// Fields are represented by a `firm.RuleMap`.
			//
			// For the `Config.Queries` slice field (implied `[]firm.Rule{}`),
			// for each element (`firm.Elems()`),
			// validate using the `Query` definition backed by `firm,DefaultRegistry` `(`firm.Backed()`)
			Validates(firm.RuleMap{
				"Queries": {firm.Elems[[]Query](firm.Backed())},
			}),
	)
	firm.MustRegisterType(
		// For the `Query` struct
		firm.NewDefinition[Query]().
			// For the `Str` string field (implied `[]firm.Rule{}`),
			// validate whether the string is present--a non-empty value (`rule.Present{}`)
			Validates(firm.RuleMap{
				"Str": {rule.Present{}},
			}))
}

func readConfig(body []byte) (Config, error) {
	config := Config{}
	if err := json.Unmarshal(body, &config); err != nil {
		return Config{}, err
	}
	//
	// Run validation (Step 2 of 2)
	//
	if errMap := firm.ValidateAny(&config); errMap != nil {
		return Config{}, errMap
	}
	return config, nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s '<json>'\n", os.Args[0])
		os.Exit(2)
	}
	if _, err := readConfig([]byte(os.Args[1])); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("valid")
}
