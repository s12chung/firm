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
	// Register a type to `firm.DefaultRegistry`
	firm.MustRegisterType(
		// For the `firm.Definition` of the `Config` struct, which represents a `firm.FieldsAnyVldr`
		firm.NewDefinition[Config]().
			// On the `Config` struct "itself", NOT the `Config`'s fields,
			// validate whether the struct is present (a non-empty value)
			ValidatesSelf(rule.Present{}).
			// Fields are represented by a `firm.RuleMap`.
			//
			// For the `Config.Queries` slice field,
			// for each element (`firm.Elems()`),
			// validate using the validation defined for `Query` "backed" by `firm.DefaultRegistry` `(`firm.Backed()`)
			//
			// Replacing `firm.Backed()` with `firm.Fields[Query](firm.RuleMap{"Str": {rule.Present{}}})`
			// will do the same behavior--repeating the `Definition` below. `firm.Backed()` is basically explicit recursion.
			//
			// `firm.Backed()` will NOT be applied to `nil` pointers
			Validates(firm.RuleMap{
				"Queries": {firm.Elems[[]Query](firm.Backed())},
			}),
	)
	// Register a type to `firm.DefaultRegistry`
	firm.MustRegisterType(
		// For the `firm.Definition` of the `Query` struct, which represents a `firm.FieldsAnyVldr`
		firm.NewDefinition[Query]().
			// For the `Str` string field,
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
- `firm.Validator` is an interface, where built-in implementations act as wrappers around `firm.Rule`s. Handles all pointers and layers of indirection for `firm.Rule` to ensure "safe values" (see [Types, Pointers, and Invalid Values](#types-pointers-and-invalid-values)). Given `[]Child` or `*[]**Child`, validators will traverse the slice and receive the same `Child` value. The same rules will apply the same slice and same `Child`.
- `firm.Rule` are validation rules, which expect **non-pointers only**

The go-to validation function is `ValidateAny(data any) ErrorMap`, which is implemented on `firm.Registry` and `firm.Validator`. Accepts anything--values or pointers, including nil pointers. Also handles invalid types.

- `firm.Registry` unregistered types return "not found in Registry" error
- `firm.Validator` is an interface, built-in validators will return "is not matching type" error

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
```

- `Validate(data T) ErrorMap` is a typed `ValidateAny()` via generics--the `data` arg is typed, but rules can be any `firm.Rule`. More complicated, but avoids reflection and enforces type safety.
- `ValidateValue(value reflect.Value) ErrorMap` is implemented on all the types described above. Unlike `ValidateAny()/Validate()`, it expects "safe values" (see [Types, Pointers, and Invalid Values](#types-pointers-and-invalid-values)).

```go
typedValueValidator := firm.Value[int](rule.Greater[int]{To: 0})
errMap := typedValueValidator.Validate(-1)
```

### Types, Pointers, and Invalid Values

To simplify `firm.Validator` and `firm.Rule` implementations, there are caller contracts to contain complexity:

- **Type Coherence**: On validation creation (`RegisterType()` or any validator constructor), `firm.Rule.TypeCheck()` is called to ensure type coherence with the validator
- **Safe Values**: Safe values are defined as non-pointer valid `reflect.Value`s. There are two public use interfaces: `ValidateAny()/Validate()`. **Only these functions** may receive an unsafe value. Pointers are indirected and invalid  `reflect.Value`s return a `firm.ErrInvalidValue()`. They must also ensure that only safe values are passed down to:
   - `firm.Rule`
   - `ValidateMerge()` - `ValidateMerge()` may need to recurse, when doing so, it must also ensure that only safe values are passed down to `firm.Rule`

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
	// Public use interface - may receive an "unsafe" value and answers it with firm.ErrInvalidValue(),
	// must also ensure only "safe" values are passed down to:
	// - firm.Rule
	// - ValidateMerge() below
	//
	// See Types, Pointers, and Invalid Values section above
	ValidateAny(data any) ErrorMap
	ValidateMerge(value reflect.Value, key string, errorMap ErrorMap)
}
```

A wrapper around `firm.Rule` to avoid reflection and handle `firm.ErrorMap` merging. Basically, a clean way to call `ValidateAny(data any) ErrorMap`. Can be used independently, like in the example below.

```go
typedValueValidator := firm.Value[int](rule.Greater[int]{To: 0})
errMap := typedValueValidator.ValidateAny(-1)
```

The `firm` package provides:

| Constructor | Intent |
| --- | --- |
| `firm.Fields[T](ruleMap)` | validates structs, mapping fields to rules via `firm.RuleMap`. fields must be exported. returns `firm.FieldsVldr[T]` |
| `FieldsAny(type, ruleMap)` | same as above. returns `firm.FieldsAnyVldr` |
| `firm.Elems[[]T](rules...)` | slices and arrays, running rules on all elements. returns `firm.ElemsVldr[[]T]` |
| `ElemsAny(type, rules...)` | same as above. returns `firm.ElemsAnyVldr` |
| `firm.Keys[map[K]V](rules...)` | maps, running rules on all keys. returns `firm.KeysVldr[map[K]V]` |
| `KeysAny(type, rules...)` | same as above. returns `firm.KeysAnyVldr` |
| `firm.Values[map[K]V](rules...)` | maps, running rules on all values. returns `firm.ValuesVldr[map[K]V]` |
| `ValuesAny(type, rules...)` | same as above. returns `firm.ValuesAnyVldr` |
| `firm.KeyValues[map[K]V](rules...)` | maps, running rules on all key-value pairs, passing each key-value pair as a `map[K]V` with only 1 key-value pair to validate. returns `firm.KeyValuesVldr[map[K]V]` |
| `KeyValuesAny(type, rules...)` | same as above. returns `firm.KeyValuesAnyVldr` |
| `firm.Value[T](rules...)` | validates simple values. returns `firm.ValueVldr[T]` |
| `ValueAny(type, rules...)` | same as above. returns `firm.ValueAnyVldr` |

All constructors in the table above `panic()` when there is an error and have a -`WithErr` suffixed version. Naming is intended to be cleanly declarative.

```go
type ValidatorTyped[T any] interface {
	Validator
	// Public use interface - may receive an "unsafe" value and answers it with firm.ErrInvalidValue(),
	// must also ensure only "safe" values are passed down to:
	// - firm.Rule
	// - ValidateMerge() below
	//
	// See Types, Pointers, and Invalid Values section above
	Validate(data T) ErrorMap
}
```

Via generics, `firm.ValidatorTyped[T any]` can call `Validate()`. Avoids reflection and enforces type safety. Each of these validators wrap around a -`AnyVldr` suffixed validator, which require a `reflect.Type` with no generics.

`firm.ValidatorTyped[T any]` can use any `firm.Rule`--not just typed ones!

| Constructor | Intent |
| --- | --- |
| `firm.Registry{DefaultValidator}` | Registries are validators too, just based on registration, see [Recursion](#recursion) section |
| `registry.Backed()/firm.RegistryBacker{Registry}` | handles recursion, essentially a `firm.Registry` wrapper, see [Recursion](#recursion) section |
| `firm.RuleVldr{Rule}` | convenience `firm.Validator` struct that wraps around a `firm.Rule` |

Implement your own `firm.Validator` with these helpers:

- `firm.ImplValidateAny(v, data)` - implementation calls `firm.TypeCheck()` indirects pointers, and returns `firm.ErrInvalidValue()` for unsafe values
- `firm.ImplValidateValue(v, value)` - implementation assumes `TypeCheck` is called
- `firm.ImplValidateMerge(value, key, errorMap, rules)` - implementation assumes `TypeCheck` is called, as it iterates `rules` and merges them into the errorMap
- `firm.ImplValidate(v, data)` - implementation does no type checking is done on runtime because `Validate()` is a typed function (often with generics)
- `firm.TypeCheckRules(typ, rules, errContext)` - calls `firm.RuleTypeCheck()` of each rule with the indirected `typ`, wrapping any error with `errContext` and handles the [Registry.Backed() gotcha](#registries)

Looking at `firm.ValueAnyVldr` in ([validator.go](validator.go)) as an example is recommended.

## Recursion

### Registries

Recursion begins with a `firm.Registry`--a map of types to `firm.Validator`s ("registrations"). Below is the same example usage from the [Quickstart](#quickstart) section.

```go
// Register a type to `firm.DefaultRegistry`
firm.MustRegisterType(
	// For the `firm.Definition` of the `Config` struct, which represents a `firm.FieldsAnyVldr`
	firm.NewDefinition[Config]().
		// On the `Config` struct "itself", NOT the `Config`'s fields,
		// validate whether the struct is present (a non-empty value)
		ValidatesSelf(rule.Present{}).
		// Fields are represented by a `firm.RuleMap`.
		//
		// For the `Config.Queries` slice field (implied `[]firm.Rule{}`),
		// for each element (`firm.Elems()`),
		// validate using the validation defined for `Query` "backed" by `firm.DefaultRegistry` `(`firm.Backed()`)
		//
		// Replacing `firm.Backed()` with `firm.Fields[Query](firm.RuleMap{"Str": {rule.Present{}}})`
		// will do the same behavior--repeating the `Definition` below. `firm.Backed()` is basically explicit recursion.
		//
		// `firm.Backed()` will NOT be applied to `nil` pointers
		Validates(firm.RuleMap{
			"Queries": {firm.Elems[[]Query](firm.Backed())},
		}),
)
```

`firm.Registry` takes a `firm.Definition`, which creates a `firm.FieldsAnyVldr` for the type definition.

`firm.MustRegisterType()/Backed()` is shorthand for `firm.DefaultRegistry.MustRegisterType()/Backed()`. You can use your own registry too (`&firm.Registry{}`). Registries allow explicit recursion "backed" by `firm.DefaultRegistry` by using the type map to find `Query`'s validator.

Pass **anything** to `ValidateAny(data any)`--the type is inferred to validate with the correct `firm.Validator`. Like all built-in validators, all pointers are indirected to ensure "safe values" (see [Types, Pointers, and Invalid Values](#types-pointers-and-invalid-values)). Given `[]Child` or `*[]**Child`, validators will traverse the slice and receive the same `Child` value. The same rules will apply the same slice and same `Child`. Unregistered types return "not found in Registry" error.

`firm.DefaultRegistry.Backed()` returns a `firm.RegistryBacker`, which basically proxies every call to `firm.DefaultRegistry` and handles a gotcha.

In detail, `firm.Registry` can't infer the type from `nil` when `ValidateAny(nil)` is called, so a "not found in Registry" error is returned. `firm.RegistryBacker` covers the gotcha as thus:

1. From `firm.MustRegisterType()` calling the `firm.FieldsAnyWithErr()` constructor, `firm.RegistryBacker` is given `Config.Queries`'s type
2. Given the type, `firm.RegistryBacker` always has access to `Config.Queries`'s validator
3. So `Config.Queries`'s rules are applied to that field's elements (`Query` structs)

All constructors of built-in **recursive** validators (`Fields()`, `Elems()`, `Keys()`, `KeyValues()`, typed/non-typed, with/without errors, etc.):

1. Do Step 1 above. To cover the gotcha, pass `firm.RegistryBacker` into these validators **directly**
2. Implement `AllRules() []firm.Rule` of the `firm.RuleLister` interface, which allows for recursion cycle checks through `MustRegisterType()`

### Slices and Arrays

`Elems[[]T]()` returns `firm.ElemsVldr[[]T]`, which applies its rules into each element of a slice or array:

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

For arrays, the generic constraint is too narrow (`T []U` is slices-only)--use the non-pointer versions, `ElemsAny()/ElemsAnyWithErr()`, which handle arrays too, but force you to define the type via `reflect.Type`. The type is indirected as well.

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

When returning errors, pointer keys are indirected  (e.g. `[mykey]`, `[<nil>]` for a nil key), as addresses are unstable and unreadable.

To avoid [reflection](https://github.com/golang/go/issues/45591) [issues](https://github.com/golang/go/issues/54393), `KeyValues[map[K]V]()` iterates over key-value pairs a `map` (`map[K]V`) with only a single key-value pair.

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

## Examples

Extracted out of [s12chung/text2anki](https://github.com/s12chung/text2anki), where there are real examples validating database entries and HTTP requests.

## License

[MPL-2.0](LICENSE)--changes to firm's own files must be shared under the same license; apps depending on firm are unaffected.
