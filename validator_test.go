package firm

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type structValidatorTestCase struct {
	name      string
	f         func() parent
	errorKeys []string
	// invalidKeys keys Invalid errors for invalid values (e.g. nil pointer fields)
	invalidKeys []string
}

var structValidatorTestCases = []structValidatorTestCase{
	//
	// Full
	//
	{name: "Full", errorKeys: nil, f: fullParent},

	//
	// Embed
	//
	{name: "Embed___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.NoValidates = ""
		return changeParent
	}},
	{name: "Embed___child_validates_zero", errorKeys: []string{"Child.Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.Validates = ""
		return changeParent
	}},
	{name: "Embed___child_empty", errorKeys: []string{"Child", "Child.Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.Child = Child{}
		return changeParent
	}},

	//
	// Primitive
	//
	{name: "Primitive___zero", errorKeys: []string{"Primitive"}, f: func() parent {
		changeParent := fullParent()
		changeParent.Primitive = 0
		return changeParent
	}},

	//
	// Basic
	//
	{name: "Basic___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.Basic.NoValidates = ""
		return changeParent
	}},
	{name: "Basic___child_validates_zero", errorKeys: []string{"Basic.Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.Basic.Validates = ""
		return changeParent
	}},
	{name: "Basic___child_empty", errorKeys: []string{"Basic", "Basic.Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.Basic = Child{}
		return changeParent
	}},

	//
	// Pt
	//
	{name: "Pt___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.Pt.NoValidates = ""
		return changeParent
	}},
	{name: "Pt___child_validates_zero", errorKeys: []string{"Pt.Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.Pt.Validates = ""
		return changeParent
	}},
	{name: "Pt___child_empty", errorKeys: []string{"Pt", "Pt.Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.Pt = &Child{}
		return changeParent
	}},
	{name: "Pt___nil", invalidKeys: []string{"Pt"}, f: func() parent {
		changeParent := fullParent()
		changeParent.Pt = nil
		return changeParent
	}},

	//
	// Multi
	//
	{name: "Multi", errorKeys: []string{"Child", "Child.Validates", "Primitive"}, invalidKeys: []string{"Pt"}, f: func() parent {
		changeParent := fullParent()
		changeParent.Child = Child{}
		changeParent.Primitive = 0
		changeParent.Pt = nil
		return changeParent
	}},

	//
	// Any
	//
	{name: "Any___child_empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.Any = Child{}
		return changeParent
	}},
	{name: "Any___child_pointer_empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.Any = &Child{}
		return changeParent
	}},
	{name: "Any___nil", errorKeys: []string{"Any"}, f: func() parent {
		changeParent := fullParent()
		changeParent.Any = nil
		return changeParent
	}},

	//
	// Array
	//
	{name: "Array___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range changeParent.Array {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "Array___child_validates_zero", errorKeys: []string{"Array.[0].Validates", "Array.[1].Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.Array[0].Validates = ""
		changeParent.Array[1].Validates = ""
		return changeParent
	}},
	{name: "Array___child_validates_one_zero", errorKeys: []string{"Array.[0].Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.Array[0].Validates = ""
		return changeParent
	}},
	{name: "Array___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.Array = []Child{}
		return changeParent
	}},
	{name: "Array___nil", errorKeys: []string{"Array"}, f: func() parent {
		changeParent := fullParent()
		changeParent.Array = nil
		return changeParent
	}},

	//
	// ArrayPt
	//
	{name: "ArrayPt___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range changeParent.ArrayPt {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "ArrayPt___child_validates_zero", errorKeys: []string{"ArrayPt.[0].Validates", "ArrayPt.[1].Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.ArrayPt[0].Validates = ""
		changeParent.ArrayPt[1].Validates = ""
		return changeParent
	}},
	{name: "ArrayPt___child_validates_one_zero", errorKeys: []string{"ArrayPt.[0].Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.ArrayPt[0].Validates = ""
		return changeParent
	}},
	{name: "ArrayPt___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.ArrayPt = []*Child{}
		return changeParent
	}},
	{name: "ArrayPt___nil", errorKeys: []string{"ArrayPt"}, f: func() parent {
		changeParent := fullParent()
		changeParent.ArrayPt = nil
		return changeParent
	}},

	//
	// Map
	//
	{name: "Map___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range changeParent.Map {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "Map___child_validates_zero", errorKeys: []string{"Map.[1].Validates", "Map.[2].Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.Map["1"] = Child{}
		changeParent.Map["2"] = Child{}
		return changeParent
	}},
	{name: "Map___child_validates_one_zero", errorKeys: []string{"Map.[1].Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.Map["1"] = Child{}
		return changeParent
	}},
	{name: "Map___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.Map = map[string]Child{}
		return changeParent
	}},
	{name: "Map___nil", errorKeys: []string{"Map"}, f: func() parent {
		changeParent := fullParent()
		changeParent.Map = nil
		return changeParent
	}},

	//
	// MapPt
	//
	{name: "MapPt___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range changeParent.MapPt {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "MapPt___child_validates_zero", errorKeys: []string{"MapPt.[1].Validates", "MapPt.[2].Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.MapPt["1"] = &Child{}
		changeParent.MapPt["2"] = &Child{}
		return changeParent
	}},
	{name: "MapPt___child_validates_one_zero", errorKeys: []string{"MapPt.[1].Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.MapPt["1"] = &Child{}
		return changeParent
	}},
	{name: "MapPt___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapPt = map[string]*Child{}
		return changeParent
	}},
	{name: "MapPt___nil", errorKeys: []string{"MapPt"}, f: func() parent {
		changeParent := fullParent()
		changeParent.MapPt = nil
		return changeParent
	}},

	//
	// PrimitiveEmptyValidates
	//
	{name: "PrimitiveEmptyValidates___zero", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.PrimitiveEmptyValidates = 0
		return changeParent
	}},

	//
	// BasicEmptyValidates
	//
	{name: "BasicEmptyValidates___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.BasicEmptyValidates.NoValidates = ""
		return changeParent
	}},
	{name: "BasicEmptyValidates___child_validates_zero", errorKeys: []string{"BasicEmptyValidates.Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.BasicEmptyValidates.Validates = ""
		return changeParent
	}},
	{name: "BasicEmptyValidates____child_empty", errorKeys: []string{"BasicEmptyValidates.Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.BasicEmptyValidates = Child{}
		return changeParent
	}},

	//
	// PtEmptyValidates
	//
	{name: "PtEmptyValidates___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.PtEmptyValidates.NoValidates = ""
		return changeParent
	}},
	{name: "PtEmptyValidates___child_validates_zero", errorKeys: []string{"PtEmptyValidates.Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.PtEmptyValidates.Validates = ""
		return changeParent
	}},
	{name: "PtEmptyValidates___child_empty", errorKeys: []string{"PtEmptyValidates.Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.PtEmptyValidates = &Child{}
		return changeParent
	}},
	{name: "PtEmptyValidates___nil", invalidKeys: []string{"PtEmptyValidates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.PtEmptyValidates = nil
		return changeParent
	}},

	//
	// AnyEmptyValidates
	//
	{name: "AnyEmptyValidates___child_empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.AnyEmptyValidates = Child{}
		return changeParent
	}},
	{name: "AnyEmptyValidates___child_pointer_empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.AnyEmptyValidates = &Child{}
		return changeParent
	}},
	{name: "AnyEmptyValidates___nil", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.AnyEmptyValidates = nil
		return changeParent
	}},

	//
	// SliceValidates
	//
	{name: "SliceValidates___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range changeParent.SliceValidates {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "SliceValidates___child_validates_zero",
		errorKeys: []string{"SliceValidates.[0].Validates", "SliceValidates.[1].Validates"}, f: func() parent {
			changeParent := fullParent()
			changeParent.SliceValidates[0].Validates = ""
			changeParent.SliceValidates[1].Validates = ""
			return changeParent
		}},
	{name: "SliceValidates___child_validates_one_zero", errorKeys: []string{"SliceValidates.[0].Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.SliceValidates[0].Validates = ""
		return changeParent
	}},
	{name: "SliceValidates___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.SliceValidates = []Child{}
		return changeParent
	}},
	{name: "SliceValidates___nil", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.SliceValidates = nil
		return changeParent
	}},

	//
	// SlicePtValidates
	//
	{name: "SlicePtValidates___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range changeParent.SlicePtValidates {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "SlicePtValidates___child_validates_zero",
		errorKeys: []string{"SlicePtValidates.[0].Validates", "SlicePtValidates.[1].Validates"}, f: func() parent {
			changeParent := fullParent()
			changeParent.SlicePtValidates[0].Validates = ""
			changeParent.SlicePtValidates[1].Validates = ""
			return changeParent
		}},
	{name: "SlicePtValidates___child_validates_one_zero", errorKeys: []string{"SlicePtValidates.[0].Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.SlicePtValidates[0].Validates = ""
		return changeParent
	}},
	{name: "SlicePtValidates___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.SlicePtValidates = []*Child{}
		return changeParent
	}},
	{name: "SlicePtValidates___nil", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.SlicePtValidates = nil
		return changeParent
	}},
	{name: "SlicePtValidates___nil_element", errorKeys: []string{"SlicePtValidates.[0].Validates"}, invalidKeys: []string{"SlicePtValidates.[1]"},
		f: func() parent {
			changeParent := fullParent()
			// the empty child surfaces errors, while the nil element surfaces an Invalid error
			changeParent.SlicePtValidates = []*Child{{}, nil}
			return changeParent
		}},

	//
	// PtSliceValidates
	//
	{name: "PtSliceValidates___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range *changeParent.PtSliceValidates {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "PtSliceValidates___child_validates_zero",
		errorKeys: []string{"PtSliceValidates.[0].Validates", "PtSliceValidates.[1].Validates"}, f: func() parent {
			changeParent := fullParent()
			(*changeParent.PtSliceValidates)[0].Validates = ""
			(*changeParent.PtSliceValidates)[1].Validates = ""
			return changeParent
		}},
	{name: "PtSliceValidates___child_validates_one_zero", errorKeys: []string{"PtSliceValidates.[0].Validates"}, f: func() parent {
		changeParent := fullParent()
		(*changeParent.PtSliceValidates)[0].Validates = ""
		return changeParent
	}},
	{name: "PtSliceValidates___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.PtSliceValidates = &[]Child{}
		return changeParent
	}},
	{name: "PtSliceValidates___nil", invalidKeys: []string{"PtSliceValidates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.PtSliceValidates = nil
		return changeParent
	}},

	//
	// MapValidatesValues
	//
	{name: "MapValidatesValues___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range changeParent.MapValidatesValues {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "MapValidatesValues___child_validates_zero",
		errorKeys: []string{"MapValidatesValues.[1].Validates", "MapValidatesValues.[2].Validates"}, f: func() parent {
			changeParent := fullParent()
			changeParent.MapValidatesValues["1"] = Child{}
			changeParent.MapValidatesValues["2"] = Child{}
			return changeParent
		}},
	{name: "MapValidatesValues___child_validates_one_zero", errorKeys: []string{"MapValidatesValues.[1].Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.MapValidatesValues["1"] = Child{}
		return changeParent
	}},
	{name: "MapValidatesValues___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapValidatesValues = map[string]Child{}
		return changeParent
	}},
	{name: "MapValidatesValues___nil", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapValidatesValues = nil
		return changeParent
	}},

	//
	// MapPtValidatesValues
	//
	{name: "MapPtValidatesValues___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range changeParent.MapPtValidatesValues {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "MapPtValidatesValues___child_validates_zero",
		errorKeys: []string{"MapPtValidatesValues.[1].Validates", "MapPtValidatesValues.[2].Validates"}, f: func() parent {
			changeParent := fullParent()
			changeParent.MapPtValidatesValues["1"] = &Child{}
			changeParent.MapPtValidatesValues["2"] = &Child{}
			return changeParent
		}},
	{name: "MapPtValidatesValues___child_validates_one_zero", errorKeys: []string{"MapPtValidatesValues.[1].Validates"}, f: func() parent {
		changeParent := fullParent()
		changeParent.MapPtValidatesValues["1"] = &Child{}
		return changeParent
	}},
	{name: "MapPtValidatesValues___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapPtValidatesValues = map[string]*Child{}
		return changeParent
	}},
	{name: "MapPtValidatesValues___nil", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapPtValidatesValues = nil
		return changeParent
	}},
	{name: "MapPtValidatesValues___nil_element",
		errorKeys:   []string{"MapPtValidatesValues.[1].Validates"},
		invalidKeys: []string{"MapPtValidatesValues.[2]"},
		f: func() parent {
			changeParent := fullParent()
			// the empty child surfaces errors, while the nil element surfaces an Invalid error
			changeParent.MapPtValidatesValues = map[string]*Child{"1": {}, "2": nil}
			return changeParent
		}},

	//
	// PtMapValidatesValues
	//
	{name: "PtMapValidatesValues___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range *changeParent.PtMapValidatesValues {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "PtMapValidatesValues___child_validates_zero",
		errorKeys: []string{"PtMapValidatesValues.[1].Validates", "PtMapValidatesValues.[2].Validates"}, f: func() parent {
			changeParent := fullParent()
			m := *changeParent.PtMapValidatesValues
			m["1"] = Child{}
			m["2"] = Child{}
			changeParent.PtMapValidatesValues = &m
			return changeParent
		}},
	{name: "PtMapValidatesValues___child_validates_one_zero", errorKeys: []string{"PtMapValidatesValues.[1].Validates"}, f: func() parent {
		changeParent := fullParent()
		m := *changeParent.PtMapValidatesValues
		m["1"] = Child{}
		changeParent.PtMapValidatesValues = &m
		return changeParent
	}},
	{name: "PtMapValidatesValues___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.PtMapValidatesValues = &map[string]Child{}
		return changeParent
	}},
	{name: "PtMapValidatesValues___nil", invalidKeys: []string{"PtMapValidatesValues"}, f: func() parent {
		changeParent := fullParent()
		changeParent.PtMapValidatesValues = nil
		return changeParent
	}},

	//
	// MapValidatesKeys
	//
	{name: "MapValidatesKeys___key_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		// values are ignored
		changeParent.MapValidatesKeys["1"] = Child{}
		return changeParent
	}},
	{name: "MapValidatesKeys___key_zero", errorKeys: []string{"MapValidatesKeys.[]"}, f: func() parent {
		changeParent := fullParent()
		changeParent.MapValidatesKeys[""] = Child{}
		delete(changeParent.MapValidatesKeys, "1")
		return changeParent
	}},
	{name: "MapValidatesKeys___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapValidatesKeys = map[string]Child{}
		return changeParent
	}},
	{name: "MapValidatesKeys___nil", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapValidatesKeys = nil
		return changeParent
	}},

	//
	// MapPtValidatesKeys
	//
	{name: "MapPtValidatesKeys___key_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		// values are ignored
		changeParent.MapPtValidatesKeys["1"] = nil
		changeParent.MapPtValidatesKeys["2"] = nil
		return changeParent
	}},
	{name: "MapPtValidatesKeys___key_zero", errorKeys: []string{"MapPtValidatesKeys.[]"}, f: func() parent {
		changeParent := fullParent()
		changeParent.MapPtValidatesKeys[""] = nil
		delete(changeParent.MapPtValidatesKeys, "1")
		return changeParent
	}},
	{name: "MapPtValidatesKeys___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapPtValidatesKeys = map[string]*Child{}
		return changeParent
	}},
	{name: "MapPtValidatesKeys___nil", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapPtValidatesKeys = nil
		return changeParent
	}},

	//
	// PtMapValidatesKeys
	//
	{name: "PtMapValidatesKeys___key_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		// values are ignored
		m := *changeParent.PtMapValidatesKeys
		m["1"] = Child{}
		changeParent.PtMapValidatesKeys = &m
		return changeParent
	}},
	{name: "PtMapValidatesKeys___key_zero", errorKeys: []string{"PtMapValidatesKeys.[]"}, f: func() parent {
		changeParent := fullParent()
		m := *changeParent.PtMapValidatesKeys
		m[""] = Child{}
		delete(m, "1")
		changeParent.PtMapValidatesKeys = &m
		return changeParent
	}},
	{name: "PtMapValidatesKeys___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.PtMapValidatesKeys = &map[string]Child{}
		return changeParent
	}},
	{name: "PtMapValidatesKeys___nil", invalidKeys: []string{"PtMapValidatesKeys"}, f: func() parent {
		changeParent := fullParent()
		changeParent.PtMapValidatesKeys = nil
		return changeParent
	}},

	//
	// MapValidatesKeyValues
	//
	{name: "MapValidatesKeyValues___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range changeParent.MapValidatesKeyValues {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "MapValidatesKeyValues___child_validates_zero",
		errorKeys: []string{"MapValidatesKeyValues.[1].Value.Validates", "MapValidatesKeyValues.[2].Value.Validates"}, f: func() parent {
			changeParent := fullParent()
			changeParent.MapValidatesKeyValues["1"] = Child{}
			changeParent.MapValidatesKeyValues["2"] = Child{}
			return changeParent
		}},
	{name: "MapValidatesKeyValues___child_validates_one_zero",
		errorKeys: []string{"MapValidatesKeyValues.[1].Value.Validates"}, f: func() parent {
			changeParent := fullParent()
			changeParent.MapValidatesKeyValues["1"] = Child{}
			return changeParent
		}},
	{name: "MapValidatesKeyValues___key_zero", errorKeys: []string{"MapValidatesKeyValues.[].Key"}, f: func() parent {
		changeParent := fullParent()
		// move a valid child under an empty key
		changeParent.MapValidatesKeyValues[""] = changeParent.MapValidatesKeyValues["1"]
		delete(changeParent.MapValidatesKeyValues, "1")
		return changeParent
	}},
	{name: "MapValidatesKeyValues___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapValidatesKeyValues = map[string]Child{}
		return changeParent
	}},
	{name: "MapValidatesKeyValues___nil", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapValidatesKeyValues = nil
		return changeParent
	}},

	//
	// MapPtValidatesKeyValues
	//
	{name: "MapPtValidatesKeyValues___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range changeParent.MapPtValidatesKeyValues {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "MapPtValidatesKeyValues___child_validates_zero",
		errorKeys: []string{"MapPtValidatesKeyValues.[1].Value.Validates", "MapPtValidatesKeyValues.[2].Value.Validates"}, f: func() parent {
			changeParent := fullParent()
			changeParent.MapPtValidatesKeyValues["1"] = &Child{}
			changeParent.MapPtValidatesKeyValues["2"] = &Child{}
			return changeParent
		}},
	{name: "MapPtValidatesKeyValues___child_validates_one_zero",
		errorKeys: []string{"MapPtValidatesKeyValues.[1].Value.Validates"}, f: func() parent {
			changeParent := fullParent()
			changeParent.MapPtValidatesKeyValues["1"] = &Child{}
			return changeParent
		}},
	{name: "MapPtValidatesKeyValues___key_zero",
		errorKeys:   []string{"MapPtValidatesKeyValues.[].Key"},
		invalidKeys: []string{"MapPtValidatesKeyValues.[].Value"},
		f: func() parent {
			changeParent := fullParent()
			changeParent.MapPtValidatesKeyValues[""] = nil
			delete(changeParent.MapPtValidatesKeyValues, "1")
			return changeParent
		}},
	{name: "MapPtValidatesKeyValues___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapPtValidatesKeyValues = map[string]*Child{}
		return changeParent
	}},
	{name: "MapPtValidatesKeyValues___nil", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapPtValidatesKeyValues = nil
		return changeParent
	}},
	{name: "MapPtValidatesKeyValues___nil_element",
		errorKeys:   []string{"MapPtValidatesKeyValues.[1].Value.Validates"},
		invalidKeys: []string{"MapPtValidatesKeyValues.[2].Value"},
		f: func() parent {
			changeParent := fullParent()
			// the empty child surfaces errors, while the nil element surfaces an Invalid error
			changeParent.MapPtValidatesKeyValues = map[string]*Child{"1": {}, "2": nil}
			return changeParent
		}},

	//
	// PtMapValidatesKeyValues
	//
	{name: "PtMapValidatesKeyValues___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range *changeParent.PtMapValidatesKeyValues {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "PtMapValidatesKeyValues___child_validates_zero",
		errorKeys: []string{"PtMapValidatesKeyValues.[1].Value.Validates", "PtMapValidatesKeyValues.[2].Value.Validates"}, f: func() parent {
			changeParent := fullParent()
			m := *changeParent.PtMapValidatesKeyValues
			m["1"] = Child{}
			m["2"] = Child{}
			changeParent.PtMapValidatesKeyValues = &m
			return changeParent
		}},
	{name: "PtMapValidatesKeyValues___child_validates_one_zero",
		errorKeys: []string{"PtMapValidatesKeyValues.[1].Value.Validates"}, f: func() parent {
			changeParent := fullParent()
			m := *changeParent.PtMapValidatesKeyValues
			m["1"] = Child{}
			changeParent.PtMapValidatesKeyValues = &m
			return changeParent
		}},
	{name: "PtMapValidatesKeyValues___key_zero", errorKeys: []string{"PtMapValidatesKeyValues.[].Key"}, f: func() parent {
		changeParent := fullParent()
		m := *changeParent.PtMapValidatesKeyValues
		// move a valid child under an empty key
		m[""] = m["1"]
		delete(m, "1")
		changeParent.PtMapValidatesKeyValues = &m
		return changeParent
	}},
	{name: "PtMapValidatesKeyValues___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.PtMapValidatesKeyValues = &map[string]Child{}
		return changeParent
	}},
	{name: "PtMapValidatesKeyValues___nil", invalidKeys: []string{"PtMapValidatesKeyValues"}, f: func() parent {
		changeParent := fullParent()
		changeParent.PtMapValidatesKeyValues = nil
		return changeParent
	}},

	//
	// PrimitiveNoValidates
	//
	{name: "PrimitiveNoValidates___zero", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.PrimitiveNoValidates = 0
		return changeParent
	}},

	//
	// BasicNoValidates
	//
	{name: "BasicNoValidates___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.BasicNoValidates.NoValidates = ""
		return changeParent
	}},
	{name: "BasicNoValidates___child_validates_zero", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.BasicNoValidates.Validates = ""
		return changeParent
	}},
	{name: "BasicNoValidates____child_empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.BasicNoValidates = Child{}
		return changeParent
	}},

	//
	// PtNoValidates
	//
	{name: "PtNoValidates___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.PtNoValidates.NoValidates = ""
		return changeParent
	}},
	{name: "PtNoValidates___child_validates_zero", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.PtNoValidates.Validates = ""
		return changeParent
	}},
	{name: "PtNoValidates___child_empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.PtNoValidates = &Child{}
		return changeParent
	}},
	{name: "PtNoValidates___nil", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.PtNoValidates = nil
		return changeParent
	}},

	//
	// AnyNoValidates
	//
	{name: "AnyNoValidates___child_empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.AnyNoValidates = Child{}
		return changeParent
	}},
	{name: "AnyNoValidates___child_pointer_empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.AnyNoValidates = &Child{}
		return changeParent
	}},
	{name: "AnyNoValidates___nil", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.AnyNoValidates = nil
		return changeParent
	}},

	//
	// SliceNoValidates
	//
	{name: "SliceNoValidates___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range changeParent.SliceNoValidates {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "SliceNoValidates___child_validates_zero", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.SliceNoValidates[0].Validates = ""
		changeParent.SliceNoValidates[1].Validates = ""
		return changeParent
	}},
	{name: "SliceNoValidates___child_validates_one_zero", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.SliceNoValidates[0].Validates = ""
		return changeParent
	}},
	{name: "SliceNoValidates___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.SliceNoValidates = []Child{}
		return changeParent
	}},
	{name: "SliceNoValidates___nil", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.SliceNoValidates = nil
		return changeParent
	}},

	//
	// SlicePtNoValidates
	//
	{name: "SlicePtNoValidates___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range changeParent.SlicePtNoValidates {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "SlicePtNoValidates___child_validates_zero", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.SlicePtNoValidates[0].Validates = ""
		changeParent.SlicePtNoValidates[1].Validates = ""
		return changeParent
	}},
	{name: "SlicePtNoValidates___child_validates_one_zero", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.SlicePtNoValidates[0].Validates = ""
		return changeParent
	}},
	{name: "SlicePtNoValidates___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.SlicePtNoValidates = []*Child{}
		return changeParent
	}},
	{name: "SlicePtNoValidates___nil", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.SlicePtNoValidates = nil
		return changeParent
	}},

	//
	// MapNoValidates
	//
	{name: "MapNoValidates___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range changeParent.MapNoValidates {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "MapNoValidates___child_validates_zero", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapNoValidates["1"] = Child{}
		changeParent.MapNoValidates["2"] = Child{}
		return changeParent
	}},
	{name: "MapNoValidates___child_validates_one_zero", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapNoValidates["1"] = Child{}
		return changeParent
	}},
	{name: "MapNoValidates___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapNoValidates = map[string]Child{}
		return changeParent
	}},
	{name: "MapNoValidates___nil", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapNoValidates = nil
		return changeParent
	}},

	//
	// MapPtNoValidates
	//
	{name: "MapPtNoValidates___child_validates_ok", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		for _, v := range changeParent.MapPtNoValidates {
			v.NoValidates = ""
		}
		return changeParent
	}},
	{name: "MapPtNoValidates___child_validates_zero", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapPtNoValidates["1"] = &Child{}
		changeParent.MapPtNoValidates["2"] = &Child{}
		return changeParent
	}},
	{name: "MapPtNoValidates___child_validates_one_zero", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapPtNoValidates["1"] = &Child{}
		return changeParent
	}},
	{name: "MapPtNoValidates___empty", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapPtNoValidates = map[string]*Child{}
		return changeParent
	}},
	{name: "MapPtNoValidates___nil", errorKeys: nil, f: func() parent {
		changeParent := fullParent()
		changeParent.MapPtNoValidates = nil
		return changeParent
	}},
}

