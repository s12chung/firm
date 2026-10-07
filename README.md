# firm

> Declarative validations in plain Go--with recursive, composable, and customizable rules.

- Register validation rules once per type; explicitly recurse into nested structs, pointers, slices, and maps
- Validations return structured, templated, easy to inspect `error`s
- Compose validation rules or implement your own with a 2 function interface
- All validation errors are from explicitly declared validations, except type checking
- Zero runtime dependencies

## Quickstart

Register a definition per type, then validate.

```go
type Query struct {
	Str string  `json:"str"`
	POS *string `json:"pos"`
}

func init() {
	firm.MustRegisterType(firm.NewDefinition[Query]().Validates(firm.RuleMap{
		"Str": {rule.Present{}},
	}).NotNil("POS"))
}

func validateQuery(query Query) error {
	return firm.ValidateAny(query)
}
```

Or the longer and commented version:

```go
//
// cmd/firm-try/main.go
//

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
	}).NotNil("POS")) // nil is skipped otherwise
}

func readConfig(body []byte) (Config, error) {
	config := Config{}
	if err := json.Unmarshal(body, &config); err != nil {
		return Config{}, err
	}
	//
	// Run validation (Step 2 of 2)
	//
	if errMap := firm.ValidateAny(config); errMap != nil {
		return Config{}, errMap
	}
	return config, nil
}
```

Validation failures return `firm.ErrorMap` (a map of `firm.ErrorKey` to `firm.TemplateError`), which implements `error`. Error messages prefix their `ValueName` (ex. "POS" in "POS is nil"); use `ErrorWith()` to change it or add a suffix (templated with `{{value}}`, `{{.ValueName}}`, `{{RootTypeName}}` and `TemplateError.TemplateFields` key/values).

`firm.ErrorKey` is easy to inspect or remap errors programmatically. Its keys encode the path to the failure with helpers (`RootTypeName()/ValueName()/ErrorName()`):

```text
// |  root type   | |-----value-----| error
// <package>.<Type>.<field>.[<index>].<Rule>
```

Try running the code above in [cmd/firm-try/main.go](cmd/firm-try/main.go)--each command's output is commented below and demonstrates easy i18n:

```sh
go run github.com/s12chung/firm/cmd/firm-try@latest '{"queries":[{"str":""},{"pos":"Noun"}]}'
# main.Config.Queries.[0].POS.Nil: POS is nil, main.Config.Queries.[0].Str.Present: Str is not present, main.Config.Queries.[1].Str.Present: Str is not present
# main.Config.Queries.[0].POS.Nil: POS es nil, main.Config.Queries.[0].Str.Present: Str no es presento, main.Config.Queries.[1].Str.Present: Str no es presento

go run github.com/s12chung/firm/cmd/firm-try@latest '{}'
# main.Config.Present: main.Config is not present
# main.Config.Present: main.Config no es presento

go run github.com/s12chung/firm/cmd/firm-try@latest '{"queries":[{"str":"hello","pos":"Noun"}]}'
# valid
```

## Validation Levels
```text
          firm.MustRegisterType()
            firm.RegisterType()
               firm.Backed()
             firm.ValidateAny()
                     |
     proxies calls to firm.DefaultRegistry
                     v
             **firm.Registry**
                     |
             maps reflect.Type to
                     v
        **firm.Validator (interface)**
    Ensures "unsafe values" are encapsulated
                     |
                calls many
                     v
          **firm.Rule (interface)**
Validation rules, expects non-pointers only ("safe values")
```

