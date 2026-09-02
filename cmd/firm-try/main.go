// Command firm-try runs the README quickstart against a JSON string:
//
//	go run github.com/s12chung/firm/cmd/firm-try@latest '{"queries":[{"str":""},{"pos":"Noun"}]}'
//
// nolint:forbidigo
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/s12chung/firm"
	"github.com/s12chung/firm/rule"
)

type Config struct {
	Queries []Query `json:"queries"`
}
type Query struct {
	Str string  `json:"str"`
	POS *string `json:"pos"`
}

func init() {
	//
	// Define validations in `init()` to avoid concurrent `map` changes
	//
	firm.MustRegisterType(firm.NewDefinition[Config]().
		// On the `Config` struct "itself", NOT the `Config`'s fields
		ValidatesSelf(rule.Present{}).
		Validates(firm.RuleMap{
			"Queries": {firm.Elems[[]Query](
				// `firm.Backed()` - validate using registration for `Query` below
				// Basically, explicit recursion
				firm.Backed(),
			)},
		}),
	)
	// For the `Query` struct
	firm.MustRegisterType(firm.NewDefinition[Query]().Validates(firm.RuleMap{
		"Str": {rule.Present{}},
	}).ErrOnNil("POS")) // nil is skipped otherwise
}

func readConfig(body []byte) (Config, error) {
	config := Config{}
	if err := json.Unmarshal(body, &config); err != nil {
		return Config{}, err
	}
	//
	// Run validation
	//
	if errMap := firm.ValidateAny(&config); errMap != nil {
		return Config{}, errMap
	}
	return config, nil
}

// Templates make internationalizing a matter of swapping the template string
var translations = map[string]string{
	"is not present": "no es presento",
	"is nil":         "es nil",
}

func i18nErrorMap(errMap firm.ErrorMap) firm.ErrorMap {
	i18nMap := firm.ErrorMap{}
	for k, v := range errMap {
		if translated, ok := translations[v.Template]; ok {
			v.Template = translated
		}
		i18nMap[k] = v
	}
	return i18nMap
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s '<json>'\n", os.Args[0])
		os.Exit(2)
	}
	if _, err := readConfig([]byte(os.Args[1])); err != nil {
		fmt.Println(err.Error())
		var errMap firm.ErrorMap
		if errors.As(err, &errMap) {
			fmt.Println(i18nErrorMap(errMap).Error())
		}
		os.Exit(1)
	}
	fmt.Println("valid")
}
