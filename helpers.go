package firm

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
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

// typeOfKind returns the indirected type, ensuring the type is one of the kinds
func typeOfKind(vdlrClass string, typ reflect.Type, kinds ...reflect.Kind) (reflect.Type, error) {
	kindNames := make([]string, len(kinds))
	for i, kind := range kinds {
		name := kind.String()
		kindNames[i] = strings.ToUpper(name[:1]) + name[1:]
	}
	kindNamesForErr := strings.Join(kindNames, " or ")

	if typ == nil {
		return nil, errors.New(vdlrClass + ": type, nil, is not a " + kindNamesForErr)
	}
	typ = indirectType(typ)
	if !slices.Contains(kinds, typ.Kind()) {
		return nil, fmt.Errorf("%s: type, %v, is not a %s", vdlrClass, typ.String(), kindNamesForErr)
	}
	return typ, nil
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