See [Types, Pointers, and Safe Values](#types-pointers-and-safe-values) for details on "safe values".

`ValidateAny(data any) ErrorMap` is the **go-to validation function**, which is implemented on `firm.Registry` and `firm.Validator`. Accepts anything--values or pointers, including `nil` pointers. By default, `nil` pointers are skipped unless `NotNil()/NotNilSelf()` is called. Also handles invalid types. All errors are from explicit validation declarations, except for type checking:

- `firm.Registry` unregistered types return "not found in Registry" error
- `firm.Validator` is an interface, built-in validators will return "is not matching type" error

`Validate(data T) ErrorMap` is a typed `ValidateAny()` via generics--the `data` arg is typed, but rules can be any `firm.Rule`. Enforces type safety.

Feel free to make your own registry or skip the registry entirely:

```go
registry := &firm.Registry{}
registry.MustRegisterType(firm.NewDefinition[Query]().
	Validates(firm.RuleMap{
		"Str": {rule.Present{}},
	}),
)

typedValueValidator := firm.Value[int](rule.Greater[int]{To: 0})
errMap := typedValueValidator.ValidateAny(-1)
sameErrMap := typedValueValidator.Validate(-1)
```

### Rules

`firm.Rule` is a validation rule.

```go
type Rule interface {
	ValidateValue(value reflect.Value) ErrorMap
	TypeCheck(typ reflect.Type) *RuleTypeError
}
```

`ValidateValue()` and `TypeCheck()` always receive the underlying value and type--**never a pointer**. All layers of indirection are handled by the built-in `firm.Validator`s (see [Types, Pointers, and Safe Values](#types-pointers-and-safe-values)).

Built-in rules are in the `rule` package:

| Rule | Types | Checks |
| --- | --- | --- |
| **Presence** | | |
| `rule.Present{}` | any | value is non-zero (and non-empty for `Len()`-able types) |
| `rule.TrimPresent{}` | string | string is not empty after `strings.TrimSpace` |
| **Format & length** | | |
| `rule.Match{Regexp}` | string | string matches `Regexp` |
| `rule.Len{Is, Min, Max}` | `Len()`-able | length of value is `Is` or between `Min` and `Max` |
| **Comparison** | | |
| `rule.OneOf[T]{Values, ValuesFunc}` | comparable | value is one of `Values` or the result of `ValuesFunc()` (both must not be set) |
| `rule.Equal[T]{To}` | comparable | value equals `To` |
| `rule.Less[T]{OrEqual, To}` | `cmp.Ordered` | value is less (or equal) than `To` |
| `rule.Greater[T]{OrEqual, To}` | `cmp.Ordered` | value is greater (or equal) than `To` |
| **Struct** | | |
| `rule.OneNotNil{Fields}` | struct | exactly one of the named Fields (that are pointer types) is not nil |

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

// ErrorMap implements the firm.RuleBasic interface (see below)
func (e Even) ErrorMap() firm.ErrorMap {
	// Built-in rules return one error; you may return multiple errors for complex rules
	return firm.ErrorMap{"Even": firm.TemplateError{Template: "is not even"}}
}
```

Implement `firm.RuleBasic` to use firm's built-in composition rules:

```go
type RuleBasic interface {
	Rule
	ErrorMap() ErrorMap
}
```

| Rule Constructor `func()` | Checks |
| --- | --- |
| `rule.SetName(Name, Rule)` | renames the error key of `Rule` to `Name` |
| `rule.ShowValue(Rule)` | appends the failing value to `Rule`'s messages: `ID is not one of ["a"]: "b"` |
| `rule.CustomizeErr(Rule, func(ErrorMap) ErrorMap)` | customizes the ErrorMap of `Rule` through the `func` |
| `rule.Not(Rule)` | negates another rule |
| `rule.Or(Rules...)` | value is valid for any of the `Rules` |
| `rule.And(Rules...)` | value is valid for all of the `Rules` (only for easier composition--`[]Rule` is passed throughout in `firm.Validator` and `firm.Definition` as an AND) |

The following built-in rules implement `firm.RuleTyped[T any]`, which exposes `Validate(data T)` for convenience really:

- `rule.Equal[T]`, `rule.Less[T]`, `rule.Greater[T]`, `rule.OneOf[T]` - the `T` type passes the type implicitly and ensures they're `comparable` or `cmp.Ordered` at compile time
- `rule.TrimPresent`, `rule.Match` - why not

When you want to implement your own `firm.RuleTyped[T any]`, here's an example:

```go
type RuleTyped[T any] interface {
	RuleBasic
	Validate(data T) ErrorMap
}

// Same `Even` rule as the `firm.Rule` implementation above with `TypeCheck()` and `ErrorMap()`
func (e Even) ValidateValue(value reflect.Value) firm.ErrorMap {
	return e.Validate(int(value.Int()))
}

// Validate() is ValidateValue()'s logic, but typed
func (e Even) Validate(data int) firm.ErrorMap {
	if data%2 == 0 {
		return nil
	}
	return e.ErrorMap()
}
```

### Validators

```go
type Validator interface {
	Rule
	ValidateAny(data any) ErrorMap
	ValidateMerge(value reflect.Value, key string, errorMap ErrorMap)
}
```

`firm.Validator` is wrapper around `firm.Rule`s to handle pointers and `firm.ErrorMap` merging. Basically, a clean way to call `ValidateAny(data any) ErrorMap`. Can be used independently, like in the example below.

```go
typedValueValidator := firm.Value[int](rule.Greater[int]{To: 0})
errMap := typedValueValidator.ValidateAny(-1)
```

The `firm` package provides:

| Constructor | Type | Intent |
| --- | --- | --- |
| `firm.Fields[T](ruleMap)` / `firm.FieldsAny(type, ruleMap)` | struct | mapping fields to rules via `firm.RuleMap`. fields must be exported. |
| `firm.Elems[[]T](rules...)` / `firm.ElemsAny(type, rules...)` | slice/array | running rules on all elements. arrays via `ElemsAny()`. |
| `firm.Keys[map[K]V](rules...)` / `firm.KeysAny(type, rules...)` | map | running rules on all keys. |
| `firm.Values[map[K]V](rules...)` / `firm.ValuesAny(type, rules...)` | map | running rules on all values. |
| `firm.KeyValues[map[K]V](rules...)` / `firm.KeyValuesAny(type, rules...)` | map | running rules on all key-value pairs, passing each as a `map[K]V` with only 1 key-value pair to validate. |
| `firm.Value[T](rules...)` / `firm.ValueAny(type, rules...)` | any | running rules on the value. |

All constructors in the table above `panic()` when there is an error and have a -`WithErr` suffixed version. Naming is intended to be cleanly declarative.

By default, validators can recursively traverse through `Fields`, `Elems`, `Keys`, and `Values` that may contain pointers. All pointers are indirected to ensure "safe values" (see [Types, Pointers, and Safe Values](#types-pointers-and-safe-values)). Given `[]Child` or `*[]**Child`, validators will traverse the slice and receive the same `Child` value. Both forms will apply the same rules to the same `Child` values. And by default, `nil` pointers are skipped. To provide errors on `nil` pointers instead, call `NotNil()`.

```go
// For Elems (slice or array), require non-nil `Queries` elements
firm.Elems[[]*Query](firm.Backed()).NotNil()

// For Fields (struct), require non-nil `Child`
firm.Fields[Parent](firm.RuleMap{"Child": {firm.Backed()}}).NotNil("Child")

// Same as `firm.Fields`, but via a Definition
firm.NewDefinition[Parent]().Validates(firm.RuleMap{
	"Child": {firm.Backed()},
}).NotNil("Child")
```

The value arg of `ValidateAny()/Validate()` is skipped when `nil` too. To error instead, call `NotNilSelf()`.

```go
// Require non-nil, when calling `ValidateAny()/Validate()` with a `nil` `*Config`
firm.Fields[Config](firm.RuleMap{"Queries": {firm.Elems[[]Query](firm.Backed())}}).NotNilSelf()

// Same, but via a Definition--errors when `firm.Registry.ValidateAny()` receives a `nil` `*Config`
firm.NewDefinition[Config]().ValidatesSelf(rule.Present{}).NotNilSelf()
```

Via generics, `firm.ValidatorTyped[T any]` can call `Validate()`, which enforces type safety. Each of these validators wrap around a `*AnyVldr`. `firm.ValidatorTyped[T any]` can use any `firm.Rule`--not just typed ones!

```go
type ValidatorTyped[T any] interface {
	Validator
	Validate(data T) ErrorMap
}
```

See [Implementing Validators](#implementing-validators) for more.

## Recursion

### Registries

Type recursion begins with a `firm.Registry`. Below is the same example from the [Quickstart](#quickstart) section.

```go
firm.MustRegisterType(firm.NewDefinition[Config]().
	// On the `Config` struct "itself", NOT the `Config`'s fields
	ValidatesSelf(rule.Present{}).
	Validates(firm.RuleMap{
		"Queries": {firm.Elems[[]Query](
			// `firm.Backed()` - validate using registration for `Query`
			// Basically, explicit recursion
			firm.Backed(),
		)},
	}),
)
```

`firm.Registry` is a map of types to `firm.Validator`s ("registrations").

`firm.MustRegisterType()/Backed()` is shorthand for `firm.DefaultRegistry.MustRegisterType()/Backed()`. You can use your own registry too (`&firm.Registry{}`). Registries allow explicit type recursion "backed by the Registry" using the type map to find `Query`'s validator.

Pass **anything** to `ValidateAny(data any)`--the type is inferred to validate with the correct `firm.Validator`. Like all validators, all pointers are indirected to ensure "safe values" (see [Types, Pointers, and Safe Values](#types-pointers-and-safe-values)). Unregistered types return "not found in Registry" error.

`firm.DefaultRegistry.Backed()` returns a `firm.RegistryBacker`, which basically proxies every call to `firm.DefaultRegistry` and handles a gotcha--`firm.Registry` can't infer the type from `nil` when `ValidateAny(nil)` is called.

### Slices and Arrays

`Elems[[]T]()` returns `firm.ElemsVldr[[]T]`, which applies its rules into each element of a slice:

```go
// For each element (`firm.Elems()`),
// validate whether the `Child` struct is present--a non-empty value (`rule.Present{}`)
elementsValidator := firm.Elems[[]Child](rule.Present{})

toValidate := []Child{Child{Name: "Valid"}}
elementsValidator.ValidateAny(toValidate)
elementsValidator.Validate(toValidate)     // Typed validation
elementsValidator.ValidateAny(&toValidate) // All pointers are indirected, so valid too

// ptElementsValidator does the same operations as elementsValidator, but with pointer elements indirected
ptElementsValidator := firm.Elems[[]**Child](rule.Present{})

child := Child{Name: "Also Valid"}
childPt := &child
toValidatePt := []**Child{&childPt}

ptElementsValidator.ValidateAny(toValidatePt)
ptElementsValidator.Validate(toValidatePt)     // Typed validation
ptElementsValidator.ValidateAny(&toValidatePt) // All pointers are indirected, so valid too
```

`Elems[[]T]()/ElemsWithErr[[]T]()` define the type via generics.

For arrays, the generic constraint is too narrow (`T []U` is slices-only)--use the non-generic versions, `ElemsAny()/ElemsAnyWithErr()`, which handle arrays too, but force you to define the type via `reflect.Type`. The type is indirected as well.

### Maps

`Keys[map[K]V]()/Values[map[K]V]()/KeyValues[map[K]V]()` return `firm.KeysVldr/ValuesVldr/KeyValuesVldr`, which applies its rules into each key, value, key-value pair of a map:

```go
// For each value (`firm.Values()`),
// validate whether the `Child` struct is present--a non-empty value (`rule.Present{}`)
valuesValidator := firm.Values[map[string]Child](rule.Present{})

toValidate := map[string]Child{"Valid": {Name: "ok"}}
valuesValidator.ValidateAny(toValidate)
valuesValidator.Validate(toValidate)     // Typed validation
valuesValidator.ValidateAny(&toValidate) // All pointers are indirected, so valid too

// ptValuesValidator does the same operations as valuesValidator, but with pointer values indirected
ptValuesValidator := firm.Values[map[string]**Child](rule.Present{})

child := Child{Name: "ok"}
childPt := &child
toValidatePt := map[string]**Child{"Also Valid": &childPt}

ptValuesValidator.ValidateAny(toValidatePt)
ptValuesValidator.Validate(toValidatePt)     // Typed validation
ptValuesValidator.ValidateAny(&toValidatePt) // All pointers are indirected, so valid too
```

When returning errors, pointer keys are indirected (e.g. `[mykey]`, `[<nil>]` for a nil key), as addresses are unstable and unreadable.

To avoid [type](https://github.com/golang/go/issues/45591) [complications](https://github.com/golang/go/issues/54393), `KeyValues[map[K]V]()` iterates over key-value pairs, passing each down as a `map[K]V` with only a single key-value pair.

```go
// For each key-value pair, validate whether the value is true if the key is even
type IsEven struct{}

func (e IsEven) ValidateValue(value reflect.Value) firm.ErrorMap {
	errorMap := firm.ErrorMap{}
	for iter := value.MapRange(); iter.Next(); {
		if iter.Key().Int()%2 != 0 || iter.Value().Bool() {
			continue
		}
		errorMap["IsEven"] = firm.TemplateError{Template: "is not true"}
	}
	return errorMap.ToNil()
}

func (e IsEven) TypeCheck(typ reflect.Type) *firm.RuleTypeError {
	if typ.Kind() == reflect.Map && typ.Key().Kind() == reflect.Int && typ.Elem().Kind() == reflect.Bool {
		return nil
	}
	return firm.NewRuleTypeError("IsEven", typ, "is not a Map with an Int key and Bool value")
}

firm.KeyValues[map[int]bool](IsEven{})
```

## Types, Pointers, and Safe Values

Validators must uphold two caller contracts--**enforced by the caller** to simplify implementations in `firm.Validator` and `firm.Rule`. `firm` does not guard for broken caller contracts.

**Type Coherence**: On validation creation (`RegisterType()` or any validator constructor), `firm.Rule.TypeCheck()` is called to ensure type coherence with the validator

**Safe Values**: Safe values are defined as a non-pointer valid `reflect.Value`. Only `ValidateAny()/Validate()` may receive an unsafe value. Pointers are indirected and `nil` pointers at these entry points are skipped by default. Built-in validators can call `NotNilSelf()` to return a `firm.ErrNilPointer()` instead. `ValidateAny()/Validate()` must also ensure that only safe values are passed down to:

- `firm.Rule`
- `ValidateMerge()`

`ValidateMerge()` is the only place within `firm` that contains unsafe values, as it may need to recurse. When doing so, `ValidateMerge()` converts unsafe values to safe values by:

- Indirecting the pointers
- Skipping `nil` pointers by default. Built-in validators can call `NotNil()` to merge a `firm.ErrNilPointer()` before skipping.

The safe values flow looks like this:

> unsafe values -> `ValidateAny()/Validate()` -> safe values -> `ValidateMerge()` unpacks an unsafe value from a recursive value -> within `ValidateMerge()`, makes the unpacked value safe -> safe values until `ValidateMerge()` unpacks another recursive value

## Usage

Extracted out of [s12chung/text2anki](https://github.com/s12chung/text2anki), where there are real examples validating database entries and HTTP requests.

## License

[MPL-2.0](LICENSE)--changes to firm's own files must be shared under the same license; apps depending on firm are unaffected.