func TestFieldsAny(t *testing.T) {
	typ := reflect.TypeFor[Child]()
	ruleMap := RuleMap{"Validates": {presentRule{}}}

	expected, err := FieldsAnyWithErr(typ, ruleMap)
	require.NoError(t, err)
	require.Equal(t, expected, FieldsAny(typ, ruleMap))
	require.Equal(t, expected, FieldsAny(reflect.TypeFor[*Child](), ruleMap))

	require.Panics(t, func() { FieldsAny(reflect.TypeFor[int](), ruleMap) })
}

func TestFieldsAnyWithErr(t *testing.T) {
	noMatchingRule := onlyKindRule{kind: reflect.Bool}
	childPt := &Child{}
	doubleChildPt := &childPt

	tcs := []struct {
		name    string
		data    any
		ruleMap RuleMap
		err     error
	}{
		{name: "normal", data: Child{}, ruleMap: RuleMap{"Validates": {presentRule{}}}},
		{name: "pointer", data: &Child{}, ruleMap: RuleMap{"Validates": {presentRule{}}}},
		{name: "double_pointer", data: doubleChildPt, ruleMap: RuleMap{"Validates": {presentRule{}}}},
		{name: "non_exported_field", data: Child{}, ruleMap: RuleMap{"private": {presentRule{}}},
			err: errors.New("Fields: field, private, is unexported in type: firm.Child")},
		{name: "nil_type", data: nil, err: errors.New("Fields: type, nil, is not a Struct")},
		{name: "non_matching_field", data: Child{}, ruleMap: RuleMap{"No": {presentRule{}}},
			err: errors.New("Fields: field, No, not found in type: firm.Child")},
		{name: "no_matching_rule", data: Child{}, ruleMap: RuleMap{"Validates": {noMatchingRule}},
			err: fmt.Errorf("Fields: field, Validates, in firm.Child: %w", noMatchingRule.TypeCheck(reflect.TypeFor[string]()))},
		{name: "ptr_field_indirects_to_rule", data: parent{}, ruleMap: RuleMap{"Pt": {onlyKindRule{kind: reflect.Struct}}}},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			validator, err := FieldsAnyWithErr(reflect.TypeOf(tc.data), tc.ruleMap)
			if tc.err != nil {
				require.Equal(tc.err, err)
				return
			}

			require.NoError(err)
			require.Equal(indirectType(reflect.TypeOf(tc.data)), validator.typ)
			require.Len(validator.ruleMap, len(tc.ruleMap))
			for k, v := range tc.ruleMap {
				require.Equal(v, validator.ruleMap[k])
			}
		})
	}
}

