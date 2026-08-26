# firm

> Declarative validation rules in plain Go--no struct tags.

- Register validation rules once per type; recurse into nested structs, pointers, and slices
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

- `firm.Registry` is a mapping of types to `firm.Validator`. `firm.MustRegisterType()` runs `firm.DefaultRegistry.MustRegisterType()`. Also, handles recursion, see [Recursion](#recursion) section
- `firm.Validator` is an interface, where built-in implementations act as wrappers around `firm.Rule`s
- `firm.Rule` are validation rules, which expect **non-pointers only** and all layers of indirection are handled by the built-in `firm.Validator`s--given `**Child` or `*[]*Child`, the rule will receive the same `Child` value

The go-to validation function is `ValidateAny(data any) ErrorMap`, which is implemented on `firm.Registry` and `firm.Validator`. Accepts anything--values or pointers, including nil pointers. Also handles invalid types.

- `firm.Registry` unregistered types return "not found in Registry" error
- `firm.Validator` is an interface, built-in validators will return "is not matching type" error

Feel free to make your own registry or skip the registry entirely:

```go
(&firm.Registry{}).MustRegisterType(firm.NewDefinition[Query]().
	Validates(firm.RuleMap{
		"Str": {rule.Present{}},
	}),
)

typedValueValidator := firm.Value[int](rule.Greater[int]{To: 0})
errMap := typedValueValidator.ValidateAny(-1)
```

- `Validate(data T) ErrorMap` is a typed `ValidateAny()` via generics--the `data` arg is typed, but rules can be any `firm.Rule`. More complicated, but avoids reflection and enforces type safety.
- `ValidateValue(value reflect.Value) ErrorMap` is implemented on all the types described above. Annoying to use, but a good abstraction.

```go
typedValueValidator := firm.Value[int](rule.Greater[int]{To: 0})
errMap := typedValueValidator.Validate(-1)
```

### Rules

`firm.Rule` is the basic primitive.

```go
type Rule interface {
	ValidateValue(value reflect.Value) ErrorMap
	TypeCheck(typ reflect.Type) *RuleTypeError
}
```

`ValidateValue()` and `TypeCheck()` always receive the underlying value and type--**never a pointer**. All layers of indirection are handled by the built-in `firm.Validator`s.

Built-in rules are in the `rule` package:

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

// ValidateValue expects reflect.Value to NOT be a pointer
// firm will indirect pointers and pass the value into the firm.Rule
func (e Even) ValidateValue(value reflect.Value) firm.ErrorMap {
	if value.Int()%2 == 0 {
		return nil
	}
	return e.ErrorMap()
}

// TypeCheck expects reflect.Type to NOT be a pointer
// firm will indirect pointers for you and pass type value into the firm.Rule
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

// ValidateValue expects reflect.Value to NOT be a pointer
// firm will indirect pointers and pass the value into the firm.Rule
func (e Even) ValidateValue(value reflect.Value) firm.ErrorMap {
	return e.Validate(int(value.Int()))
}

// TypeCheck expects reflect.Type to NOT be a pointer
// firm will indirect pointers for you and pass type value into the firm.Rule
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

`rule.Attr` is a struct that implements `firm.Rule`, which allows for rules on derived values called attributes.

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
typedValueValidator := firm.Value[int](rule.Greater[int]{To: 0})
errMap := typedValueValidator.ValidateAny(-1)
```

The `firm` package provides:

| Constructor | Intent |
| --- | --- |
| `firm.Fields[T](ruleMap)` | validates structs, mapping fields to rules via `firm.RuleMap` (`firm.FieldsVldr[T]`) |
| `FieldsAny(type, ruleMap)` | |
| `firm.Elems[[]T](rules...)` | slices and arrays, running rules on all elements |
| `ElemsAny(type, rules...)` | |
| `firm.Value[T](rules...)` | validates simple values (not structs or slices) |
| `ValueAny(type, rules...)` | |

All constructors in the table above `panic()` when there is an error and have a -`WithErr` suffixed version. Naming is intended to be cleanly declarative.

```go
type ValidatorTyped[T any] interface {
	Validator
	Validate(data T) ErrorMap
}
```

Via generics, `firm.ValidatorTyped[T any]` can call `Validate()`. Avoids reflection and enforces type safety. Each of these validators wrap around a -`AnyVldr` suffixed validator, which require a `reflect.Type` with no generics.

`firm.ValidatorTyped[T any]` can use any `firm.Rule`--not just typed ones!

| Constructor | Intent |
| --- | --- |
| `firm.Registry{DefaultValidator}` | Registries are validators too, just based on registration, see [Recursion](#recursion) section |
| `regristry.Backed()/firm.RegistryBacker{Registry}` | handles recursion, essentially a `firm.Registry` wrapper, see [Recursion](#recursion) section |
| `firm.RuleVldr{Rule}` | convenience `firm.Validator` struct that wraps around a `firm.Rule` |

## Recursion

### Registries

Recursion begins with a `firm.Registry`--a mapping of types to `firm.Validator`. Example usage is in the [Quickstart](#quickstart) section.

```go
firm.NewDefinition[Config]().
	ValidatesSelf(rule.Present{}).
	Validates(firm.RuleMap{
		"Queries": {firm.Elems[[]Query](firm.Backed())},
	})
```

1. Register types with `firm.NewDefinition[Config]()`
    a. `ValidatesSelf()` - defines `firm.Rule`s on the type's value "itself"--`Config{}` (usually a `struct`, but you do you)
    b. `Validates()` - defines `firm.Rule`s on the fields
2. Pass in **anything** to `ValidateAny(data any)`--the type is inferred to validate with the correct `firm.Validator`. Like all built-in validators, all layers of indirection are handled--given `**Child` or `*[]*Child`, rules will receive the same `Child` value. Unregistered types return "not found in Registry" error

`ValidateAny()` will basically use these rules:

- `append(validatesSelfRules, firm.FieldsAnyVldr)` - applied to "self"--`Config{}`
- `firm.RegistryBacker{ firm.DefaultRegistry }` - applied to each element of `Config.Queries`

`firm.Backed()` is shorthand for `firm.DefaultRegistry.Backed()`, which reads nice.

`RegistryBacker` basically wraps around `firm.DefaultRegistry`, every call goes to it. The difference is when `RegistryBacker` is passed to `FieldsAnyWithErr`, `Config.Queries`'s type is given to it, allowing access to `Config.Queries`'s validator when `ValidateAny(nil)` is called. `firm.Registry` can't infer the type from `nil` and returns a "not found in Registry" error instead.

### Slices and Arrays

`firm.Elems()` steps down into each element of a slice or array--its element rules apply to every element:

```go
firm.NewDefinition[Config]().
	Validates(firm.RuleMap{
		// For the `Config.Queries` slice field (implied `[]firm.Rule{}`),
		// for each element (`firm.Elems()`),
		// validate using the `Query` definition backed by `firm,DefaultRegistry` `(`firm.Backed()`)
		"Queries": {firm.Elems[[]Query](firm.Backed())},
	})
```

`ElemsWithErr()`/`Elems()` define the type via generics. All rules expect values, so pointer types are indirected, so `ElemsAny(*[]Query)` works on `[]Query`.

For arrays, the generic constraint is too narrow (`T []U` is slices-only)--use `ElemsAny()`/`ElemsAnyWithErr()`, which handle arrays too, but force you to define the type via `reflect`.

## Examples

Extracted out of [s12chung/text2anki](https://github.com/s12chung/text2anki), where there are real examples validating database entries and HTTP requests.

## License

[MPL-2.0](LICENSE)--changes to firm's own files must be shared under the same license; apps depending on firm are unaffected.
