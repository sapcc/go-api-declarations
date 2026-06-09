// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package opts

import (
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	. "go.xyrillian.de/gg/option"
)

// ParseQueryString parses the query parameters of url.Values into an opt struct.
// Fields are mapped by their "q" tag, mirroring the behavior of [opts.BuildQueryString].
// For example:
//
//	type struct Something {
//	   Bar string `q:"x_bar"`
//	   Baz int    `q:"lorem_ipsum"`
//	}
//
// and a request with the query string
//
//	?x_bar=AAA&lorem_ipsum=1
//
// will produce
//
//	result := Something{
//	   Bar: "AAA",
//	   Baz: 1,
//	}
//
// On configuration errors (e.g. non-struct opts, opts with non-q-tagged fields)
// the function panics. On user errors (unknown query parameter, type conversion
// failure, missing required field) an error is returned. On success, the returned
// opts are populated according to the http.Request.
//
// The parser supports all scalars except complex. Additionally, it allows Slices
// (for multiple values), [option.Option] (recommended for optional values) and
// pointers (as alternative to [option.Option]) of these. Maps are supported when
// the key and value types are supported. Only selected structs work
// (embedded structs and [time.Time]).
// Some inputs might work but are untested.
//
// Slice fields use repeated query parameters:
//
//	Foo []string `q:"foo"`                       // ?foo=a&foo=b
//
// Map fields accept plain repeated key:value pairs:
//
//	Bar map[string]string `q:"bar"`              // ?bar=k1:v1&bar=k2:v2
//
// [time.Time] fields support the formats RFC3339Nano, RFC3339, DateTime, DateOnly, Unix
// (seconds since epoch). A single "format" option must be set, to limit what the parser accepts:
//
//	Baz time.Time `q:"baz,format:RFC3339"`       // ?baz=1999-01-01T00:00:00
//
// A "required" option can be set to define that a missing value will produce an error.
//
//	Quux string `q:"quux,required"`               // ?foo=bar --> error
//
// Bool fields can use a "value" option to participate in value-discriminant parsing.
// Multiple bool fields sharing the same key each declare a specific value they match.
// When the query contains that value for the key, the corresponding bool is set to true:
//
//	WithDetails       bool `q:"with,value:details"`
//	WithSubresources  bool `q:"with,value:subresources"`
//	WithSubcapacities bool `q:"with,value:subcapacities"`
//
// Given the query string ?with=details&with=subcapacities, WithDetails and
// WithSubcapacities will be true while WithSubresources remains false.
//
// [option.Option]: https://pkg.go.dev/go.xyrillian.de/gg/option#Option
func ParseQueryString[T any](query url.Values) (T, error) {
	var opts T
	optsValue := reflect.ValueOf(&opts).Elem()
	if optsValue.Kind() != reflect.Struct {
		panic("expected opts to point to a struct")
	}

	// build map of q-tags to field values
	optsType := optsValue.Type()
	type optMeta struct {
		structField       reflect.StructField
		fieldValue        reflect.Value
		maybeTimeFormat   Option[string]
		requiredAndUnseen bool
	}
	// valueField tracks a bool field with a value-discriminant tag.
	type valueField struct {
		fieldValue reflect.Value
		value      string
	}
	knownOpts := make(map[string]optMeta, optsType.NumField())
	valueFields := make(map[string][]valueField) // key → list of value-discriminant bools
	for _, field := range reflect.VisibleFields(optsType) {
		fieldValue := optsValue.FieldByIndex(field.Index)
		qTag := field.Tag.Get("q")
		if field.Anonymous {
			// this is for struct embedding to work
			if qTag != "" {
				panic(fmt.Sprintf(`expected embedded struct %q to have no "q:"-tag`, field.Name))
			}
			continue
		}
		if qTag == "" {
			panic(fmt.Sprintf(`expected %q to have a "q:"-tag`, field.Name))
		}
		optKey, maybeTimeFormat, maybeValue, required := parseQTag(qTag)
		if !fieldValue.CanSet() {
			panic(fmt.Sprintf(`field %q is unexported and therefore cannot be set`, optKey))
		}

		// value-discriminant fields go into a separate structure
		if value, ok := maybeValue.Unpack(); ok {
			if fieldValue.Kind() != reflect.Bool {
				panic(fmt.Sprintf(`field %q has "value:" option but is not a bool`, field.Name))
			}
			if maybeTimeFormat.IsSome() {
				panic(fmt.Sprintf(`field %q cannot have both "value:" and "format:" options`, field.Name))
			}
			valueFields[optKey] = append(valueFields[optKey], valueField{
				fieldValue: fieldValue,
				value:      value,
			})
			continue
		}

		// check if format is set when required
		if typeNeedsTimeFormat(fieldValue.Type()) && maybeTimeFormat.IsNone() {
			panic(fmt.Sprintf(`time format is missing for field %q`, field.Name))
		}

		knownOpts[optKey] = optMeta{
			structField:       field,
			fieldValue:        fieldValue,
			maybeTimeFormat:   maybeTimeFormat,
			requiredAndUnseen: required,
		}
	}

	// iterate the query
	for optKey, rawValues := range query {
		// check value-discriminant fields first
		if vfs, ok := valueFields[optKey]; ok {
			for _, rawValue := range rawValues {
				matched := false
				for _, vf := range vfs {
					if rawValue == vf.value {
						vf.fieldValue.SetBool(true)
						matched = true
						break
					}
				}
				if !matched {
					return opts, fmt.Errorf("unknown value %q for query parameter %q", rawValue, optKey)
				}
			}
			continue
		}

		meta, ok := knownOpts[optKey]
		if !ok {
			return opts, fmt.Errorf("unknown query parameter %q", optKey)
		}
		err := setField(meta.structField, meta.fieldValue, rawValues, meta.maybeTimeFormat)
		if err != nil {
			return opts, fmt.Errorf("invalid value for query parameter %q: %w", optKey, err)
		}
		if !isOnlyEmptyStrings(rawValues) && meta.requiredAndUnseen {
			meta.requiredAndUnseen = false
			knownOpts[optKey] = meta
		}
	}

	// check that no required fields are missing
	for optKey, meta := range knownOpts {
		if meta.requiredAndUnseen {
			return opts, fmt.Errorf("missing value for query parameter %q", optKey)
		}
	}
	return opts, nil
}