func TestFieldsVldr_Validate(t *testing.T) {
	errorKey := ErrorKey("firm.Child.Validates." + presentRuleKey)

	tcs := []validateTC[Child]{
		{name: "valid", data: Child{Validates: "ok"}},
		{name: "invalid", data: Child{NoValidates: "not_ok"}, result: ErrorMap{errorKey: *presentRuleError(errorKey)}},
	}
	testValidate(t, tcs, func() (ValidatorTyped[Child], error) {
		return FieldsWithErr[Child](RuleMap{"Validates": []Rule{presentRule{}}})
	})
}

func TestFieldsVldr_Validate_PtrType(t *testing.T) {
	errorKey := ErrorKey("firm.Child.Validates." + presentRuleKey)
	child := Child{Validates: "ok"}

	tcs := []validateTC[*Child]{
		{name: "valid", data: &child},
		{name: "invalid", data: &Child{}, result: ErrorMap{errorKey: *presentRuleError(errorKey)}},
	}
	testValidate(t, tcs, func() (ValidatorTyped[*Child], error) {
		return FieldsWithErr[*Child](RuleMap{"Validates": []Rule{presentRule{}}})
	})
}

func TestFieldsAnyVldr_ValidateAll(t *testing.T) {
	validator := testRegistry.Validator(reflect.TypeFor[parent]())

	tcs := []struct {
		name   string
		data   any
		result ErrorMap
	}{
		{name: "not_struct", data: 1, result: typeCheckErrorResult(validator, 1)},
		{name: "invalid", data: nil, result: ErrInvalidValue()},
		{name: "nil_pointer", data: (*parent)(nil), result: ErrInvalidValue()},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) { require.Equal(t, tc.result, validator.ValidateAny(tc.data)) })
	}

	for _, tc := range structValidatorTestCases {
		t.Run(tc.name, func(t *testing.T) {
			rawData := tc.f()
			invalidKeySuffixes := joinAll(tc.invalidKeys, invalidKey)
			testValidateAllKeys(t, validator, rawData, joinAll(tc.errorKeys, presentRuleKey), invalidKeySuffixes)
			testValidateAllKeys(t, validator, &rawData, joinAll(tc.errorKeys, presentRuleKey), invalidKeySuffixes)
		})
	}
}

