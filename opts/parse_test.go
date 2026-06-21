// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package opts_test

import (
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.xyrillian.de/gg/assert"
	. "go.xyrillian.de/gg/option"

	"github.com/sapcc/go-api-declarations/opts"
)

type EmbeddedOpts struct {
	EmbeddedString string `q:"embedded_string"`
}

type testOpts struct {
	EmbeddedOpts
	StringMap         map[string]string `q:"string_map"`
	IntStringMap      map[int]string    `q:"int_string_map"`
	StringIntMap      map[string]int    `q:"string_int_map"`
	Bool              bool              `q:"bool"`
	Time              time.Time         `q:"time,format:RFC3339"`
	TimeUnix          time.Time         `q:"time_unix,format:Unix"`
	String            string            `q:"string"`
	Int               int               `q:"int"`
	Int8              int8              `q:"int8"`
	Int16             int16             `q:"int16"`
	Int32             int32             `q:"int32"`
	Int64             int64             `q:"int64"`
	Uint              uint              `q:"uint"`
	Uint8             uint8             `q:"uint8"`
	Uint16            uint16            `q:"uint16"`
	Uint32            uint32            `q:"uint32"`
	Uint64            uint64            `q:"uint64"`
	Float32           float32           `q:"float32"`
	Float64           float64           `q:"float64"`
	PointerBool       *bool             `q:"pointer_bool"`
	PointerTime       *time.Time        `q:"pointer_time,format:RFC3339"`
	PointerTimeUnix   *time.Time        `q:"pointer_time_unix,format:Unix"`
	PointerString     *string           `q:"pointer_string"`
	PointerInt        *int              `q:"pointer_int"`
	PointerInt8       *int8             `q:"pointer_int8"`
	PointerInt16      *int16            `q:"pointer_int16"`
	PointerInt32      *int32            `q:"pointer_int32"`
	PointerInt64      *int64            `q:"pointer_int64"`
	PointerUint       *uint             `q:"pointer_uint"`
	PointerUint8      *uint8            `q:"pointer_uint8"`
	PointerUint16     *uint16           `q:"pointer_uint16"`
	PointerUint32     *uint32           `q:"pointer_uint32"`
	PointerUint64     *uint64           `q:"pointer_uint64"`
	PointerFloat32    *float32          `q:"pointer_float32"`
	PointerFloat64    *float64          `q:"pointer_float64"`
	BoolSlice         []bool            `q:"bool_slice"`
	TimeSlice         []time.Time       `q:"time_slice,format:RFC3339"`
	TimeUnixSlice     []time.Time       `q:"time_unix_slice,format:Unix"`
	StringSlice       []string          `q:"string_slice"`
	IntSlice          []int             `q:"int_slice"`
	Int8Slice         []int8            `q:"int8_slice"`
	Int16Slice        []int16           `q:"int16_slice"`
	Int32Slice        []int32           `q:"int32_slice"`
	Int64Slice        []int64           `q:"int64_slice"`
	UintSlice         []uint            `q:"uint_slice"`
	Uint8Slice        []uint8           `q:"uint8_slice"`
	Uint16Slice       []uint16          `q:"uint16_slice"`
	Uint32Slice       []uint32          `q:"uint32_slice"`
	Uint64Slice       []uint64          `q:"uint64_slice"`
	Float32Slice      []float32         `q:"float32_slice"`
	Float64Slice      []float64         `q:"float64_slice"`
	OptionBool        Option[bool]      `q:"option_bool"`
	OptionTime        Option[time.Time] `q:"option_time,format:RFC3339"`
	OptionTimeUnix    Option[time.Time] `q:"option_time_unix,format:Unix"`
	OptionString      Option[string]    `q:"option_string"`
	OptionInt         Option[int]       `q:"option_int"`
	OptionInt8        Option[int8]      `q:"option_int8"`
	OptionInt16       Option[int16]     `q:"option_int16"`
	OptionInt32       Option[int32]     `q:"option_int32"`
	OptionInt64       Option[int64]     `q:"option_int64"`
	OptionUint        Option[uint]      `q:"option_uint"`
	OptionUint8       Option[uint8]     `q:"option_uint8"`
	OptionUint16      Option[uint16]    `q:"option_uint16"`
	OptionUint32      Option[uint32]    `q:"option_uint32"`
	OptionUint64      Option[uint64]    `q:"option_uint64"`
	OptionFloat32     Option[float32]   `q:"option_float32"`
	OptionFloat64     Option[float64]   `q:"option_float64"`
	WithDetails       bool              `q:"with,value:details"`
	WithSubresources  bool              `q:"with,value:subresources"`
	WithSubcapacities bool              `q:"with,value:subcapacities"`
}

func checkParsingHappyPath(t *testing.T, variable, query string, result testOpts) {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "/some/unimportant/path"+query, http.NoBody)
	to, err := opts.ParseQueryString[testOpts](r.URL.Query())
	if err != nil {
		t.Fatal(variable + ": " + err.Error())
	}
	assert.Equal(t, to, result)
}

