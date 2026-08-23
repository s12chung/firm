# firm

> Declarative validation rules in plain Go--no struct tags.

- Register validation rules once per type; nested structs, pointers, and slices recurse automatically
- Validations return structured, templated, easy to inspect `error`s
- Compose validation rules or implement your own with a 2 function interface
- Zero runtime dependencies

## Quickstart

Register a definition per type (typically in `init()`, which avoids concurrent `map` changes), then validate:

```go
//
// cmd/firm-try/main.go
//

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
			//
			// Can be used to register non-Structs (not recommended though)
			ValidatesSelf(rule.Present{}).
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
```

An invalid config returns a `firm.ErrorMap`, which implements `Error() string`. Try running the code above in [cmd/firm-try/main.go](cmd/firm-try/main.go)--each command's output is commented below:

```sh
go run github.com/s12chung/firm/cmd/firm-try@latest '{"queries":[{"str":""},{"pos":"Noun"}]}'
# main.Config.Queries.[0].Str.Present: Str is not present, main.Config.Queries.[1].Str.Present: Str is not present

go run github.com/s12chung/firm/cmd/firm-try@latest '{}'
# main.Config.Present: main.Config is not present

go run github.com/s12chung/firm/cmd/firm-try@latest '{"queries":[{"str":"hello","pos":"Noun"}]}'
# valid
```

## Validation `error`s

Validation failures return `firm.ErrorMap`, a map of `firm.ErrorKey` to `firm.TemplateError`.

- Keys encode the path to the failure, for example: `<package>.<Type>.<field>.[<index>].<Rule>`
- `firm.ErrorKey` helpers (`RootTypeName()`, `ValueName()`, `ErrorName()`) make it easy to inspect or remap errors programmatically.

## Validation Levels

Validation occurs recursively down levels on these types:

- `firm.Registry` is a mapping of types to `firm.Validator`. `firm.MustRegisterType()` runs `firm.DefaultRegistry.MustRegisterType()`. Validate through the `firm.Registry` for recursion.
- `firm.Validator`s are wrappers around `firm.Rule`s
- `firm.Rule` are validation rules

The default go-to validation function is `ValidateAny(data any) ErrorMap`, which is implemented on `firm.Registry` and `firm.Validator`. Accepts anything--values or pointers, including nil pointers. Also handles invalid types.

- `firm.Registry` unregistered types return "not found in Registry" error. Nice for recursion.
- `firm.Validator` is an interface, built-in validators will return "is not matching type" error

Feel free to make your own registry or skip the registry entirely:

```go
(&firm.Registry{}).MustRegisterType(firm.NewDefinition[Query]().
	Validates(firm.RuleMap{
		"Str": {rule.Present{}},
	}),
)

typedValueValidator := firm.MustNewValue[int](rule.Greater[int]{To: 0})
errMap := typedValueValidator.ValidateAny(-1)
```

- `Validate(data T) ErrorMap` is a typed `ValidateAny()` via generics--the `data` arg is typed, but rules can be any `firm.Rule`. More complicated, but avoids reflection and enforces type safety.
- `ValidateValue(value reflect.Value) ErrorMap` is implemented on all the types described above. Annoying to use, but good abstraction.

```go
typedValueValidator := firm.MustNewValue[int](rule.Greater[int]{To: 0})
errMap := typedValueValidator.Validate(-1)
```

### Rules

```go
type Rule interface {
	ValidateValue(value reflect.Value) ErrorMap
	TypeCheck(typ reflect.Type) *RuleTypeError
}
```

`firm.Rule` is the basic primitive. Built-in rules are in the `rule` package:

| Rule | Checks |
| --- | --- |
| `rule.Present{}` | value is non-zero (and non-empty for strings, slices, arrays, maps, chans) |
| `rule.TrimPresent{}` | string is not empty after `strings.TrimSpace` |
| `rule.Equal[T]{To}` | value equals `To` |
| `rule.Less[T]{OrEqual, To}` | value is less (or equal) than `To` |
| `rule.Greater[T]{OrEqual, To}` | value is greater (or equal) than `To` |
| `rule.Not{Rule}` | negates another rule |
| `rule.Attr{Of, Rule}` | applies a rule to a `rule.Attribute` of the value |

You can implement your own too:

```go
type Even struct{}

func (e Even) ValidateValue(value reflect.Value) firm.ErrorMap {
	if value.Int()%2 == 0 {
		return nil
	}
	return e.ErrorMap()
}

func (e Even) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	if typ.Kind() == reflect.Int {
		return nil
	}
	return firm.NewRuleTypeError("Even", typ, "is not an Int")
}

func (e Even) ErrorMap() firm.ErrorMap {
	return firm.ErrorMap{"Even": firm.TemplateError{Template: "is not even"}}
}
```

The following built-in rules implement `firm.RuleTyped[T any]`, which exposes `Validate(data T)` for convenience really:

- `rule.Equal[T]`, `rule.Less[T]`, `rule.Greater[T]` - the `T` type passes the type implicitly and ensures they're `comparable` or `cmp.Ordered` at compile time
- `rule.TrimPresent` - why not

When you want to implement your own `firm.RuleTyped[T any]`, here's an example:

```go
type RuleBasic interface {
	Rule
	ErrorMap() ErrorMap
}

type RuleTyped[T any] interface {
	RuleBasic
	Validate(data T) ErrorMap
}

type Even struct{}

func (e Even) ValidateValue(value reflect.Value) firm.ErrorMap {
	return e.Validate(int(value.Int()))
}

func (e Even) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	if typ.Kind() == reflect.Int {
		return nil
	}
	return firm.NewRuleTypeError("Even", typ, "is not an Int")
}

func (e Even) ErrorMap() firm.ErrorMap {
	return firm.ErrorMap{"Even": firm.TemplateError{Template: "is not even"}}
}

// Same rule as the `firm.Rule` implementation above
// Validate() is ValidateValue()'s logic, but typed
func (e Even) Validate(data int) firm.ErrorMap {
	if data%2 == 0 {
		return nil
	}
	return e.ErrorMap()
}
```

### Attributes

`rule.Attr` is a struct that implements `firm.Rule`, which allows for rules on derived values.

```go
type Attr struct {
	Of   Attribute
	Rule firm.RuleBasic
}

// trimPresent = Not (TrimSpace Value Equal To "")
trimPresent := rule.Not{ // 1. Not
	Rule: rule.Attr{
		Of:   attr.TrimSpace{},           // 2. TrimSpace Value
		Rule: rule.Equal[string]{To: ""}, // 3. Equal To ""
	},
}
```

The `rule.Attribute` interface does the value derivations. Built-in attributes are in the `attr` package:

| Attribute | Extracts |
| --- | --- |
| `attr.Len{}` | `len(value)` (slices, arrays, maps, chans, strings) |
| `attr.TrimSpace{}` | `strings.TrimSpace(value)` |

### Validators

```go
type Validator interface {
	Rule
	ValidateAny(data any) ErrorMap
	ValidateMerge(value reflect.Value, key string, errorMap ErrorMap)
}
```

A wrapper around `firm.Rule` to avoid reflection and handle `firm.ErrorMap` merging. Basically, a clean way to call `ValidateAny(data any) ErrorMap`. Can be used independently:

```go
typedValueValidator := firm.MustNewValue[int](rule.Greater[int]{To: 0})
errMap := typedValueValidator.ValidateAny(-1)
```

The `firm` package provides:

| Validator | Intent |
| --- | --- |
| `firm.Registry` | registries are validators too, just based on registration with unregistered type checking and recursion |
| `firm.ValueAny` | validates simple values (not structs or slices) |
| `firm.StructAny` | validates structs, mapping fields to rules via `firm.RuleMap` |
| `firm.SliceAny` | slices and arrays, running rules on all elements |
| `firm.RuleValidator{Rule}` | convenience `firm.Validator` struct that wraps around a `firm.Rule` |

Via generics, `firm.ValidatorTyped[T any]` can call `Validate(data T) ErrorMap`. Avoids reflection and enforces type safety. `Any`-suffixed validators above have typed variants which implement `firm.ValidatorTyped[T any]`:

- `firm.Value[T]`, `firm.Struct[T]`, and `firm.Slice[T, U]`

```go
type ValidatorTyped[T any] interface {
	Validator
	Validate(data T) ErrorMap
}
```

`firm.ValidatorTyped[T any]` can use any `firm.Rule`--not just typed ones!

## Examples

Extracted out of [s12chung/text2anki](https://github.com/s12chung/text2anki), where there are real examples validating database entries and HTTP requests.

## License

[MPL-2.0](LICENSE)--changes to firm's own files must be shared under the same license; apps depending on firm are unaffected.