func TestFieldsAnyVldr_NilEmbeddedPointerField(t *testing.T) {
	require := require.New(t)

	validator, err := FieldsAnyWithErr(reflect.TypeFor[embeddedPtFields](), RuleMap{
		"Validates": {presentRule{}},
		"Str":       {presentRule{}},
	})
	require.NoError(err)

	// Child: nil is the embedded pointer field, so the Validates field's value is invalid and never reaches the rule
	expected := ErrorMap{}
	ErrInvalidValue().MergeInto("Validates", expected)
	expected = expected.Finish()
	require.Equal(expected, validator.ValidateValue(reflect.ValueOf(embeddedPtFields{Child: nil, Str: "ok"})))
	require.Nil(validator.ValidateValue(reflect.ValueOf(embeddedPtFields{Child: &Child{Validates: "ok"}, Str: "ok"})))
}

func TestFieldsAnyVldr_TypeCheck(t *testing.T) {
	validator, err := FieldsAnyWithErr(reflect.TypeFor[parent](), RuleMap{})
	require.NoError(t, err)
	badCondition := "is not matching Struct of type firm.parent"

	tcs := []struct {
		name         string
		data         any
		badCondition string
	}{
		{name: "matching struct", data: parent{}},
		{name: "matching struct pointer", data: &parent{}},
		{name: "other struct", data: Child{}, badCondition: badCondition},
		{name: "not struct", data: 1, badCondition: badCondition},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			testTypeCheck(t, tc.data, "FieldsAnyVldr", tc.badCondition, func() (Rule, error) {
				return validator, nil
			})
		})
	}
}