func TestOptParserHappyPaths(t *testing.T) {
	// empty query
	checkParsingHappyPath(t, "empty opts", "", testOpts{})

	// embedded string
	checkParsingHappyPath(t, "embedded_string", "?embedded_string=hello",
		testOpts{EmbeddedOpts: EmbeddedOpts{EmbeddedString: "hello"}})

	// map
	checkParsingHappyPath(t, "string_map", "?string_map=k1:v1&string_map=k2:v2",
		testOpts{StringMap: map[string]string{"k1": "v1", "k2": "v2"}})
	checkParsingHappyPath(t, "int_string_map", "?int_string_map=1:foo&int_string_map=2:bar",
		testOpts{IntStringMap: map[int]string{1: "foo", 2: "bar"}})
	checkParsingHappyPath(t, "string_int_map", "?string_int_map=foo:1&string_int_map=bar:2",
		testOpts{StringIntMap: map[string]int{"foo": 1, "bar": 2}})

	// time
	checkParsingHappyPath(t, "time RFC3339", "?time=2000-01-01T00:00:00Z",
		testOpts{Time: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)})
	checkParsingHappyPath(t, "time Unix", "?time_unix=946684800",
		testOpts{TimeUnix: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)})

	// pointer time
	checkParsingHappyPath(t, "pointer time RFC3339", "?pointer_time=2000-01-01T00:00:00Z",
		testOpts{PointerTime: new(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))})
	checkParsingHappyPath(t, "pointer time Unix", "?pointer_time_unix=946684800",
		testOpts{PointerTimeUnix: new(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))})

	// time slice
	checkParsingHappyPath(t, "slice time RFC3339", "?time_slice=2000-01-01T00:00:00Z&time_slice=2001-01-01T00:00:00Z&time_slice=2002-01-01T00:00:00Z",
		testOpts{TimeSlice: []time.Time{
			time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2002, 1, 1, 0, 0, 0, 0, time.UTC),
		}})
	checkParsingHappyPath(t, "slice time Unix", "?time_unix_slice=946684800&time_unix_slice=978307200&time_unix_slice=1009843200",
		testOpts{TimeUnixSlice: []time.Time{
			time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2002, 1, 1, 0, 0, 0, 0, time.UTC),
		}})

	// option time
	checkParsingHappyPath(t, "option time RFC3339", "?option_time=2000-01-01T00:00:00Z",
		testOpts{OptionTime: Some(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))})
	checkParsingHappyPath(t, "option time Unix", "?option_time_unix=946684800",
		testOpts{OptionTimeUnix: Some(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))})

	// plain scalars
	checkParsingHappyPath(t, "bool", "?bool=true", testOpts{Bool: true})
	checkParsingHappyPath(t, "string", "?string=hello", testOpts{String: "hello"})
	checkParsingHappyPath(t, "int", "?int=42", testOpts{Int: 42})
	checkParsingHappyPath(t, "int8", "?int8=8", testOpts{Int8: 8})
	checkParsingHappyPath(t, "int16", "?int16=16", testOpts{Int16: 16})
	checkParsingHappyPath(t, "int32", "?int32=32", testOpts{Int32: 32})
	checkParsingHappyPath(t, "int64", "?int64=64", testOpts{Int64: 64})
	checkParsingHappyPath(t, "uint", "?uint=42", testOpts{Uint: 42})
	checkParsingHappyPath(t, "uint8", "?uint8=8", testOpts{Uint8: 8})
	checkParsingHappyPath(t, "uint16", "?uint16=16", testOpts{Uint16: 16})
	checkParsingHappyPath(t, "uint32", "?uint32=32", testOpts{Uint32: 32})
	checkParsingHappyPath(t, "uint64", "?uint64=64", testOpts{Uint64: 64})
	checkParsingHappyPath(t, "float32", "?float32=1.5", testOpts{Float32: 1.5})
	checkParsingHappyPath(t, "float64", "?float64=2.5", testOpts{Float64: 2.5})

	// pointer scalars
	checkParsingHappyPath(t, "pointer_bool", "?pointer_bool=true", testOpts{PointerBool: new(true)})
	checkParsingHappyPath(t, "pointer_string", "?pointer_string=world", testOpts{PointerString: new("world")})
	checkParsingHappyPath(t, "pointer_int", "?pointer_int=7", testOpts{PointerInt: new(7)})
	checkParsingHappyPath(t, "pointer_int8", "?pointer_int8=8", testOpts{PointerInt8: new(int8(8))})
	checkParsingHappyPath(t, "pointer_int16", "?pointer_int16=16", testOpts{PointerInt16: new(int16(16))})
	checkParsingHappyPath(t, "pointer_int32", "?pointer_int32=32", testOpts{PointerInt32: new(int32(32))})
	checkParsingHappyPath(t, "pointer_int64", "?pointer_int64=64", testOpts{PointerInt64: new(int64(64))})
	checkParsingHappyPath(t, "pointer_uint", "?pointer_uint=7", testOpts{PointerUint: new(uint(7))})
	checkParsingHappyPath(t, "pointer_uint8", "?pointer_uint8=8", testOpts{PointerUint8: new(uint8(8))})
	checkParsingHappyPath(t, "pointer_uint16", "?pointer_uint16=16", testOpts{PointerUint16: new(uint16(16))})
	checkParsingHappyPath(t, "pointer_uint32", "?pointer_uint32=32", testOpts{PointerUint32: new(uint32(32))})
	checkParsingHappyPath(t, "pointer_uint64", "?pointer_uint64=64", testOpts{PointerUint64: new(uint64(64))})
	checkParsingHappyPath(t, "pointer_float32", "?pointer_float32=3.14", testOpts{PointerFloat32: new(float32(3.14))})
	checkParsingHappyPath(t, "pointer_float64", "?pointer_float64=2.718", testOpts{PointerFloat64: new(2.718)})

	// slices
	checkParsingHappyPath(t, "bool_slice", "?bool_slice=true&bool_slice=false&bool_slice=true",
		testOpts{BoolSlice: []bool{true, false, true}})
	checkParsingHappyPath(t, "string_slice", "?string_slice=a&string_slice=b&string_slice=c",
		testOpts{StringSlice: []string{"a", "b", "c"}})
	checkParsingHappyPath(t, "int_slice", "?int_slice=1&int_slice=2&int_slice=3",
		testOpts{IntSlice: []int{1, 2, 3}})
	checkParsingHappyPath(t, "int8_slice", "?int8_slice=1&int8_slice=2&int8_slice=3",
		testOpts{Int8Slice: []int8{1, 2, 3}})
	checkParsingHappyPath(t, "int16_slice", "?int16_slice=1&int16_slice=2&int16_slice=3",
		testOpts{Int16Slice: []int16{1, 2, 3}})
	checkParsingHappyPath(t, "int32_slice", "?int32_slice=1&int32_slice=2&int32_slice=3",
		testOpts{Int32Slice: []int32{1, 2, 3}})
	checkParsingHappyPath(t, "int64_slice", "?int64_slice=1&int64_slice=2&int64_slice=3",
		testOpts{Int64Slice: []int64{1, 2, 3}})
	checkParsingHappyPath(t, "uint_slice", "?uint_slice=1&uint_slice=2&uint_slice=3",
		testOpts{UintSlice: []uint{1, 2, 3}})
	checkParsingHappyPath(t, "uint8_slice", "?uint8_slice=1&uint8_slice=2&uint8_slice=3",
		testOpts{Uint8Slice: []uint8{1, 2, 3}})
	checkParsingHappyPath(t, "uint16_slice", "?uint16_slice=1&uint16_slice=2&uint16_slice=3",
		testOpts{Uint16Slice: []uint16{1, 2, 3}})
	checkParsingHappyPath(t, "uint32_slice", "?uint32_slice=1&uint32_slice=2&uint32_slice=3",
		testOpts{Uint32Slice: []uint32{1, 2, 3}})
	checkParsingHappyPath(t, "uint64_slice", "?uint64_slice=1&uint64_slice=2&uint64_slice=3",
		testOpts{Uint64Slice: []uint64{1, 2, 3}})
	checkParsingHappyPath(t, "float32_slice", "?float32_slice=1.5&float32_slice=2.5",
		testOpts{Float32Slice: []float32{1.5, 2.5}})
	checkParsingHappyPath(t, "float64_slice", "?float64_slice=1.5&float64_slice=2.5",
		testOpts{Float64Slice: []float64{1.5, 2.5}})

	// Option scalars
	checkParsingHappyPath(t, "option_bool", "?option_bool=true", testOpts{OptionBool: Some(true)})
	checkParsingHappyPath(t, "option_string", "?option_string=hello", testOpts{OptionString: Some("hello")})
	checkParsingHappyPath(t, "option_int", "?option_int=42", testOpts{OptionInt: Some(42)})
	checkParsingHappyPath(t, "option_int8", "?option_int8=8", testOpts{OptionInt8: Some(int8(8))})
	checkParsingHappyPath(t, "option_int16", "?option_int16=16", testOpts{OptionInt16: Some(int16(16))})
	checkParsingHappyPath(t, "option_int32", "?option_int32=32", testOpts{OptionInt32: Some(int32(32))})
	checkParsingHappyPath(t, "option_int64", "?option_int64=64", testOpts{OptionInt64: Some(int64(64))})
	checkParsingHappyPath(t, "option_uint", "?option_uint=42", testOpts{OptionUint: Some(uint(42))})
	checkParsingHappyPath(t, "option_uint8", "?option_uint8=8", testOpts{OptionUint8: Some(uint8(8))})
	checkParsingHappyPath(t, "option_uint16", "?option_uint16=16", testOpts{OptionUint16: Some(uint16(16))})
	checkParsingHappyPath(t, "option_uint32", "?option_uint32=32", testOpts{OptionUint32: Some(uint32(32))})
	checkParsingHappyPath(t, "option_uint64", "?option_uint64=64", testOpts{OptionUint64: Some(uint64(64))})
	checkParsingHappyPath(t, "option_float32", "?option_float32=1.5", testOpts{OptionFloat32: Some(float32(1.5))})
	checkParsingHappyPath(t, "option_float64", "?option_float64=2.5", testOpts{OptionFloat64: Some(2.5)})

	// multiple fields at once
	checkParsingHappyPath(t, "multiple fields", "?bool=true&string=hi&int=5&option_string=world",
		testOpts{Bool: true, String: "hi", Int: 5, OptionString: Some("world")})

	// value-discriminant bools
	checkParsingHappyPath(t, "value-discriminant bools", "?with=details&with=subcapacities",
		testOpts{WithDetails: true, WithSubresources: false, WithSubcapacities: true})
}

