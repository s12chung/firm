package firm

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNotFoundRule_ValidateValue(t *testing.T) {
	require := require.New(t)

	typed := ErrorMap{"NotFound": TemplateError{
		Template:       "type, {{.ValueTypeName}}, not found in Registry",
		TemplateFields: map[string]string{"ValueTypeName": "int"},
	}}
	require.Equal(typed, NotFoundRule{}.ValidateValue(reflect.ValueOf(1)))
	require.Equal("NotFound: value type, int, not found in Registry", typed.Finish().Error())

	require.Equal(NotFoundRule{}.ErrorMap(), NotFoundRule{}.ValidateValue(reflect.Value{}))
	require.Equal("NotFound: value type, NoType, not found in Registry", NotFoundRule{}.ErrorMap().Finish().Error())

	// nilType stand-ins are not named
	require.Equal(NotFoundRule{}.ErrorMap(), NotFoundRule{}.ValidateValue(nilValue))
	require.Equal("NotFound: value type, NoType, not found in Registry", NotFoundRule{}.ValidateValue(nilValue).Finish().Error())
}
