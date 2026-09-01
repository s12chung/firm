package firm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type parent struct {
	Child

	Primitive               int
	Basic                   Child
	Pt                      *Child
	Any                     any
	Array                   []Child
	ArrayPt                 []*Child
	PtArray                 *[]Child
	Map                     map[string]Child
	MapPt                   map[string]*Child
	PtMap                   *map[string]Child
	PrimitiveEmptyValidates int
	BasicEmptyValidates     Child
	PtEmptyValidates        *Child
	AnyEmptyValidates       any
	SliceValidates          []Child
	SlicePtValidates        []*Child
	PtSliceValidates        *[]Child
	MapValidatesValues      map[string]Child
	MapPtValidatesValues    map[string]*Child
	PtMapValidatesValues    *map[string]Child
	MapValidatesKeys        map[string]Child
	MapPtValidatesKeys      map[string]*Child
	PtMapValidatesKeys      *map[string]Child
	MapValidatesKeyValues   map[string]Child
	MapPtValidatesKeyValues map[string]*Child
	PtMapValidatesKeyValues *map[string]Child
	PrimitiveNoValidates    int
	BasicNoValidates        Child
	PtNoValidates           *Child
	AnyNoValidates          any
	SliceNoValidates        []Child
	SlicePtNoValidates      []*Child
	PtSliceNoValidates      *[]Child
	MapNoValidates          map[string]Child
	MapPtNoValidates        map[string]*Child
	PtMapNoValidates        *map[string]Child
}

type Child struct {
	Validates   string
	NoValidates string
	private     string //nolint:unused // it's used
}

func fullParent() parent {
	fc := func() *Child {
		return &Child{Validates: "Child validates", NoValidates: "no validates"}
	}
	return parent{
		Child: *fc(),
		// validate field + Child
		Primitive: 1, Basic: *fc(), Pt: fc(), Any: *fc(),
		Array: []Child{*fc(), *fc()}, ArrayPt: []*Child{fc(), fc()}, PtArray: &[]Child{*fc(), *fc()},
		Map: map[string]Child{"1": *fc(), "2": *fc()}, MapPt: map[string]*Child{"1": fc(), "2": fc()},
		PtMap: &map[string]Child{"1": *fc(), "2": *fc()},
		// validate Child
		PrimitiveEmptyValidates: 1, BasicEmptyValidates: *fc(), PtEmptyValidates: fc(), AnyEmptyValidates: *fc(),
		SliceValidates: []Child{*fc(), *fc()}, SlicePtValidates: []*Child{fc(), fc()}, PtSliceValidates: &[]Child{*fc(), *fc()},
		MapValidatesValues: map[string]Child{"1": *fc(), "2": *fc()}, MapPtValidatesValues: map[string]*Child{"1": fc(), "2": fc()},
		PtMapValidatesValues: &map[string]Child{"1": *fc(), "2": *fc()},
		MapValidatesKeys:     map[string]Child{"1": *fc(), "2": *fc()}, MapPtValidatesKeys: map[string]*Child{"1": fc(), "2": fc()},
		PtMapValidatesKeys:    &map[string]Child{"1": *fc(), "2": *fc()},
		MapValidatesKeyValues: map[string]Child{"1": *fc(), "2": *fc()}, MapPtValidatesKeyValues: map[string]*Child{"1": fc(), "2": fc()},
		PtMapValidatesKeyValues: &map[string]Child{"1": *fc(), "2": *fc()},
		// validate none
		PrimitiveNoValidates: 1, BasicNoValidates: *fc(), PtNoValidates: fc(), AnyNoValidates: *fc(),
		SliceNoValidates: []Child{*fc(), *fc()}, SlicePtNoValidates: []*Child{fc(), fc()}, PtSliceNoValidates: &[]Child{*fc(), *fc()},
		MapNoValidates: map[string]Child{"1": *fc(), "2": *fc()}, MapPtNoValidates: map[string]*Child{"1": fc(), "2": fc()},
		PtMapNoValidates: &map[string]Child{"1": *fc(), "2": *fc()},
	}
}

type selfValidates struct {
	Primitive  int
	Primitive2 int
}

type unregistered struct{}

var testRegistry = &Registry{}

// validates each key-value pair's Key (present) and Value (registered Child)
var (
	childKeyValueRule   = newKeyValueRule[string, Child]([]Rule{presentRule{}}, []Rule{testRegistry.Backed()})
	childPtKeyValueRule = newKeyValueRule[string, *Child]([]Rule{presentRule{}}, []Rule{testRegistry.Backed()})
)