type embeddedPtFields struct {
	*Child

	Str string
}

type sliceValidatorElement struct {
	Int  int
	UInt uint
}

type sliceValidatorTestCase struct {
	name string
	f    func() any
	// defaults to elemsValidator
	validator Validator
	errorKeys []string
	// invalidKeys keys Invalid errors for invalid values (e.g. nil pointer elements)
	invalidKeys []string
}

var sliceValidatorTestCases = []sliceValidatorTestCase{
	{name: "Full", errorKeys: nil, f: func() any {
		return []sliceValidatorElement{{1, 1}, {2, 2}, {3, 3}}
	}},
	{name: "Empty", errorKeys: nil, f: func() any {
		return []sliceValidatorElement{}
	}},
	{name: "Nil", errorKeys: nil, f: func() any {
		return []sliceValidatorElement(nil)
	}},
	{name: "Element_Not_Full", errorKeys: nil, f: func() any {
		return []sliceValidatorElement{{Int: 1}, {Int: 2}}
	}},
	{name: "Element_Invalid", errorKeys: []string{"[0].Int", "[1].Int"}, f: func() any {
		return []sliceValidatorElement{{UInt: 1}, {UInt: 2}}
	}},
	{name: "Element_Empty", errorKeys: []string{"[0]", "[1]", "[0].Int", "[1].Int"}, f: func() any {
		return []sliceValidatorElement{{}, {}}
	}},

	//
	// Pointer elements
	//
	{name: "Ptr_Element_valid", validator: ptrElemsValidator, errorKeys: nil, f: func() any {
		return []*sliceValidatorElement{{Int: 1}, {Int: 2}}
	}},
	{name: "Ptr_Element_nil", validator: ptrElemsValidator,
		// the invalid value errors and never reaches the rules
		invalidKeys: []string{"[0]"}, f: func() any {
			return []*sliceValidatorElement{nil}
		}},
	{name: "Ptr_Element_nil_mixed", validator: ptrElemsValidator,
		errorKeys: []string{"[2]", "[2].Int"}, invalidKeys: []string{"[1]"}, f: func() any {
			return []*sliceValidatorElement{{Int: 1}, nil, {}}
		}},

	//
	// Pointer slice, double-pointer elements
	//
	{name: "Double_Ptr_Element_valid", validator: ptrPtElemsValidator, errorKeys: nil, f: func() any {
		return &[]**sliceValidatorElement{toElemPtrPtr(sliceValidatorElement{Int: 1})}
	}},
	{name: "Double_Ptr_Element_nil_mixed", validator: ptrPtElemsValidator,
		errorKeys: []string{"[2]", "[2].Int"}, invalidKeys: []string{"[1]"}, f: func() any {
			return &[]**sliceValidatorElement{toElemPtrPtr(sliceValidatorElement{Int: 1}), nil, toElemPtrPtr(sliceValidatorElement{})}
		}},
}

