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
//	?x_bar=AAA&lorem_ipsum=BBB
//
// will produce
//
//	result := Something{
//	   Bar: "AAA",
//	   Baz: "BBB",
//	}
//
// On configuration errors (e.g. non-struct opts, opts with non-q-tagged fields)
// the function panics. On user errors (unknown query parameter, type conversion
// failure, missing required field) an error is returned. On success, the returned
// opts are populated according to the http.Request.
//
// The parser supports all scalars except complex. Additionally, it allows Slices
// (for multiple values), [option.Option] (recommended for optional values) and
// pointers (as alternative to [option.Option]) of these. Only map[string]string
// is supported as a map type. Embedded structs and [time.Time] are supported.
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
// [time.Time] fields support the formats defined under [opts.TimeFormats].
// A single "format" option must be set, to limit what the parser accepts:
//
//	Baz time.Time `q:"baz,format:RFC3339"`       // ?baz=1999-01-01T00:00:00
//
// A "required" option can be set to define that a missing value will produce an error.
//
//	Quux string `q:"quux,required"`               // ?foo=bar --> error
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
		timeFormat        string // "" or time-formats
		requiredAndUnseen bool
	}
	knownOpts := make(map[string]optMeta, optsType.NumField())
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
		optKey, timeFormat, required := parseQTag(qTag)
		if !fieldValue.CanSet() {
			panic(fmt.Sprintf(`field %q is unexported and therefore cannot be set`, optKey))
		}
		// all known formats are currently timeFormats
		if _, ok := TimeFormats[timeFormat]; timeFormat != "" && timeFormat != UnixFormat && !ok {
			panic(fmt.Sprintf("unsupported time format %q; accepted: %s", timeFormat, supportedHumanReadableFormats))
		}
		knownOpts[optKey] = optMeta{
			structField:       field,
			fieldValue:        fieldValue,
			timeFormat:        timeFormat,
			requiredAndUnseen: required,
		}
	}

	// iterate the query
	for optKey, rawValues := range query {
		meta, ok := knownOpts[optKey]
		if !ok {
			return opts, fmt.Errorf("unknown query parameter %q", optKey)
		}
		// now, just set the value of the field accordingly
		err := setField(meta.structField, meta.fieldValue, rawValues, meta.timeFormat)
		if err != nil {
			return opts, fmt.Errorf("invalid value for query parameter %q: %w", optKey, err)
		}
		if !isZero(meta.fieldValue) && meta.requiredAndUnseen {
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

// setField writes values into a single struct field.
// The timeFormat parameter carries the format option from the q tag (may be empty).
func setField(f reflect.StructField, fv reflect.Value, values []string, timeFormat string) error {
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
		tmp := reflect.New(fv.Type().Field(0).Type).Elem()
		if err := setField(f, tmp, values, timeFormat); err != nil {
			return err
		}
		unmarshal := func(dest any) error {
			// dest is **T (pointer to the *T that UnmarshalYAML allocated).
			// We need to set *dest to point to our parsed tmp value.
			ptr := reflect.New(tmp.Type())
			ptr.Elem().Set(tmp)
			reflect.ValueOf(dest).Elem().Set(ptr)
			return nil
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
		if timeFormat == "" {
			panic("time format is missing for field " + f.Name)
		}
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
		m, err := parseMapValues(values, fv.Type().Key(), fv.Type().Elem())
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

// parseMapValues parses a list of raw string values into a map with the given key and value types.
// Each value must be in "key:value" notation (e.g. ?m=k1:v1&m=k2:v2).
func parseMapValues(values []string, keyType, valType reflect.Type) (reflect.Value, error) {
	m := reflect.MakeMap(reflect.MapOf(keyType, valType))
	for _, raw := range values {
		raw = strings.TrimSpace(raw)
		kv := strings.SplitN(raw, ":", 2)
		if len(kv) != 2 {
			return reflect.Value{}, fmt.Errorf("invalid map entry %q: expected key:value", raw)
		}
		key, err := parseScalar(kv[0], keyType)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid map key %q: %w", kv[0], err)
		}
		val, err := parseScalar(kv[1], valType)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid map value %q: %w", kv[1], err)
		}
		m.SetMapIndex(key, val)
	}
	return m, nil
}

// parseTime parses a time string. Accepted formats are defined in opts.TimeFormats.
func parseTime(s, timeFormat string) (time.Time, error) {
	if timeFormat == UnixFormat {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("cannot parse %q as %s seconds: %w", s, UnixFormat, err)
		}
		return time.Unix(n, 0).UTC(), nil
	}
	// we checked this already when building knownOpts
	layout := TimeFormats[timeFormat]
	t, err := time.Parse(layout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse %q as %s: %w", s, timeFormat, err)
	}
	return t, nil
}