func init() {
	testRegistry.MustRegisterType(NewDefinition[parent]().Validates(RuleMap{
		"Child":     {presentRule{}, testRegistry.Backed()},
		"Primitive": {presentRule{}},
		"Basic":     {presentRule{}, testRegistry.Backed()},
		"Pt":        {presentRule{}, testRegistry.Backed()},
		"Any":       {presentRule{}},
		"Array":     {presentRule{}, Elems[[]Child](testRegistry.Backed())},
		"ArrayPt":   {presentRule{}, Elems[[]*Child](testRegistry.Backed())},
		"PtArray":   {presentRule{}, Elems[[]Child](testRegistry.Backed())},
		"Map":       {presentRule{}, Values[map[string]Child](testRegistry.Backed())},
		"MapPt":     {presentRule{}, Values[map[string]*Child](testRegistry.Backed())},
		"PtMap":     {presentRule{}, Values[map[string]Child](testRegistry.Backed())},

		"PrimitiveEmptyValidates": {},
		"BasicEmptyValidates":     {testRegistry.Backed()},
		"PtEmptyValidates":        {testRegistry.Backed()},
		"AnyEmptyValidates":       {},
		"SliceValidates":          {Elems[[]Child](testRegistry.Backed())},
		"SlicePtValidates":        {Elems[[]*Child](testRegistry.Backed())},
		"PtSliceValidates":        {Elems[[]Child](testRegistry.Backed())},
		"MapValidatesValues":      {Values[map[string]Child](testRegistry.Backed())},
		"MapPtValidatesValues":    {Values[map[string]*Child](testRegistry.Backed())},
		"PtMapValidatesValues":    {Values[map[string]Child](testRegistry.Backed())},
		"MapValidatesKeys":        {Keys[map[string]Child](presentRule{})},
		"MapPtValidatesKeys":      {Keys[map[string]*Child](presentRule{})},
		"PtMapValidatesKeys":      {Keys[map[string]Child](presentRule{})},
		"MapValidatesKeyValues":   {KeyValues[map[string]Child](childKeyValueRule)},
		"MapPtValidatesKeyValues": {KeyValues[map[string]*Child](childPtKeyValueRule)},
		"PtMapValidatesKeyValues": {KeyValues[map[string]Child](childKeyValueRule)},
	}))
	testRegistry.MustRegisterType(NewDefinition[Child]().Validates(RuleMap{
		"Validates": {presentRule{}},
	}))
	testRegistry.MustRegisterType(NewDefinition[selfValidates]().ValidatesSelf(presentRule{}))
}

type integrationTestCase struct {
	name    string
	isValid bool
	f       func() parent
	anyF    func() any
}

var integrationAnyTestCases = []integrationTestCase{
	//
	// Any
	//
	{name: "Data___int_raw", isValid: false, anyF: func() any {
		return 1
	}},
	{name: "Data___int_pt", isValid: false, anyF: func() any {
		i := 1
		return &i
	}},
	{name: "Data___unregistered_raw", isValid: false, anyF: func() any {
		return unregistered{}
	}},
	{name: "Data___unregistered_pt", isValid: false, anyF: func() any {
		return &unregistered{}
	}},
	{name: "Data___nil_raw", isValid: false, anyF: func() any {
		return nil
	}},
	{name: "Data___nil_pt", isValid: false, anyF: func() any {
		var i any
		return &i
	}},
	{name: "Data___selfValidates_full", isValid: true, anyF: func() any { return selfValidates{Primitive: 1, Primitive2: 2} }},
	{name: "Data___selfValidates_half_raw", isValid: true, anyF: func() any { return selfValidates{Primitive: 1} }},
	{name: "Data___selfValidates_half_pt", isValid: true, anyF: func() any { return &selfValidates{Primitive: 1} }},
	{name: "Data___selfValidates_empty_raw", isValid: false, anyF: func() any { return selfValidates{} }},
	{name: "Data___selfValidates_empty_pt", isValid: false, anyF: func() any {
		return &selfValidates{}
	}},
	{name: "Empty", isValid: false, f: func() parent {
		return parent{}
	}},
	{name: "Empty___any_raw", isValid: false, anyF: func() any {
		return parent{}
	}},
	{name: "Empty___any_pt", isValid: false, anyF: func() any {
		return &parent{}
	}},
	{name: "Full___any_raw", isValid: true, anyF: func() any {
		return fullParent()
	}},
	{name: "Full___any_pt", isValid: true, anyF: func() any {
		full := fullParent()
		return &full
	}},
}

func TestIntegration(t *testing.T) {
	integrationTestCases := make([]integrationTestCase, len(structValidatorTestCases))
	for i, v := range structValidatorTestCases {
		integrationTestCases[i] = integrationTestCase{
			name:    v.name,
			isValid: len(v.errorKeys) == 0,
			f:       v.f,
		}
	}

	for _, tc := range append(integrationAnyTestCases, integrationTestCases...) {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			if tc.f != nil {
				data := tc.f()
				require.Equal(tc.isValid, testRegistry.ValidateAny(data) == nil)
				require.Equal(tc.isValid, testRegistry.ValidateAny(&data) == nil)
				return
			}
			require.Equal(tc.isValid, testRegistry.ValidateAny(tc.anyF()) == nil)
		})
	}
}
