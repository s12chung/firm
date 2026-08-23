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
			// On the `Config` struct "itself", NOT the `Config`'s `StructField`s,
			// validate whether the struct is present (a non-empty value)
			ValidatesTopLevel(rule.Present{}).
			// `StructField`s are represented by a `firm.RuleMap`.
			//
			// For the `Queries` `StructField` as a slice,
			// recurse into each element's registered `Query` definition below
			Validates(firm.RuleMap{"Queries": {}}),
	)
	firm.MustRegisterType(
		// For the `Query` struct
		firm.NewDefinition[Query]().
			// For the `Str` `StructField` as a string,
			// validate whether the string is present (a non-empty value)
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