// toElemPtrPtr wraps the element into **sliceValidatorElement--composite literals can't be **T
func toElemPtrPtr(elem sliceValidatorElement) **sliceValidatorElement {
	elemPt := &elem
	return &elemPt
}

var elemsValidator = Elems[[]sliceValidatorElement](
	presentRule{}, Fields[sliceValidatorElement](RuleMap{"Int": {presentRule{}}}))

var ptrElemsValidator = Elems[[]*sliceValidatorElement](
	presentRule{}, Fields[sliceValidatorElement](RuleMap{"Int": {presentRule{}}}))

var ptrPtElemsValidator = ElemsAny(
	reflect.TypeFor[*[]**sliceValidatorElement](),
	presentRule{}, Fields[sliceValidatorElement](RuleMap{"Int": {presentRule{}}}))

func TestElemsAny(t *testing.T) {
	typ := reflect.TypeFor[[]Child]()
	expected, err := ElemsAnyWithErr(typ, presentRule{})
	require.NoError(t, err)
	require.Equal(t, expected, ElemsAny(typ, presentRule{}))
	require.Equal(t, expected, ElemsAny(reflect.TypeFor[*[]Child](), presentRule{}))

	require.Panics(t, func() { ElemsAny(reflect.TypeFor[Child](), presentRule{}) })
}

func TestElemsAnyWithErr(t *testing.T) {
	noMatchingRule := onlyKindRule{kind: reflect.Bool}
	slicePt := &[]Child{}
	doubleSlicePt := &slicePt
	ptrSlicePtrElems := &[]**Child{}
	doublePtrSlicePtrElems := &ptrSlicePtrElems

	tcs := []struct {
		name  string
		data  any
		rules []Rule
		err   error
	}{
		{name: "normal", data: []Child{}, rules: []Rule{presentRule{}}},
		{name: "slice_pointer", data: &[]Child{}, rules: []Rule{presentRule{}}},
		{name: "double_pointer", data: doubleSlicePt, rules: []Rule{presentRule{}}},
		// element type, *Child, is indirected before the rule's TypeCheck
		{name: "ptr_element_indirects_to_rule", data: []*Child{}, rules: []Rule{onlyKindRule{kind: reflect.Struct}}},
		{name: "ptr_slice_ptr_elements", data: &[]*Child{}, rules: []Rule{onlyKindRule{kind: reflect.Struct}}},
		{name: "double_ptr_slice_double_ptr_elements", data: doublePtrSlicePtrElems, rules: []Rule{onlyKindRule{kind: reflect.Struct}}},
		{name: "ptr_array_ptr_elements", data: &[2]*Child{}, rules: []Rule{onlyKindRule{kind: reflect.Struct}}},
		{name: "slice_pointer_with_onlyKindRule", data: &[]Child{}, rules: []Rule{onlyKindRule{kind: reflect.Struct}}},
		{name: "nil_type", data: nil, err: errors.New("Elems: type, nil, is not a Slice or Array")},
		{name: "not_slice", data: Child{}, err: errors.New("Elems: type, firm.Child, is not a Slice or Array")},
		{name: "pointer_to_not_slice", data: &Child{}, err: errors.New("Elems: type, firm.Child, is not a Slice or Array")},
		{name: "no_matching_rule", data: []Child{}, rules: []Rule{noMatchingRule},
			err: fmt.Errorf("Elems: element type: %w", noMatchingRule.TypeCheck(reflect.TypeFor[Child]()))},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			validator, err := ElemsAnyWithErr(reflect.TypeOf(tc.data), tc.rules...)
			if tc.err != nil {
				require.Equal(tc.err, err)
				return
			}

			require.NoError(err)
			require.Equal(indirectType(reflect.TypeOf(tc.data)), validator.typ)
			require.Equal(tc.rules, validator.elementRules)
		})
	}
}

