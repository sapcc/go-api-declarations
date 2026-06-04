// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package opts_test

import (
	"testing"
	"time"

	. "go.xyrillian.de/gg/option"

	th "github.com/sapcc/go-api-declarations/internal/testhelper"
	"github.com/sapcc/go-api-declarations/opts"
)

func checkSerializingHappyPath(t *testing.T, variable string, input any, expectedQuery string) {
	t.Helper()
	t.Run(variable, func(t *testing.T) {
		t.Helper()
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()
		v, err := opts.BuildQueryString(input)
		if err != nil {
			t.Fatal(variable + ": " + err.Error())
		}
		th.CheckEquals(t, expectedQuery, v.Encode())
	})
}

func TestBuildQueryStringHappyPaths(t *testing.T) {
	// empty struct (all zero values omitted)
	checkSerializingHappyPath(t, "empty", testOpts{}, "")

	// embedded string
	checkSerializingHappyPath(t, "embedded_string", testOpts{EmbeddedOpts: EmbeddedOpts{EmbeddedString: "hello"}}, "embedded_string=hello")

	// map
	checkSerializingHappyPath(t, "string_map", testOpts{StringMap: map[string]string{"k1": "v1", "k2": "v2"}}, "string_map=k1%3Av1&string_map=k2%3Av2")
	checkSerializingHappyPath(t, "int_string_map", testOpts{IntStringMap: map[int]string{1: "foo", 2: "bar"}}, "int_string_map=1%3Afoo&int_string_map=2%3Abar")
	checkSerializingHappyPath(t, "string_int_map", testOpts{StringIntMap: map[string]int{"foo": 1, "bar": 2}}, "string_int_map=bar%3A2&string_int_map=foo%3A1")

	// time
	checkSerializingHappyPath(t, "time",
		testOpts{Time: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
		"time=2000-01-01T00%3A00%3A00Z")
	type testTimeRFCNanoOpts struct {
		Time time.Time `q:"time,format:RFC3339Nano"`
	}
	checkSerializingHappyPath(t, "time RFC3339Nano", testTimeRFCNanoOpts{Time: time.Date(2000, 1, 1, 0, 0, 0, 1, time.UTC)}, "time=2000-01-01T00%3A00%3A00.000000001Z")
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	checkSerializingHappyPath(t, "time RFC3339Nano", testTimeRFCNanoOpts{Time: time.Date(2000, 1, 1, 0, 0, 0, 1, loc)}, "time=2000-01-01T00%3A00%3A00.000000001%2B01%3A00")
	type testTimeRFCOpts struct {
		Time time.Time `q:"time,format:RFC3339"`
	}
	checkSerializingHappyPath(t, "time RFC3339", testTimeRFCOpts{Time: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}, "time=2000-01-01T00%3A00%3A00Z")
	type testTimeDateTimeOpts struct {
		Time time.Time `q:"time,format:DateTime"`
	}
	checkSerializingHappyPath(t, "time DateTime", testTimeDateTimeOpts{Time: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}, "time=2000-01-01+00%3A00%3A00")
	type testTimeDateOpts struct {
		Time time.Time `q:"time,format:DateOnly"`
	}
	checkSerializingHappyPath(t, "time Date", testTimeDateOpts{Time: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}, "time=2000-01-01")
	type testTimeUnixOpts struct {
		Time time.Time `q:"time,format:Unix"`
	}
	checkSerializingHappyPath(t, "time Date", testTimeUnixOpts{Time: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}, "time=946684800")

	// plain scalars
	checkSerializingHappyPath(t, "bool", testOpts{Bool: true}, "bool=true")
	checkSerializingHappyPath(t, "string", testOpts{String: "hello"}, "string=hello")
	checkSerializingHappyPath(t, "int", testOpts{Int: -42}, "int=-42")
	checkSerializingHappyPath(t, "int8", testOpts{Int8: -8}, "int8=-8")
	checkSerializingHappyPath(t, "int16", testOpts{Int16: -16}, "int16=-16")
	checkSerializingHappyPath(t, "int32", testOpts{Int32: -32}, "int32=-32")
	checkSerializingHappyPath(t, "int64", testOpts{Int64: -64}, "int64=-64")
	checkSerializingHappyPath(t, "uint", testOpts{Uint: 42}, "uint=42")
	checkSerializingHappyPath(t, "uint8", testOpts{Uint8: 8}, "uint8=8")
	checkSerializingHappyPath(t, "uint16", testOpts{Uint16: 16}, "uint16=16")
	checkSerializingHappyPath(t, "uint32", testOpts{Uint32: 32}, "uint32=32")
	checkSerializingHappyPath(t, "uint64", testOpts{Uint64: 64}, "uint64=64")
	checkSerializingHappyPath(t, "float32", testOpts{Float32: -1.5}, "float32=-1.5")
	checkSerializingHappyPath(t, "float64", testOpts{Float64: -2.5}, "float64=-2.5")

	// pointers
	checkSerializingHappyPath(t, "pointer_bool", testOpts{PointerBool: new(true)}, "pointer_bool=true")
	checkSerializingHappyPath(t, "pointer_time", testOpts{PointerTime: new(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))}, "pointer_time=2000-01-01T00%3A00%3A00Z")
	checkSerializingHappyPath(t, "pointer_string", testOpts{PointerString: new("world")}, "pointer_string=world")
	checkSerializingHappyPath(t, "pointer_int", testOpts{PointerInt: new(7)}, "pointer_int=7")
	checkSerializingHappyPath(t, "pointer_int8", testOpts{PointerInt8: new(int8(8))}, "pointer_int8=8")
	checkSerializingHappyPath(t, "pointer_int16", testOpts{PointerInt16: new(int16(16))}, "pointer_int16=16")
	checkSerializingHappyPath(t, "pointer_int32", testOpts{PointerInt32: new(int32(32))}, "pointer_int32=32")
	checkSerializingHappyPath(t, "pointer_int64", testOpts{PointerInt64: new(int64(64))}, "pointer_int64=64")
	checkSerializingHappyPath(t, "pointer_uint", testOpts{PointerUint: new(uint(7))}, "pointer_uint=7")
	checkSerializingHappyPath(t, "pointer_uint8", testOpts{PointerUint8: new(uint8(8))}, "pointer_uint8=8")
	checkSerializingHappyPath(t, "pointer_uint16", testOpts{PointerUint16: new(uint16(16))}, "pointer_uint16=16")
	checkSerializingHappyPath(t, "pointer_uint32", testOpts{PointerUint32: new(uint32(32))}, "pointer_uint32=32")
	checkSerializingHappyPath(t, "pointer_uint64", testOpts{PointerUint64: new(uint64(64))}, "pointer_uint64=64")
	checkSerializingHappyPath(t, "pointer_float32", testOpts{PointerFloat32: new(float32(3.14))}, "pointer_float32=3.14")
	checkSerializingHappyPath(t, "pointer_float64", testOpts{PointerFloat64: new(2.718)}, "pointer_float64=2.718")

	// slices
	checkSerializingHappyPath(t, "bool_slice", testOpts{BoolSlice: []bool{true, false, true}}, "bool_slice=true&bool_slice=false&bool_slice=true")
	checkSerializingHappyPath(t, "time_slice",
		testOpts{TimeSlice: []time.Time{
			time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
		}},
		"time_slice=2000-01-01T00%3A00%3A00Z&time_slice=2001-01-01T00%3A00%3A00Z")
	checkSerializingHappyPath(t, "string_slice", testOpts{StringSlice: []string{"a", "b", "c"}}, "string_slice=a&string_slice=b&string_slice=c")
	checkSerializingHappyPath(t, "int_slice", testOpts{IntSlice: []int{1, 2, 3}}, "int_slice=1&int_slice=2&int_slice=3")
	checkSerializingHappyPath(t, "int8_slice", testOpts{Int8Slice: []int8{1, 2, 3}}, "int8_slice=1&int8_slice=2&int8_slice=3")
	checkSerializingHappyPath(t, "int16_slice", testOpts{Int16Slice: []int16{1, 2, 3}}, "int16_slice=1&int16_slice=2&int16_slice=3")
	checkSerializingHappyPath(t, "int32_slice", testOpts{Int32Slice: []int32{1, 2, 3}}, "int32_slice=1&int32_slice=2&int32_slice=3")
	checkSerializingHappyPath(t, "int64_slice", testOpts{Int64Slice: []int64{1, 2, 3}}, "int64_slice=1&int64_slice=2&int64_slice=3")
	checkSerializingHappyPath(t, "uint_slice", testOpts{UintSlice: []uint{1, 2, 3}}, "uint_slice=1&uint_slice=2&uint_slice=3")
	checkSerializingHappyPath(t, "uint8_slice", testOpts{Uint8Slice: []uint8{1, 2, 3}}, "uint8_slice=1&uint8_slice=2&uint8_slice=3")
	checkSerializingHappyPath(t, "uint16_slice", testOpts{Uint16Slice: []uint16{1, 2, 3}}, "uint16_slice=1&uint16_slice=2&uint16_slice=3")
	checkSerializingHappyPath(t, "uint32_slice", testOpts{Uint32Slice: []uint32{1, 2, 3}}, "uint32_slice=1&uint32_slice=2&uint32_slice=3")
	checkSerializingHappyPath(t, "uint64_slice", testOpts{Uint64Slice: []uint64{1, 2, 3}}, "uint64_slice=1&uint64_slice=2&uint64_slice=3")
	checkSerializingHappyPath(t, "float32_slice", testOpts{Float32Slice: []float32{1.5, 2.5}}, "float32_slice=1.5&float32_slice=2.5")
	checkSerializingHappyPath(t, "float64_slice", testOpts{Float64Slice: []float64{1.5, 2.5}}, "float64_slice=1.5&float64_slice=2.5")

	// options
	checkSerializingHappyPath(t, "option_bool", testOpts{OptionBool: Some(true)}, "option_bool=true")
	checkSerializingHappyPath(t, "option_time",
		testOpts{OptionTime: Some(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))},
		"option_time=2000-01-01T00%3A00%3A00Z")
	checkSerializingHappyPath(t, "option_string", testOpts{OptionString: Some("hello")}, "option_string=hello")
	checkSerializingHappyPath(t, "option_int", testOpts{OptionInt: Some(42)}, "option_int=42")
	checkSerializingHappyPath(t, "option_int8", testOpts{OptionInt8: Some(int8(8))}, "option_int8=8")
	checkSerializingHappyPath(t, "option_int16", testOpts{OptionInt16: Some(int16(16))}, "option_int16=16")
	checkSerializingHappyPath(t, "option_int32", testOpts{OptionInt32: Some(int32(32))}, "option_int32=32")
	checkSerializingHappyPath(t, "option_int64", testOpts{OptionInt64: Some(int64(64))}, "option_int64=64")
	checkSerializingHappyPath(t, "option_uint", testOpts{OptionUint: Some(uint(42))}, "option_uint=42")
	checkSerializingHappyPath(t, "option_uint8", testOpts{OptionUint8: Some(uint8(8))}, "option_uint8=8")
	checkSerializingHappyPath(t, "option_uint16", testOpts{OptionUint16: Some(uint16(16))}, "option_uint16=16")
	checkSerializingHappyPath(t, "option_uint32", testOpts{OptionUint32: Some(uint32(32))}, "option_uint32=32")
	checkSerializingHappyPath(t, "option_uint64", testOpts{OptionUint64: Some(uint64(64))}, "option_uint64=64")
	checkSerializingHappyPath(t, "option_float32", testOpts{OptionFloat32: Some(float32(1.5))}, "option_float32=1.5")
	checkSerializingHappyPath(t, "option_float64", testOpts{OptionFloat64: Some(2.5)}, "option_float64=2.5")

	// multiple fields at once (url.Values.Encode sorts alphabetically)
	checkSerializingHappyPath(t, "multiple fields",
		testOpts{Bool: true, String: "hi", Int: 5, OptionString: Some("world")},
		"bool=true&int=5&option_string=world&string=hi")
}