// isOnlyEmptyStrings checks if all rawValues are only emptyStrings.
func isOnlyEmptyStrings(rawValues []string) bool {
	for _, rawValue := range rawValues {
		if rawValue != "" {
			return false
		}
	}
	return true
}

// setField writes values into a single struct field.
// The timeFormat parameter carries the format option from the q tag (may be empty).
func setField(f reflect.StructField, fv reflect.Value, values []string, timeFormat Option[string]) error {
	if len(values) == 0 {
		return nil
	}

	// unwrap pointer: allocate if nil
	if fv.Kind() == reflect.Pointer {
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		fv = fv.Elem()
	}

	// handle Option[T] fields: detected by the presence of an IsSome() method
	// We parse into a temporary value of the inner type by recursing into setField,
	// then use UnmarshalYAML to set the Option value directly.
	if _, isOption := fv.Type().MethodByName("IsSome"); isOption {
		unmarshal := func(dest any) error {
			// dest is **T; allocate *T and parse into T directly
			destVal := reflect.ValueOf(dest).Elem()         // *T
			destVal.Set(reflect.New(destVal.Type().Elem())) // allocate T, set *T
			inner := destVal.Elem()                         // T (the actual value to fill)
			return setField(f, inner, values, timeFormat)
		}
		type yamlUnmarshaler interface {
			UnmarshalYAML(func(any) error) error
		}
		return fv.Addr().Interface().(yamlUnmarshaler).UnmarshalYAML(unmarshal)
	}

	// some common error checks
	if slices.Contains([]reflect.Kind{reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Bool, reflect.Float32, reflect.Float64}, fv.Kind()) || fv.Type() == reflect.TypeFor[time.Time]() {
		if len(values) == 0 {
			return nil
		}
		if len(values) > 1 {
			return fmt.Errorf("expected a single value, got %d", len(values))
		}
	}

	// handle time.Time fields
	if fv.Type() == reflect.TypeFor[time.Time]() {
		t, err := parseTime(values[0], timeFormat)
		if err != nil {
			return err
		}
		fv.Set(reflect.ValueOf(t))
		return nil
	}

	// set scalars
	switch fv.Kind() {
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Bool, reflect.Float32, reflect.Float64:
		v, err := parseScalar(values[0], fv.Type())
		if err != nil {
			return err
		}
		fv.Set(v)
	// set slices
	case reflect.Slice:
		elemType := fv.Type().Elem()
		if elemType == reflect.TypeFor[time.Time]() {
			sl := reflect.MakeSlice(fv.Type(), len(values), len(values))
			for i, v := range values {
				t, err := parseTime(v, timeFormat)
				if err != nil {
					return fmt.Errorf("element %d: %w", i, err)
				}
				sl.Index(i).Set(reflect.ValueOf(t))
			}
			fv.Set(sl)
		} else {
			sl := reflect.MakeSlice(fv.Type(), len(values), len(values))
			for i, v := range values {
				elem, err := parseScalar(v, elemType)
				if err != nil {
					return fmt.Errorf("element %d: %w", i, err)
				}
				sl.Index(i).Set(elem)
			}
			fv.Set(sl)
		}
	// set maps
	case reflect.Map:
		m, err := parseMapValues(values, fv.Type())
		if err != nil {
			return err
		}
		fv.Set(m)
	default:
		return fmt.Errorf("unsupported field type %s", fv.Type())
	}
	return nil
}