func TestElemsVldr_Validate(t *testing.T) {
	errorKey := ErrorKey("[]firm.Child.[0]." + presentRuleKey)
	tcs := []validateTC[[]Child]{
		{name: "valid", data: []Child{{Validates: "ok"}}},
		{name: "invalid", data: []Child{{}}, result: ErrorMap{errorKey: *presentRuleError(errorKey)}},
	}
	testValidate(t, tcs, func() (ValidatorTyped[[]Child], error) {
		return ElemsWithErr[[]Child](presentRule{})
	})
}

func TestElemsAnyVldr_ValidateAll(t *testing.T) {
	validator := elemsValidator

	tcs := []struct {
		name   string
		data   any
		result ErrorMap
	}{
		{name: "not_slice", data: 1, result: typeCheckErrorResult(validator, 1)},
		{name: "invalid", data: nil, result: ErrInvalidValue()},
		{name: "nil_pointer", data: (*[]sliceValidatorElement)(nil), result: ErrInvalidValue()},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) { require.Equal(t, tc.result, validator.ValidateAny(tc.data)) })
	}

	for _, tc := range sliceValidatorTestCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := tc.validator
			if validator == nil {
				validator = elemsValidator
			}
			rawData := tc.f()
			// rawData comes boxed in an any, so &rawData would be a *any; build a typed pointer instead
			ptrData := reflect.New(reflect.TypeOf(rawData))
			ptrData.Elem().Set(reflect.ValueOf(rawData))
			testValidateAllKeys(t, validator, rawData,
				joinAll(tc.errorKeys, presentRuleKey), joinAll(tc.invalidKeys, invalidKey))
			testValidateAllKeys(t, validator, ptrData.Interface(),
				joinAll(tc.errorKeys, presentRuleKey), joinAll(tc.invalidKeys, invalidKey))
		})
	}
}

func TestElemsAnyVldr_TypeCheck(t *testing.T) {
	validator := elemsValidator
	badCondition := "is not matching Slice or Array of type []firm.sliceValidatorElement"

	tcs := []struct {
		name         string
		data         any
		badCondition string
	}{
		{name: "matching slice", data: []sliceValidatorElement{}},
		{name: "matching slice pointer", data: &[]sliceValidatorElement{}},
		{name: "other slice", data: []int{}, badCondition: badCondition},
		{name: "not slice", data: 1, badCondition: badCondition},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			testTypeCheck(t, tc.data, "ElemsAnyVldr", tc.badCondition, func() (Rule, error) {
				return validator, nil
			})
		})
	}
}

func TestValueAny(t *testing.T) {
	typ := reflect.TypeFor[int]()
	expected, err := ValueAnyWithErr(typ, presentRule{})
	require.NoError(t, err)
	require.Equal(t, expected, ValueAny(typ, presentRule{}))
	require.Equal(t, expected, ValueAny(reflect.TypeFor[*int](), presentRule{}))

	require.Panics(t, func() { ValueAny(reflect.TypeFor[[]int](), onlyKindRule{kind: reflect.Int}) })
}

func TestValueAnyWithErr(t *testing.T) {
	i := 0
	intPt := &i
	doubleIntPt := &intPt
	intRule := onlyKindRule{kind: reflect.Int}

	tcs := []struct {
		name  string
		data  any
		rules []Rule
		err   error
	}{
		{name: "normal", data: i, rules: []Rule{intRule}},
		{name: "int_pointer", data: intPt, rules: []Rule{intRule}},
		{name: "int_double_pointer", data: doubleIntPt, rules: []Rule{intRule}},
		{name: "nil_type", data: nil, err: errors.New("Value: type is nil")},
		{name: "not_int", data: []int{}, rules: []Rule{intRule}, err: intRule.TypeCheck(reflect.TypeFor[[]int]())},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			validator, err := ValueAnyWithErr(reflect.TypeOf(tc.data), tc.rules...)
			if tc.err != nil {
				require.Equal(tc.err, err)
				return
			}

			require.NoError(err)
			require.Equal(indirectType(reflect.TypeOf(tc.data)), validator.typ)
			require.Equal(tc.rules, validator.rules)
		})
	}
}

func TestValueVldr_Validate(t *testing.T) {
	errorKey := ErrorKey("firm.Child." + presentRuleKey)
	tcs := []validateTC[Child]{
		{name: "valid", data: Child{Validates: "ok"}},
		{name: "invalid", data: Child{}, result: ErrorMap{errorKey: *presentRuleError(errorKey)}},
	}
	testValidate(t, tcs, func() (ValidatorTyped[Child], error) {
		return ValueWithErr[Child](presentRule{})
	})
}

func TestValueVldr_Validate_PtrType(t *testing.T) {
	errorKey := ErrorKey("int." + presentRuleKey)
	i, zero := 1, 0

	tcs := []validateTC[*int]{
		{name: "valid", data: &i},
		{name: "invalid", data: &zero, result: ErrorMap{errorKey: *presentRuleError(errorKey)}},
	}
	testValidate(t, tcs, func() (ValidatorTyped[*int], error) {
		return ValueWithErr[*int](presentRule{})
	})
}