func checkSerializingPanic(t *testing.T, panicMsg string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected panic %q, but function did not panic", panicMsg)
		}
		got, ok := r.(string)
		if !ok {
			t.Errorf("expected panic with string %q, but got non-string panic: %v", panicMsg, r)
		}
		if got != panicMsg {
			t.Errorf("expected panic: %s, but got: %s", panicMsg, got)
		}
	}()
	fn()
}

func TestBuildQueryStringErrors(t *testing.T) {
	// non-struct input (panics)
	checkSerializingPanic(t, "options type is not a struct", func() {
		opts.BuildQueryString(42) //nolint:errcheck // won't get to this part
	})

	// missing q-tag (panics)
	type testNonOpts struct {
		String string `yaml:"string"`
	}
	checkSerializingPanic(t, `expected "String" to have a "q:"-tag`, func() {
		opts.BuildQueryString(testNonOpts{}) //nolint:errcheck // won't get to this part
	})

	// embedded struct with q-tag (panics)
	type testEmbeddedQTagOpts struct {
		EmbeddedOpts `q:"embedded"`
	}
	checkSerializingPanic(t, `expected embedded struct "EmbeddedOpts" to have no "q:"-tag`, func() {
		opts.BuildQueryString(testEmbeddedQTagOpts{}) //nolint:errcheck // won't get to this part
	})

	// unknown struct parameter (panics)
	type testNested struct {
		String string `q:"string"`
	}
	type testNested2 struct {
		Nested testNested `q:"nested"`
	}
	checkSerializingPanic(t, "for structs only implementers of isZeroer are supported", func() {
		opts.BuildQueryString(testNested2{}) //nolint:errcheck // won't get to this part
	})

	// unknown time format (panics)
	type testBadTimeFormatOpts struct {
		Time time.Time `q:"time,format:foo"`
	}
	checkSerializingPanic(t, `unsupported time format "foo"; accepted: DateOnly, DateTime, RFC3339, RFC3339Nano, Unix`, func() {
		opts.BuildQueryString(testBadTimeFormatOpts{}) //nolint:errcheck // won't get to this part
	})
	type testMissingTimeFormatOpts struct {
		Time time.Time `q:"time"`
	}
	checkParsingPanic(t, `time format is missing for field "Time"`, func() {
		opts.BuildQueryString(testMissingTimeFormatOpts{}) //nolint:errcheck // won't get to this part
	})

	//  missing required parameter
	type requiredOpts struct {
		Name string `q:"name,required"`
	}
	_, err := opts.BuildQueryString(requiredOpts{})
	th.AssertErr(t, "required query parameter [Name] not set", err)
}