// parseScalar parses a single string into a reflect.Value of the given type.
// Supported kinds: string, int*, uint*, float*, bool.
func parseScalar(s string, t reflect.Type) (reflect.Value, error) {
	v := reflect.New(t).Elem()
	switch t.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, t.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 10, t.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		v.SetUint(n)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, t.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		v.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return reflect.Value{}, err
		}
		v.SetBool(b)
	default:
		return reflect.Value{}, fmt.Errorf("unsupported type %s", t)
	}
	return v, nil
}

// parseMapValues parses a list of raw string values into a map with the given type.
// Each value must be in "key:value" notation (e.g. ?m=k1:v1&m=k2:v2).
func parseMapValues(values []string, mapType reflect.Type) (reflect.Value, error) {
	m := reflect.MakeMapWithSize(mapType, len(values))
	for _, raw := range values {
		raw = strings.TrimSpace(raw)
		kv := strings.SplitN(raw, ":", 2)
		if len(kv) != 2 {
			return reflect.Value{}, fmt.Errorf("invalid map entry %q: expected key:value", raw)
		}
		key, err := parseScalar(kv[0], mapType.Key())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid map key %q: %w", kv[0], err)
		}
		val, err := parseScalar(kv[1], mapType.Elem())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid map value %q: %w", kv[1], err)
		}
		m.SetMapIndex(key, val)
	}
	return m, nil
}

// parseTime parses a time string. Accepted non-unix formats are defined in opts.nonUnixTimeFormats.
func parseTime(s string, timeFormat Option[string]) (time.Time, error) {
	tf := timeFormat.UnwrapOrPanic("timeFormat should have been set")
	if tf == unixTimeFormat {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("cannot parse %q as %s seconds: %w", s, unixTimeFormat, err)
		}
		return time.Unix(n, 0).UTC(), nil
	}
	// we checked this already when building knownOpts
	layout := nonUnixTimeFormats[tf]
	t, err := time.Parse(layout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse %q as %s: %w", s, tf, err)
	}
	return t, nil
}