func TestValueAnyVldr_ValidateAll(t *testing.T) {
	edgeTcs := []struct {
		name string
		rule Rule
		data any

		newError       bool
		result         ErrorMap
		typeCheckError bool
	}{
		{name: "invalid", rule: presentRule{}, data: nil, result: ErrInvalidValue()},
		{name: "nil_pointer", rule: presentRule{}, data: (*bool)(nil), result: ErrInvalidValue()},
		{name: "bad_type_with_rule_validator", rule: onlyKindRule{kind: reflect.String}, data: 1, newError: true},
		{name: "bad_type_after_new", rule: onlyKindRule{kind: reflect.Bool}, data: 1, typeCheckError: true},
	}
	for _, tc := range edgeTcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			validator, err := ValueAnyWithErr(reflect.TypeFor[bool](), tc.rule)
			if tc.newError {
				require.Equal(NewRuleTypeError("onlyKindRule", reflect.TypeFor[bool](), "is not string"), err)
				return
			}

			require.NoError(err)
			result := tc.result
			if result == nil && tc.typeCheckError {
				result = typeCheckErrorResult(validator, tc.data)
			}
			require.Equal(result, validator.ValidateAny(tc.data))
		})
	}

	validator, err := ValueAnyWithErr(reflect.TypeFor[int](), presentRule{})
	require.NoError(t, err)
	type testCase struct {
		name string
		data any
		err  *TemplateError
	}
	tcs := []testCase{
		{name: "not_zero", data: 1},
		{name: "zero", data: 0, err: presentRuleError("")},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			testValidateAll(t, validator, tc.data, tc.err, presentRuleKey)
		})
	}
}

func TestValueAnyVldr_TypeCheck(t *testing.T) {
	i := 0

	tcs := []struct {
		name         string
		data         any
		extraRule    Rule
		badCondition string
	}{
		{name: "matching int", data: 0},
		{name: "matching int pointer", data: &i},
		{name: "not int", data: []int{}, badCondition: "is not matching type int"},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			rules := []Rule{presentRule{}, onlyKindRule{kind: reflect.Int}, presentRule{}}
			if tc.extraRule != nil {
				rules = append(rules, tc.extraRule)
			}
			testTypeCheck(t, tc.data, "ValueAnyVldr", tc.badCondition, func() (Rule, error) {
				return ValueAnyWithErr(reflect.TypeOf(i), rules...)
			})
		})
	}
}

func TestRuleVldr_ValidateAll(t *testing.T) {
	edgeTcs := []struct {
		name           string
		rule           Rule
		data           any
		result         ErrorMap
		typeCheckError bool
	}{
		{name: "invalid", rule: presentRule{}, data: nil, result: ErrInvalidValue()},
		{name: "nil_pointer", rule: presentRule{}, data: (*bool)(nil), result: ErrInvalidValue()},
		{name: "bad_type", rule: onlyKindRule{kind: reflect.Bool}, data: 1, typeCheckError: true},
	}
	for _, tc := range edgeTcs {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.result
			if result == nil && tc.typeCheckError {
				result = typeCheckErrorResult(tc.rule, tc.data)
			}

			validator := RuleVldr{Rule: tc.rule}
			require.Equal(t, result, validator.ValidateAny(tc.data))
		})
	}

	validator := RuleVldr{Rule: presentRule{}}
	type testCase struct {
		name string
		data any
		err  *TemplateError
	}
	tcs := []testCase{
		{name: "not_zero", data: 1},
		{name: "zero", data: 0, err: presentRuleError("")},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			testValidateAll(t, validator, tc.data, tc.err, presentRuleKey)
			testValidateAll(t, validator, &tc.data, nil)
		})
	}
}

func TestRuleGettersReturnCopies(t *testing.T) {
	fieldsV := Fields[sliceValidatorElement](RuleMap{"Int": {presentRule{}}})
	elemsV := Elems[[]sliceValidatorElement](presentRule{})
	valueV := Value[int](presentRule{})

	t.Run("fields_rule_map", func(t *testing.T) {
		ruleMap := fieldsV.RuleMap()
		require.Equal(t, []Rule{presentRule{}}, ruleMap["Int"])
		testRulesGetterIsolation(t, func() []Rule { return fieldsV.RuleMap()["Int"] })
	})
	t.Run("elems_element_rules", func(t *testing.T) {
		testRulesGetterIsolation(t, elemsV.ElementRules)
	})
	t.Run("value_rules", func(t *testing.T) {
		testRulesGetterIsolation(t, valueV.Rules)
	})
}

// TestInvalidValuePanics asserts invalid values into ValidateValue()/ValidateMerge() panic with
// safeValuePanic, instead of opaque reflect panics
func TestInvalidValuePanics(t *testing.T) {
	require := require.New(t)

	registry := &Registry{}
	require.NoError(registry.RegisterType(NewDefinition[Child]().ValidatesSelf(presentRule{})))
	backer := registry.Backed()

	t.Run("registry", func(*testing.T) {
		require.PanicsWithValue(safeValuePanic, func() { _ = registry.ValidateValue(reflect.Value{}) })
		require.PanicsWithValue(safeValuePanic, func() { registry.ValidateMerge(reflect.Value{}, "", ErrorMap{}) })
	})
	t.Run("registry_backer", func(*testing.T) {
		require.PanicsWithValue(safeValuePanic, func() { _ = backer.ValidateValue(reflect.Value{}) })
		require.PanicsWithValue(safeValuePanic, func() { backer.ValidateMerge(reflect.Value{}, "", ErrorMap{}) })
	})
	t.Run("validators", func(*testing.T) {
		validators := []Validator{
			Fields[Child](RuleMap{"Validates": {presentRule{}}}),
			Elems[[]Child](presentRule{}),
			Value[Child](presentRule{}),
			Keys[map[string]Child](presentRule{}),
			Values[map[string]Child](presentRule{}),
			KeyValues[map[string]Child](presentRule{}),
			RuleVldr{Rule: presentRule{}},
		}
		for _, validator := range validators {
			require.PanicsWithValue(safeValuePanic, func() { _ = validator.ValidateValue(reflect.Value{}) })
			require.PanicsWithValue(safeValuePanic, func() { validator.ValidateMerge(reflect.Value{}, "", ErrorMap{}) })
		}
	})
	t.Run("impl_helpers", func(*testing.T) {
		require.PanicsWithValue(safeValuePanic, func() { ImplValidateMerge(reflect.Value{}, "", ErrorMap{}, nil) })
	})
}
