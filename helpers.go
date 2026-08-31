package firm

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

func indirect(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	return value
}

func indirectType(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return typ
}

func copyRuleMap(ruleMap RuleMap) RuleMap {
	copied := make(RuleMap, len(ruleMap))
	for fieldName, rules := range ruleMap {
		copied[fieldName] = slices.Clone(rules)
	}
	return copied
}

// fieldByIndex returns the field Value corresponding to the index
func fieldByIndex(value reflect.Value, index []int) reflect.Value {
	for i, x := range index {
		if i > 0 {
			if value.Kind() == reflect.Pointer {
				if value.IsNil() {
					return reflect.Value{}
				}
				value = value.Elem()
			}
			if value.Kind() != reflect.Struct {
				return reflect.Value{}
			}
		}
		value = value.Field(x)
	}
	return value
}

// typeOfKind returns the indirected type, ensuring the type is one of the kinds
func typeOfKind(vdlrClass string, typ reflect.Type, kinds ...reflect.Kind) (reflect.Type, error) {
	badType := "nil"
	if typ != nil {
		typ = indirectType(typ)
		if slices.Contains(kinds, typ.Kind()) {
			return typ, nil
		}
		badType = typ.String()
	}
	return nil, fmt.Errorf("%s: type, %s, is not a %s", vdlrClass, badType, kindNamesForErr(kinds))
}

func kindNamesForErr(kinds []reflect.Kind) string {
	kindNames := make([]string, len(kinds))
	for i, kind := range kinds {
		name := kind.String()
		kindNames[i] = strings.ToUpper(name[:1]) + name[1:]
	}
	return strings.Join(kindNames, " or ")
}

const keySeparator = "."

func joinKeys[T ~string](keys ...T) T {
	var key T
	for _, v := range keys {
		if v == "" {
			continue
		}
		if key != "" {
			key += keySeparator
		}
		key += v
	}
	return key
}

func sliceErrorKey(i int) string { return "[" + strconv.Itoa(i) + "]" }

func mapErrorKey(key reflect.Value) string {
	if !key.IsValid() {
		return "[<nil>]"
	}
	return "[" + fmt.Sprintf("%v", key.Interface()) + "]"
}