func checkParsingError(t *testing.T, query, errMsg string) {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "/some/unimportant/path"+query, http.NoBody)
	_, resultingError := opts.ParseQueryString[testOpts](r.URL.Query())
	assert.ErrEqual(t, resultingError, errMsg)
}

func TestOptParserErrors(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/some/unimportant/path", http.NoBody)

	// non-struct type parameter (panics)
	expectPanic(t, "options type is not a struct", func() {
		opts.ParseQueryString[int](r.URL.Query()) //nolint:errcheck // won't get to this part
	})

	// unknown value in flagset
	checkParsingError(t, "?with=details&with=nonsense", `unknown value "nonsense" for query parameter "with"`)

	// missing required parameter
	type testStringRequiredOpts struct {
		String string `q:"string,required"`
	}
	_, resultingError := opts.ParseQueryString[testStringRequiredOpts](r.URL.Query())
	assert.ErrEqual(t, resultingError, `missing value for query parameter "string"`)
	type testStringSliceRequiredOpts struct {
		StringSlice []string `q:"string_slice,required"`
	}
	_, resultingError = opts.ParseQueryString[testStringSliceRequiredOpts](r.URL.Query())
	assert.ErrEqual(t, resultingError, `missing value for query parameter "string_slice"`)

	// unknown parameter
	checkParsingError(t, "?someRandomParam=foo", `unknown query parameter "someRandomParam"`)

	// wrong type: map
	checkParsingError(t, "?string_map=foo", `invalid value for query parameter "string_map": invalid map entry "foo": expected key:value`)

	// wrong type: time
	checkParsingError(t, "?time=foo", `invalid value for query parameter "time": cannot parse "foo" as RFC3339: parsing time "foo" as "2006-01-02T15:04:05Z07:00": cannot parse "foo" as "2006"`)
	checkParsingError(t, "?pointer_time=foo", `invalid value for query parameter "pointer_time": cannot parse "foo" as RFC3339: parsing time "foo" as "2006-01-02T15:04:05Z07:00": cannot parse "foo" as "2006"`)
	checkParsingError(t, "?time_slice=foo", `invalid value for query parameter "time_slice": element 0: cannot parse "foo" as RFC3339: parsing time "foo" as "2006-01-02T15:04:05Z07:00": cannot parse "foo" as "2006"`)
	checkParsingError(t, "?option_time=foo", `invalid value for query parameter "option_time": cannot parse "foo" as RFC3339: parsing time "foo" as "2006-01-02T15:04:05Z07:00": cannot parse "foo" as "2006"`)

	checkParsingError(t, "?time_unix=foo", `invalid value for query parameter "time_unix": cannot parse "foo" as Unix seconds: strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?pointer_time_unix=foo", `invalid value for query parameter "pointer_time_unix": cannot parse "foo" as Unix seconds: strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?time_unix_slice=foo", `invalid value for query parameter "time_unix_slice": element 0: cannot parse "foo" as Unix seconds: strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?option_time_unix=foo", `invalid value for query parameter "option_time_unix": cannot parse "foo" as Unix seconds: strconv.ParseInt: parsing "foo": invalid syntax`)

	// multiple values: time
	checkParsingError(t, "?time=2000-01-01&time=2001-01-01", `invalid value for query parameter "time": expected a single value, got 2`)
	checkParsingError(t, "?pointer_time=2000-01-01&pointer_time=2001-01-01", `invalid value for query parameter "pointer_time": expected a single value, got 2`)
	checkParsingError(t, "?option_time=2000-01-01&option_time=2001-01-01", `invalid value for query parameter "option_time": expected a single value, got 2`)

	// wrong type: bool
	checkParsingError(t, "?bool=foo", `invalid value for query parameter "bool": strconv.ParseBool: parsing "foo": invalid syntax`)
	checkParsingError(t, "?pointer_bool=foo", `invalid value for query parameter "pointer_bool": strconv.ParseBool: parsing "foo": invalid syntax`)
	checkParsingError(t, "?option_bool=foo", `invalid value for query parameter "option_bool": strconv.ParseBool: parsing "foo": invalid syntax`)
	checkParsingError(t, "?bool_slice=foo", `invalid value for query parameter "bool_slice": element 0: strconv.ParseBool: parsing "foo": invalid syntax`)

	// wrong time format
	checkParsingError(t, "?time=2000-01-01", `invalid value for query parameter "time": cannot parse "2000-01-01" as RFC3339: parsing time "2000-01-01" as "2006-01-02T15:04:05Z07:00": cannot parse "" as "T"`)
	r = httptest.NewRequest(http.MethodGet, "/some/unimportant/path?time=2000-01-01%2000:00", http.NoBody)

	// wrong type: numbers
	checkParsingError(t, "?int=foo", `invalid value for query parameter "int": strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?int8=foo", `invalid value for query parameter "int8": strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?int16=foo", `invalid value for query parameter "int16": strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?int32=foo", `invalid value for query parameter "int32": strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?int64=foo", `invalid value for query parameter "int64": strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?uint=foo", `invalid value for query parameter "uint": strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?uint8=foo", `invalid value for query parameter "uint8": strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?uint16=foo", `invalid value for query parameter "uint16": strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?uint32=foo", `invalid value for query parameter "uint32": strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?uint64=foo", `invalid value for query parameter "uint64": strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?float32=foo", `invalid value for query parameter "float32": strconv.ParseFloat: parsing "foo": invalid syntax`)
	checkParsingError(t, "?float64=foo", `invalid value for query parameter "float64": strconv.ParseFloat: parsing "foo": invalid syntax`)
	checkParsingError(t, "?int_slice=1&int_slice=foo", `invalid value for query parameter "int_slice": element 1: strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?int_string_map=foo:bar", `invalid value for query parameter "int_string_map": invalid map key "foo": strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?string_int_map=foo:bar", `invalid value for query parameter "string_int_map": invalid map value "bar": strconv.ParseInt: parsing "bar": invalid syntax`)

	// wrong type: pointer of numbers
	checkParsingError(t, "?pointer_int=foo", `invalid value for query parameter "pointer_int": strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?pointer_int8=foo", `invalid value for query parameter "pointer_int8": strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?pointer_int16=foo", `invalid value for query parameter "pointer_int16": strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?pointer_int32=foo", `invalid value for query parameter "pointer_int32": strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?pointer_int64=foo", `invalid value for query parameter "pointer_int64": strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?pointer_uint=foo", `invalid value for query parameter "pointer_uint": strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?pointer_uint8=foo", `invalid value for query parameter "pointer_uint8": strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?pointer_uint16=foo", `invalid value for query parameter "pointer_uint16": strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?pointer_uint32=foo", `invalid value for query parameter "pointer_uint32": strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?pointer_uint64=foo", `invalid value for query parameter "pointer_uint64": strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?pointer_float32=foo", `invalid value for query parameter "pointer_float32": strconv.ParseFloat: parsing "foo": invalid syntax`)
	checkParsingError(t, "?pointer_float64=foo", `invalid value for query parameter "pointer_float64": strconv.ParseFloat: parsing "foo": invalid syntax`)

	// wrong type: slice of numbers
	checkParsingError(t, "?int_slice=foo", `invalid value for query parameter "int_slice": element 0: strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?int8_slice=foo", `invalid value for query parameter "int8_slice": element 0: strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?int16_slice=foo", `invalid value for query parameter "int16_slice": element 0: strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?int32_slice=foo", `invalid value for query parameter "int32_slice": element 0: strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?int64_slice=foo", `invalid value for query parameter "int64_slice": element 0: strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?uint_slice=foo", `invalid value for query parameter "uint_slice": element 0: strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?uint8_slice=foo", `invalid value for query parameter "uint8_slice": element 0: strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?uint16_slice=foo", `invalid value for query parameter "uint16_slice": element 0: strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?uint32_slice=foo", `invalid value for query parameter "uint32_slice": element 0: strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?uint64_slice=foo", `invalid value for query parameter "uint64_slice": element 0: strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?float32_slice=foo", `invalid value for query parameter "float32_slice": element 0: strconv.ParseFloat: parsing "foo": invalid syntax`)
	checkParsingError(t, "?float64_slice=foo", `invalid value for query parameter "float64_slice": element 0: strconv.ParseFloat: parsing "foo": invalid syntax`)

	// wrong type: Option of numbers
	checkParsingError(t, "?option_int=foo", `invalid value for query parameter "option_int": strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?option_int8=foo", `invalid value for query parameter "option_int8": strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?option_int16=foo", `invalid value for query parameter "option_int16": strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?option_int32=foo", `invalid value for query parameter "option_int32": strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?option_int64=foo", `invalid value for query parameter "option_int64": strconv.ParseInt: parsing "foo": invalid syntax`)
	checkParsingError(t, "?option_uint=foo", `invalid value for query parameter "option_uint": strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?option_uint8=foo", `invalid value for query parameter "option_uint8": strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?option_uint16=foo", `invalid value for query parameter "option_uint16": strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?option_uint32=foo", `invalid value for query parameter "option_uint32": strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?option_uint64=foo", `invalid value for query parameter "option_uint64": strconv.ParseUint: parsing "foo": invalid syntax`)
	checkParsingError(t, "?option_float32=foo", `invalid value for query parameter "option_float32": strconv.ParseFloat: parsing "foo": invalid syntax`)
	checkParsingError(t, "?option_float64=foo", `invalid value for query parameter "option_float64": strconv.ParseFloat: parsing "foo": invalid syntax`)

	// multiple values: scalar
	checkParsingError(t, "?bool=true&bool=false", `invalid value for query parameter "bool": expected a single value, got 2`)
	checkParsingError(t, "?string=foo&string=bar", `invalid value for query parameter "string": expected a single value, got 2`)
	checkParsingError(t, "?int=foo&int=bar", `invalid value for query parameter "int": expected a single value, got 2`)
	checkParsingError(t, "?int8=foo&int8=bar", `invalid value for query parameter "int8": expected a single value, got 2`)
	checkParsingError(t, "?int16=foo&int16=bar", `invalid value for query parameter "int16": expected a single value, got 2`)
	checkParsingError(t, "?int32=foo&int32=bar", `invalid value for query parameter "int32": expected a single value, got 2`)
	checkParsingError(t, "?int64=foo&int64=bar", `invalid value for query parameter "int64": expected a single value, got 2`)
	checkParsingError(t, "?uint=foo&uint=bar", `invalid value for query parameter "uint": expected a single value, got 2`)
	checkParsingError(t, "?uint8=foo&uint8=bar", `invalid value for query parameter "uint8": expected a single value, got 2`)
	checkParsingError(t, "?uint16=foo&uint16=bar", `invalid value for query parameter "uint16": expected a single value, got 2`)
	checkParsingError(t, "?uint32=foo&uint32=bar", `invalid value for query parameter "uint32": expected a single value, got 2`)
	checkParsingError(t, "?uint64=foo&uint64=bar", `invalid value for query parameter "uint64": expected a single value, got 2`)
	checkParsingError(t, "?float32=foo&float32=bar", `invalid value for query parameter "float32": expected a single value, got 2`)
	checkParsingError(t, "?float64=foo&float64=bar", `invalid value for query parameter "float64": expected a single value, got 2`)

	// multiple vales: pointer of scalar
	checkParsingError(t, "?pointer_bool=true&pointer_bool=false", `invalid value for query parameter "pointer_bool": expected a single value, got 2`)
	checkParsingError(t, "?pointer_string=foo&pointer_string=bar", `invalid value for query parameter "pointer_string": expected a single value, got 2`)
	checkParsingError(t, "?pointer_int=foo&pointer_int=bar", `invalid value for query parameter "pointer_int": expected a single value, got 2`)
	checkParsingError(t, "?pointer_int8=foo&pointer_int8=bar", `invalid value for query parameter "pointer_int8": expected a single value, got 2`)
	checkParsingError(t, "?pointer_int16=foo&pointer_int16=bar", `invalid value for query parameter "pointer_int16": expected a single value, got 2`)
	checkParsingError(t, "?pointer_int32=foo&pointer_int32=bar", `invalid value for query parameter "pointer_int32": expected a single value, got 2`)
	checkParsingError(t, "?pointer_int64=foo&pointer_int64=bar", `invalid value for query parameter "pointer_int64": expected a single value, got 2`)
	checkParsingError(t, "?pointer_uint=foo&pointer_uint=bar", `invalid value for query parameter "pointer_uint": expected a single value, got 2`)
	checkParsingError(t, "?pointer_uint8=foo&pointer_uint8=bar", `invalid value for query parameter "pointer_uint8": expected a single value, got 2`)
	checkParsingError(t, "?pointer_uint16=foo&pointer_uint16=bar", `invalid value for query parameter "pointer_uint16": expected a single value, got 2`)
	checkParsingError(t, "?pointer_uint32=foo&pointer_uint32=bar", `invalid value for query parameter "pointer_uint32": expected a single value, got 2`)
	checkParsingError(t, "?pointer_uint64=foo&pointer_uint64=bar", `invalid value for query parameter "pointer_uint64": expected a single value, got 2`)
	checkParsingError(t, "?pointer_float32=foo&pointer_float32=bar", `invalid value for query parameter "pointer_float32": expected a single value, got 2`)
	checkParsingError(t, "?pointer_float64=foo&pointer_float64=bar", `invalid value for query parameter "pointer_float64": expected a single value, got 2`)

	// multiple values: Option of scalar
	checkParsingError(t, "?option_bool=true&option_bool=false", `invalid value for query parameter "option_bool": expected a single value, got 2`)
	checkParsingError(t, "?option_string=foo&option_string=bar", `invalid value for query parameter "option_string": expected a single value, got 2`)
	checkParsingError(t, "?option_int=foo&option_int=bar", `invalid value for query parameter "option_int": expected a single value, got 2`)
	checkParsingError(t, "?option_int8=foo&option_int8=bar", `invalid value for query parameter "option_int8": expected a single value, got 2`)
	checkParsingError(t, "?option_int16=foo&option_int16=bar", `invalid value for query parameter "option_int16": expected a single value, got 2`)
	checkParsingError(t, "?option_int32=foo&option_int32=bar", `invalid value for query parameter "option_int32": expected a single value, got 2`)
	checkParsingError(t, "?option_int64=foo&option_int64=bar", `invalid value for query parameter "option_int64": expected a single value, got 2`)
	checkParsingError(t, "?option_uint=foo&option_uint=bar", `invalid value for query parameter "option_uint": expected a single value, got 2`)
	checkParsingError(t, "?option_uint8=foo&option_uint8=bar", `invalid value for query parameter "option_uint8": expected a single value, got 2`)
	checkParsingError(t, "?option_uint16=foo&option_uint16=bar", `invalid value for query parameter "option_uint16": expected a single value, got 2`)
	checkParsingError(t, "?option_uint32=foo&option_uint32=bar", `invalid value for query parameter "option_uint32": expected a single value, got 2`)
	checkParsingError(t, "?option_uint64=foo&option_uint64=bar", `invalid value for query parameter "option_uint64": expected a single value, got 2`)
	checkParsingError(t, "?option_float32=foo&option_float32=bar", `invalid value for query parameter "option_float32": expected a single value, got 2`)
	checkParsingError(t, "?option_float64=foo&option_float64=bar", `invalid value for query parameter "option_float64": expected a single value, got 2`)

	// int overflows
	maxIntPlus1 := uint64(math.MaxInt) + 1
	checkParsingError(t, "?int="+strconv.FormatUint(maxIntPlus1, 10), fmt.Sprintf(`invalid value for query parameter "int": strconv.ParseInt: parsing "%d": value out of range`, maxIntPlus1))
	maxInt8Plus1 := uint64(math.MaxInt8) + 1
	checkParsingError(t, "?int8="+strconv.FormatUint(maxInt8Plus1, 10), fmt.Sprintf(`invalid value for query parameter "int8": strconv.ParseInt: parsing "%d": value out of range`, maxInt8Plus1))
	maxInt16Plus1 := uint64(math.MaxInt16) + 1
	checkParsingError(t, "?int16="+strconv.FormatUint(maxInt16Plus1, 10), fmt.Sprintf(`invalid value for query parameter "int16": strconv.ParseInt: parsing "%d": value out of range`, maxInt16Plus1))
	maxInt32Plus1 := uint64(math.MaxInt32) + 1
	checkParsingError(t, "?int32="+strconv.FormatUint(maxInt32Plus1, 10), fmt.Sprintf(`invalid value for query parameter "int32": strconv.ParseInt: parsing "%d": value out of range`, maxInt32Plus1))
	maxInt64Plus1 := uint64(math.MaxInt64) + 1
	checkParsingError(t, "?int64="+strconv.FormatUint(maxInt64Plus1, 10), fmt.Sprintf(`invalid value for query parameter "int64": strconv.ParseInt: parsing "%d": value out of range`, maxInt64Plus1))

	// int underflows
	minIntMinus1 := "-" + strconv.FormatUint(uint64(math.MaxInt)+2, 10)
	checkParsingError(t, "?int="+minIntMinus1, fmt.Sprintf(`invalid value for query parameter "int": strconv.ParseInt: parsing "%s": value out of range`, minIntMinus1))
	minInt8Minus1 := strconv.FormatInt(int64(math.MinInt8)-1, 10)
	checkParsingError(t, "?int8="+minInt8Minus1, fmt.Sprintf(`invalid value for query parameter "int8": strconv.ParseInt: parsing "%s": value out of range`, minInt8Minus1))
	minInt16Minus1 := strconv.FormatInt(int64(math.MinInt16)-1, 10)
	checkParsingError(t, "?int16="+minInt16Minus1, fmt.Sprintf(`invalid value for query parameter "int16": strconv.ParseInt: parsing "%s": value out of range`, minInt16Minus1))
	minInt32Minus1 := strconv.FormatInt(int64(math.MinInt32)-1, 10)
	checkParsingError(t, "?int32="+minInt32Minus1, fmt.Sprintf(`invalid value for query parameter "int32": strconv.ParseInt: parsing "%s": value out of range`, minInt32Minus1))
	minInt64Minus1 := "-" + strconv.FormatUint(uint64(math.MaxInt64)+2, 10)
	checkParsingError(t, "?int64="+minInt64Minus1, fmt.Sprintf(`invalid value for query parameter "int64": strconv.ParseInt: parsing "%s": value out of range`, minInt64Minus1))

	// uint overflows
	maxUintPlus1 := "18446744073709551616" // math.MaxUint64 + 1, cannot be represented in uint64
	checkParsingError(t, "?uint="+maxUintPlus1, fmt.Sprintf(`invalid value for query parameter "uint": strconv.ParseUint: parsing "%s": value out of range`, maxUintPlus1))
	maxUint8Plus1 := strconv.FormatUint(uint64(math.MaxUint8)+1, 10)
	checkParsingError(t, "?uint8="+maxUint8Plus1, fmt.Sprintf(`invalid value for query parameter "uint8": strconv.ParseUint: parsing "%s": value out of range`, maxUint8Plus1))
	maxUint16Plus1 := strconv.FormatUint(uint64(math.MaxUint16)+1, 10)
	checkParsingError(t, "?uint16="+maxUint16Plus1, fmt.Sprintf(`invalid value for query parameter "uint16": strconv.ParseUint: parsing "%s": value out of range`, maxUint16Plus1))
	maxUint32Plus1 := strconv.FormatUint(uint64(math.MaxUint32)+1, 10)
	checkParsingError(t, "?uint32="+maxUint32Plus1, fmt.Sprintf(`invalid value for query parameter "uint32": strconv.ParseUint: parsing "%s": value out of range`, maxUint32Plus1))
	checkParsingError(t, "?uint64="+maxUintPlus1, fmt.Sprintf(`invalid value for query parameter "uint64": strconv.ParseUint: parsing "%s": value out of range`, maxUintPlus1))

	// uint underflows (negative values rejected)
	checkParsingError(t, "?uint=-1", `invalid value for query parameter "uint": strconv.ParseUint: parsing "-1": invalid syntax`)
	checkParsingError(t, "?uint8=-1", `invalid value for query parameter "uint8": strconv.ParseUint: parsing "-1": invalid syntax`)
	checkParsingError(t, "?uint16=-1", `invalid value for query parameter "uint16": strconv.ParseUint: parsing "-1": invalid syntax`)
	checkParsingError(t, "?uint32=-1", `invalid value for query parameter "uint32": strconv.ParseUint: parsing "-1": invalid syntax`)
	checkParsingError(t, "?uint64=-1", `invalid value for query parameter "uint64": strconv.ParseUint: parsing "-1": invalid syntax`)

	// float overflows
	overflowFloat32 := strconv.FormatFloat(math.MaxFloat32*2, 'g', -1, 64)
	overflowFloat32Query := strings.ReplaceAll(overflowFloat32, "+", "%2B")
	checkParsingError(t, "?float32="+overflowFloat32Query, fmt.Sprintf(`invalid value for query parameter "float32": strconv.ParseFloat: parsing "%s": value out of range`, overflowFloat32))
	overflowFloat64 := "1.8e+309"
	overflowFloat64Query := strings.ReplaceAll(overflowFloat64, "+", "%2B")
	checkParsingError(t, "?float64="+overflowFloat64Query, fmt.Sprintf(`invalid value for query parameter "float64": strconv.ParseFloat: parsing "%s": value out of range`, overflowFloat64))

	// float underflows (negative overflow)
	underflowFloat32 := strconv.FormatFloat(-math.MaxFloat32*2, 'g', -1, 64)
	underflowFloat32Query := strings.ReplaceAll(underflowFloat32, "+", "%2B")
	checkParsingError(t, "?float32="+underflowFloat32Query, fmt.Sprintf(`invalid value for query parameter "float32": strconv.ParseFloat: parsing "%s": value out of range`, underflowFloat32))
	underflowFloat64 := "-1.8e+309"
	underflowFloat64Query := strings.ReplaceAll(underflowFloat64, "+", "%2B")
	checkParsingError(t, "?float64="+underflowFloat64Query, fmt.Sprintf(`invalid value for query parameter "float64": strconv.ParseFloat: parsing "%s": value out of range`, underflowFloat64))

	// pointer int overflows
	checkParsingError(t, "?pointer_int="+strconv.FormatUint(maxIntPlus1, 10), fmt.Sprintf(`invalid value for query parameter "pointer_int": strconv.ParseInt: parsing "%d": value out of range`, maxIntPlus1))
	checkParsingError(t, "?pointer_int8="+strconv.FormatUint(maxInt8Plus1, 10), fmt.Sprintf(`invalid value for query parameter "pointer_int8": strconv.ParseInt: parsing "%d": value out of range`, maxInt8Plus1))
	checkParsingError(t, "?pointer_int16="+strconv.FormatUint(maxInt16Plus1, 10), fmt.Sprintf(`invalid value for query parameter "pointer_int16": strconv.ParseInt: parsing "%d": value out of range`, maxInt16Plus1))
	checkParsingError(t, "?pointer_int32="+strconv.FormatUint(maxInt32Plus1, 10), fmt.Sprintf(`invalid value for query parameter "pointer_int32": strconv.ParseInt: parsing "%d": value out of range`, maxInt32Plus1))
	checkParsingError(t, "?pointer_int64="+strconv.FormatUint(maxInt64Plus1, 10), fmt.Sprintf(`invalid value for query parameter "pointer_int64": strconv.ParseInt: parsing "%d": value out of range`, maxInt64Plus1))

	// pointer int underflows
	checkParsingError(t, "?pointer_int="+minIntMinus1, fmt.Sprintf(`invalid value for query parameter "pointer_int": strconv.ParseInt: parsing "%s": value out of range`, minIntMinus1))
	checkParsingError(t, "?pointer_int8="+minInt8Minus1, fmt.Sprintf(`invalid value for query parameter "pointer_int8": strconv.ParseInt: parsing "%s": value out of range`, minInt8Minus1))
	checkParsingError(t, "?pointer_int16="+minInt16Minus1, fmt.Sprintf(`invalid value for query parameter "pointer_int16": strconv.ParseInt: parsing "%s": value out of range`, minInt16Minus1))
	checkParsingError(t, "?pointer_int32="+minInt32Minus1, fmt.Sprintf(`invalid value for query parameter "pointer_int32": strconv.ParseInt: parsing "%s": value out of range`, minInt32Minus1))
	checkParsingError(t, "?pointer_int64="+minInt64Minus1, fmt.Sprintf(`invalid value for query parameter "pointer_int64": strconv.ParseInt: parsing "%s": value out of range`, minInt64Minus1))

	// pointer float overflows
	checkParsingError(t, "?pointer_float32="+overflowFloat32Query, fmt.Sprintf(`invalid value for query parameter "pointer_float32": strconv.ParseFloat: parsing "%s": value out of range`, overflowFloat32))
	checkParsingError(t, "?pointer_float64="+overflowFloat64Query, fmt.Sprintf(`invalid value for query parameter "pointer_float64": strconv.ParseFloat: parsing "%s": value out of range`, overflowFloat64))

	// pointer float underflows (negative overflow)
	checkParsingError(t, "?pointer_float32="+underflowFloat32Query, fmt.Sprintf(`invalid value for query parameter "pointer_float32": strconv.ParseFloat: parsing "%s": value out of range`, underflowFloat32))
	checkParsingError(t, "?pointer_float64="+underflowFloat64Query, fmt.Sprintf(`invalid value for query parameter "pointer_float64": strconv.ParseFloat: parsing "%s": value out of range`, underflowFloat64))

	// slice int overflows
	checkParsingError(t, "?int_slice="+strconv.FormatUint(maxIntPlus1, 10), fmt.Sprintf(`invalid value for query parameter "int_slice": element 0: strconv.ParseInt: parsing "%d": value out of range`, maxIntPlus1))
	checkParsingError(t, "?int8_slice="+strconv.FormatUint(maxInt8Plus1, 10), fmt.Sprintf(`invalid value for query parameter "int8_slice": element 0: strconv.ParseInt: parsing "%d": value out of range`, maxInt8Plus1))
	checkParsingError(t, "?int16_slice="+strconv.FormatUint(maxInt16Plus1, 10), fmt.Sprintf(`invalid value for query parameter "int16_slice": element 0: strconv.ParseInt: parsing "%d": value out of range`, maxInt16Plus1))
	checkParsingError(t, "?int32_slice="+strconv.FormatUint(maxInt32Plus1, 10), fmt.Sprintf(`invalid value for query parameter "int32_slice": element 0: strconv.ParseInt: parsing "%d": value out of range`, maxInt32Plus1))
	checkParsingError(t, "?int64_slice="+strconv.FormatUint(maxInt64Plus1, 10), fmt.Sprintf(`invalid value for query parameter "int64_slice": element 0: strconv.ParseInt: parsing "%d": value out of range`, maxInt64Plus1))

	// slice int underflows
	checkParsingError(t, "?int_slice="+minIntMinus1, fmt.Sprintf(`invalid value for query parameter "int_slice": element 0: strconv.ParseInt: parsing "%s": value out of range`, minIntMinus1))
	checkParsingError(t, "?int8_slice="+minInt8Minus1, fmt.Sprintf(`invalid value for query parameter "int8_slice": element 0: strconv.ParseInt: parsing "%s": value out of range`, minInt8Minus1))
	checkParsingError(t, "?int16_slice="+minInt16Minus1, fmt.Sprintf(`invalid value for query parameter "int16_slice": element 0: strconv.ParseInt: parsing "%s": value out of range`, minInt16Minus1))
	checkParsingError(t, "?int32_slice="+minInt32Minus1, fmt.Sprintf(`invalid value for query parameter "int32_slice": element 0: strconv.ParseInt: parsing "%s": value out of range`, minInt32Minus1))
	checkParsingError(t, "?int64_slice="+minInt64Minus1, fmt.Sprintf(`invalid value for query parameter "int64_slice": element 0: strconv.ParseInt: parsing "%s": value out of range`, minInt64Minus1))

	// slice uint overflows
	checkParsingError(t, "?uint_slice="+maxUintPlus1, fmt.Sprintf(`invalid value for query parameter "uint_slice": element 0: strconv.ParseUint: parsing "%s": value out of range`, maxUintPlus1))
	checkParsingError(t, "?uint8_slice="+maxUint8Plus1, fmt.Sprintf(`invalid value for query parameter "uint8_slice": element 0: strconv.ParseUint: parsing "%s": value out of range`, maxUint8Plus1))
	checkParsingError(t, "?uint16_slice="+maxUint16Plus1, fmt.Sprintf(`invalid value for query parameter "uint16_slice": element 0: strconv.ParseUint: parsing "%s": value out of range`, maxUint16Plus1))
	checkParsingError(t, "?uint32_slice="+maxUint32Plus1, fmt.Sprintf(`invalid value for query parameter "uint32_slice": element 0: strconv.ParseUint: parsing "%s": value out of range`, maxUint32Plus1))
	checkParsingError(t, "?uint64_slice="+maxUintPlus1, fmt.Sprintf(`invalid value for query parameter "uint64_slice": element 0: strconv.ParseUint: parsing "%s": value out of range`, maxUintPlus1))

	// slice uint underflows (negative values rejected)
	checkParsingError(t, "?uint_slice=-1", `invalid value for query parameter "uint_slice": element 0: strconv.ParseUint: parsing "-1": invalid syntax`)
	checkParsingError(t, "?uint8_slice=-1", `invalid value for query parameter "uint8_slice": element 0: strconv.ParseUint: parsing "-1": invalid syntax`)
	checkParsingError(t, "?uint16_slice=-1", `invalid value for query parameter "uint16_slice": element 0: strconv.ParseUint: parsing "-1": invalid syntax`)
	checkParsingError(t, "?uint32_slice=-1", `invalid value for query parameter "uint32_slice": element 0: strconv.ParseUint: parsing "-1": invalid syntax`)
	checkParsingError(t, "?uint64_slice=-1", `invalid value for query parameter "uint64_slice": element 0: strconv.ParseUint: parsing "-1": invalid syntax`)

	// slice float overflows
	checkParsingError(t, "?float32_slice="+overflowFloat32Query, fmt.Sprintf(`invalid value for query parameter "float32_slice": element 0: strconv.ParseFloat: parsing "%s": value out of range`, overflowFloat32))
	checkParsingError(t, "?float64_slice="+overflowFloat64Query, fmt.Sprintf(`invalid value for query parameter "float64_slice": element 0: strconv.ParseFloat: parsing "%s": value out of range`, overflowFloat64))

	// slice float underflows (negative overflow)
	checkParsingError(t, "?float32_slice="+underflowFloat32Query, fmt.Sprintf(`invalid value for query parameter "float32_slice": element 0: strconv.ParseFloat: parsing "%s": value out of range`, underflowFloat32))
	checkParsingError(t, "?float64_slice="+underflowFloat64Query, fmt.Sprintf(`invalid value for query parameter "float64_slice": element 0: strconv.ParseFloat: parsing "%s": value out of range`, underflowFloat64))

	// option int overflows
	checkParsingError(t, "?option_int="+strconv.FormatUint(maxIntPlus1, 10), fmt.Sprintf(`invalid value for query parameter "option_int": strconv.ParseInt: parsing "%d": value out of range`, maxIntPlus1))
	checkParsingError(t, "?option_int8="+strconv.FormatUint(maxInt8Plus1, 10), fmt.Sprintf(`invalid value for query parameter "option_int8": strconv.ParseInt: parsing "%d": value out of range`, maxInt8Plus1))
	checkParsingError(t, "?option_int16="+strconv.FormatUint(maxInt16Plus1, 10), fmt.Sprintf(`invalid value for query parameter "option_int16": strconv.ParseInt: parsing "%d": value out of range`, maxInt16Plus1))
	checkParsingError(t, "?option_int32="+strconv.FormatUint(maxInt32Plus1, 10), fmt.Sprintf(`invalid value for query parameter "option_int32": strconv.ParseInt: parsing "%d": value out of range`, maxInt32Plus1))
	checkParsingError(t, "?option_int64="+strconv.FormatUint(maxInt64Plus1, 10), fmt.Sprintf(`invalid value for query parameter "option_int64": strconv.ParseInt: parsing "%d": value out of range`, maxInt64Plus1))

	// option int underflows
	checkParsingError(t, "?option_int="+minIntMinus1, fmt.Sprintf(`invalid value for query parameter "option_int": strconv.ParseInt: parsing "%s": value out of range`, minIntMinus1))
	checkParsingError(t, "?option_int8="+minInt8Minus1, fmt.Sprintf(`invalid value for query parameter "option_int8": strconv.ParseInt: parsing "%s": value out of range`, minInt8Minus1))
	checkParsingError(t, "?option_int16="+minInt16Minus1, fmt.Sprintf(`invalid value for query parameter "option_int16": strconv.ParseInt: parsing "%s": value out of range`, minInt16Minus1))
	checkParsingError(t, "?option_int32="+minInt32Minus1, fmt.Sprintf(`invalid value for query parameter "option_int32": strconv.ParseInt: parsing "%s": value out of range`, minInt32Minus1))
	checkParsingError(t, "?option_int64="+minInt64Minus1, fmt.Sprintf(`invalid value for query parameter "option_int64": strconv.ParseInt: parsing "%s": value out of range`, minInt64Minus1))

	// option float overflows
	checkParsingError(t, "?option_float32="+overflowFloat32Query, fmt.Sprintf(`invalid value for query parameter "option_float32": strconv.ParseFloat: parsing "%s": value out of range`, overflowFloat32))
	checkParsingError(t, "?option_float64="+overflowFloat64Query, fmt.Sprintf(`invalid value for query parameter "option_float64": strconv.ParseFloat: parsing "%s": value out of range`, overflowFloat64))

	// option float underflows (negative overflow)
	checkParsingError(t, "?option_float32="+underflowFloat32Query, fmt.Sprintf(`invalid value for query parameter "option_float32": strconv.ParseFloat: parsing "%s": value out of range`, underflowFloat32))
	checkParsingError(t, "?option_float64="+underflowFloat64Query, fmt.Sprintf(`invalid value for query parameter "option_float64": strconv.ParseFloat: parsing "%s": value out of range`, underflowFloat64))

	// option uint overflows
	checkParsingError(t, "?option_uint="+maxUintPlus1, fmt.Sprintf(`invalid value for query parameter "option_uint": strconv.ParseUint: parsing "%s": value out of range`, maxUintPlus1))
	checkParsingError(t, "?option_uint8="+maxUint8Plus1, fmt.Sprintf(`invalid value for query parameter "option_uint8": strconv.ParseUint: parsing "%s": value out of range`, maxUint8Plus1))
	checkParsingError(t, "?option_uint16="+maxUint16Plus1, fmt.Sprintf(`invalid value for query parameter "option_uint16": strconv.ParseUint: parsing "%s": value out of range`, maxUint16Plus1))
	checkParsingError(t, "?option_uint32="+maxUint32Plus1, fmt.Sprintf(`invalid value for query parameter "option_uint32": strconv.ParseUint: parsing "%s": value out of range`, maxUint32Plus1))
	checkParsingError(t, "?option_uint64="+maxUintPlus1, fmt.Sprintf(`invalid value for query parameter "option_uint64": strconv.ParseUint: parsing "%s": value out of range`, maxUintPlus1))

	// option uint underflows (negative values rejected)
	checkParsingError(t, "?option_uint=-1", `invalid value for query parameter "option_uint": strconv.ParseUint: parsing "-1": invalid syntax`)
	checkParsingError(t, "?option_uint8=-1", `invalid value for query parameter "option_uint8": strconv.ParseUint: parsing "-1": invalid syntax`)
	checkParsingError(t, "?option_uint16=-1", `invalid value for query parameter "option_uint16": strconv.ParseUint: parsing "-1": invalid syntax`)
	checkParsingError(t, "?option_uint32=-1", `invalid value for query parameter "option_uint32": strconv.ParseUint: parsing "-1": invalid syntax`)
	checkParsingError(t, "?option_uint64=-1", `invalid value for query parameter "option_uint64": strconv.ParseUint: parsing "-1": invalid syntax`)

	// pointer uint overflows
	checkParsingError(t, "?pointer_uint="+maxUintPlus1, fmt.Sprintf(`invalid value for query parameter "pointer_uint": strconv.ParseUint: parsing "%s": value out of range`, maxUintPlus1))
	checkParsingError(t, "?pointer_uint8="+maxUint8Plus1, fmt.Sprintf(`invalid value for query parameter "pointer_uint8": strconv.ParseUint: parsing "%s": value out of range`, maxUint8Plus1))
	checkParsingError(t, "?pointer_uint16="+maxUint16Plus1, fmt.Sprintf(`invalid value for query parameter "pointer_uint16": strconv.ParseUint: parsing "%s": value out of range`, maxUint16Plus1))
	checkParsingError(t, "?pointer_uint32="+maxUint32Plus1, fmt.Sprintf(`invalid value for query parameter "pointer_uint32": strconv.ParseUint: parsing "%s": value out of range`, maxUint32Plus1))
	checkParsingError(t, "?pointer_uint64="+maxUintPlus1, fmt.Sprintf(`invalid value for query parameter "pointer_uint64": strconv.ParseUint: parsing "%s": value out of range`, maxUintPlus1))

	// pointer uint underflows (negative values rejected)
	checkParsingError(t, "?pointer_uint=-1", `invalid value for query parameter "pointer_uint": strconv.ParseUint: parsing "-1": invalid syntax`)
	checkParsingError(t, "?pointer_uint8=-1", `invalid value for query parameter "pointer_uint8": strconv.ParseUint: parsing "-1": invalid syntax`)
	checkParsingError(t, "?pointer_uint16=-1", `invalid value for query parameter "pointer_uint16": strconv.ParseUint: parsing "-1": invalid syntax`)
	checkParsingError(t, "?pointer_uint32=-1", `invalid value for query parameter "pointer_uint32": strconv.ParseUint: parsing "-1": invalid syntax`)
	checkParsingError(t, "?pointer_uint64=-1", `invalid value for query parameter "pointer_uint64": strconv.ParseUint: parsing "-1": invalid syntax`)
}
