// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package opts

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	. "go.xyrillian.de/gg/option"
)

// ParseQueryString parses the query parameters of a *http.Request into an opt struct.
// Fields are mapped by their "q" tag, mirroring the behavior of [opts.BuildQueryString].
// For example:
//
//	type struct Something {
//	   Bar string `q:"x_bar"`
//	   Baz int    `q:"lorem_ipsum"`
//	}
//
// and a request with the query string "?x_bar=AAA&lorem_ipsum=BBB" will produce
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
// The parser supports all scalars except complex. Structs of scalars are supported
// and special structs like time.Time. Additionally, it allows Slices (for multiple
// values), option.Option (recommended for optional values) and pointers (as alternative
// to option.Option) of these. Only map[string]string is supported as a map type.
// Some inputs might work but are untested.
//
// Slice fields use repeated query parameters:
// Foo []string `q:"foo"`                       // ?foo=a&foo=b
//
// Map fields accept plain repeated key:value pairs:
// Bar map[string]string `q:"bar"`              // ?bar=k1:v1&bar=k2:v2
//
// time.Time fields support the formats RFC3339Nano, RFC3339, DateTime, Date, Unix.
// A single `format` option can be set, to limit what the parser accepts:
// Baz time.Time `q:"baz,format:RFC3339"`       // ?baz=1999-01-01T00:00:00
//
// A `required` option can be set to define that a missing value will produce an error.
// Quux string `q:"quux,required"`               // ?foo=bar --> error
func ParseQueryString[T any](r *http.Request) (T, error) {
	var opts T
	optsValue := reflect.ValueOf(&opts).Elem()
	if optsValue.Kind() != reflect.Struct {
		panic("expected opts to point to a struct")
	}

	// build map of q-tags to field values
	optsType := optsValue.Type()
	type optMeta struct {
		index             int
		format            string // "" or time-formats
		requiredAndUnseen bool
	}
	knownOpts := make(map[string]optMeta, optsType.NumField())
	for i := range optsType.NumField() {
		field := optsType.Field(i)
		fieldValue := optsValue.Field(i)
		qTag := field.Tag.Get("q")
		if qTag == "" {
			panic(fmt.Sprintf(`expected %q to have a "q:"-tag`, field.Name))
		}
		optKey, format, required := parseQTag(qTag)
		if !fieldValue.CanSet() {
			panic(fmt.Sprintf(`field %q is unexported and therefore cannot be set`, optKey))
		}
		// When format is set, it must be a known time format.
		if _, ok := timeFormats[format]; format != "" && format != unixFormat && !ok {
			return opts, fmt.Errorf("unsupported time format %q; accepted: %s, %s", format, supportedHumanReadableFormats, unixFormat)
		}
		knownOpts[optKey] = optMeta{
			index:             i,
			format:            format,
			requiredAndUnseen: required}
	}

	// iterate the query
	query := r.URL.Query()
	for optKey, rawValues := range query {
		meta, ok := knownOpts[optKey]
		if !ok {
			return opts, fmt.Errorf("unknown query parameter %q", optKey)
		}
		fieldValue := optsValue.Field(meta.index)
		// now, just set the value of the field accordingly
		err := setField(fieldValue, rawValues, meta.format)
		if err != nil {
			return opts, fmt.Errorf("invalid value for query parameter %q: %w", optKey, err)
		}
		if !isZero(fieldValue) && meta.requiredAndUnseen {
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
// The format parameter carries the format option from the q tag (may be empty).
func setField(fv reflect.Value, values []string, format string) error {
	// unwrap pointer: allocate if nil
	if fv.Kind() == reflect.Pointer {
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		fv = fv.Elem()
	}

	// Handle Option[T] fields: detected by the presence of an IsSome() bool method.
	// We parse into a temporary value of the inner type by recursing into setField,
	// then use UnmarshalYAML to set the Option value directly.
	if _, isOption := fv.Type().MethodByName("IsSome"); isOption {
		if len(values) == 0 {
			return nil // leave as None (zero value)
		}
		tmp := reflect.New(fv.Type().Field(0).Type).Elem()
		if err := setField(tmp, values, format); err != nil {
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

	// Handle time.Time fields: if a format is specified, use only that format;
	// otherwise try multiple formats in order of specificity.
	if fv.Type() == reflect.TypeFor[time.Time]() {
		var formatOpt Option[string]
		if format != "" {
			formatOpt = Some(format)
		}
		t, err := parseTime(values[0], formatOpt)
		if err != nil {
			return err
		}
		fv.Set(reflect.ValueOf(t))
		return nil
	}

	// set scalars
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(values[0])
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(values[0], 10, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(values[0], 10, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetUint(n)
	case reflect.Bool:
		b, err := strconv.ParseBool(values[0])
		if err != nil {
			return err
		}
		fv.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(values[0], fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetFloat(f)
	// set slices
	case reflect.Slice:
		elemType := fv.Type().Elem()
		switch elemType.Kind() {
		case reflect.String:
			sl := reflect.MakeSlice(fv.Type(), len(values), len(values))
			for i, v := range values {
				sl.Index(i).SetString(v)
			}
			fv.Set(sl)
		case reflect.Bool:
			sl := reflect.MakeSlice(fv.Type(), len(values), len(values))
			for i, v := range values {
				b, err := strconv.ParseBool(v)
				if err != nil {
					return fmt.Errorf("element %d: %w", i, err)
				}
				sl.Index(i).SetBool(b)
			}
			fv.Set(sl)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			sl := reflect.MakeSlice(fv.Type(), len(values), len(values))
			for i, v := range values {
				n, err := strconv.ParseInt(v, 10, elemType.Bits())
				if err != nil {
					return fmt.Errorf("element %d: %w", i, err)
				}
				sl.Index(i).SetInt(n)
			}
			fv.Set(sl)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			sl := reflect.MakeSlice(fv.Type(), len(values), len(values))
			for i, v := range values {
				n, err := strconv.ParseUint(v, 10, elemType.Bits())
				if err != nil {
					return fmt.Errorf("element %d: %w", i, err)
				}
				sl.Index(i).SetUint(n)
			}
			fv.Set(sl)
		case reflect.Float32, reflect.Float64:
			sl := reflect.MakeSlice(fv.Type(), len(values), len(values))
			for i, v := range values {
				f, err := strconv.ParseFloat(v, elemType.Bits())
				if err != nil {
					return fmt.Errorf("element %d: %w", i, err)
				}
				sl.Index(i).SetFloat(f)
			}
			fv.Set(sl)
		case reflect.Struct:
			if elemType != reflect.TypeFor[time.Time]() {
				return fmt.Errorf("unsupported slice element type %s", elemType)
			}
			sl := reflect.MakeSlice(fv.Type(), len(values), len(values))
			for i, v := range values {
				var formatOpt Option[string]
				if format != "" {
					formatOpt = Some(format)
				}
				t, err := parseTime(v, formatOpt)
				if err != nil {
					return fmt.Errorf("element %d: %w", i, err)
				}
				sl.Index(i).Set(reflect.ValueOf(t))
			}
			fv.Set(sl)
		default:
			return fmt.Errorf("unsupported slice element type %s", elemType)
		}
	// set maps
	case reflect.Map:
		if fv.Type().Key().Kind() != reflect.String || fv.Type().Elem().Kind() != reflect.String {
			return errors.New("only map[string]string is supported")
		}
		m, err := parseMapValues(values)
		if err != nil {
			return err
		}
		fv.Set(reflect.ValueOf(m))
	default:
		return fmt.Errorf("unsupported field type %s", fv.Type())
	}
	return nil
}

// parseMapValues parses a list of raw string values into a map[string]string.
// It supports a repeated key:value notation (?m=k1:v1&m=k2:v2).
func parseMapValues(values []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, raw := range values {
		raw = strings.TrimSpace(raw)
		kv := strings.SplitN(raw, ":", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid map entry %q: expected key:value", raw)
		}
		m[kv[0]] = kv[1]
	}
	return m, nil
}

// parseTime parses a time string. If format is Some, only that specific format
// is accepted (one of "RFC3339Nano", "RFC3339", "DateTime", "DateOnly", "Unix").
// If format is None, all formats are tried in order of specificity.
func parseTime(s string, format Option[string]) (time.Time, error) {
	if f, ok := format.Unpack(); ok {
		if f == "Unix" {
			n, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return time.Time{}, fmt.Errorf("cannot parse %q as %s seconds: %w", s, unixFormat, err)
			}
			return time.Unix(n, 0).UTC(), nil
		}
		// we checked this already when building knownOpts
		layout := timeFormats[f]
		t, err := time.Parse(layout, s)
		if err != nil {
			return time.Time{}, fmt.Errorf("cannot parse %q as %s: %w", s, f, err)
		}
		return t, nil
	}

	// No specific format — try all in order
	for _, layout := range timeFormats {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
	}
	// Try Unix timestamp (non-negative seconds since epoch)
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(n, 0).UTC(), nil
	}
	return time.Time{}, fmt.Errorf("cannot parse %q as time; accepted: %s, %s", s, supportedHumanReadableFormats, unixFormat)
}
