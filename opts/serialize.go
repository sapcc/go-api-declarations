// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package opts

import (
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"time"
)

// BuildQueryString is a function to be used by request methods with opts structs.
// It's inspired by gophercloud.BuildQueryString with partially stricter behavior.
// It accepts a tagged structure and expands it into a URL struct. Field names are
// converted into query parameters based on a "q" tag. For example:
//
//	type struct Something {
//	   Bar string `q:"x_bar"`
//	   Baz int    `q:"lorem_ipsum"`
//	}
//	instance := Something{
//	   Bar: "AAA",
//	   Baz: "BBB",
//	}
//
// will be converted into "?x_bar=AAA&lorem_ipsum=BBB".
//
// On configuration errors (e.g. non-struct opts, opts with non-q-tagged fields)
// the function panics. On user errors (e.g. missing required field) an error
// is returned. On success, url.Values are returned according to the opts.
//
// The serialization supports all scalars except complex. Structs of scalars are supported
// and special structs like time.Time. Additionally, it allows Slices (for multiple
// values), option.Option (recommended for optional values) and pointers (as alternative
// to option.Option) of these. Only map[string]string is supported as a map type.
// Fields left at their type's zero value will be omitted from the query.
//
// Slice fields use repeated query parameters:
// Foo []string `q:"foo"`                       // ?foo=a&foo=b
//
// Map fields accept plain repeated key:value pairs:
// Bar map[string]string `q:"bar"`              // ?bar=k1:v1&bar=k2:v2
//
// time.Time fields support the formats RFC3339Nano, RFC3339, DateTime, Date, Unix.
// A single `format` option can be set, to define what is used for serialization, otherwise
// the default RFC3339 gets used:
// Baz time.Time `q:"baz,format:DateTime"`       // ?baz=1999-01-01 00:00:00
//
// A `required` option can be set to define that a missing value will produce an error.
// Quux string `q:"quux,required"`               // ?foo=bar --> error
func BuildQueryString(opts any) (url.Values, error) {
	optsValue := reflect.ValueOf(opts)
	if optsValue.Kind() == reflect.Ptr { //nolint: govet // won't inline this...
		optsValue = optsValue.Elem()
	}
	optsType := optsValue.Type()
	params := url.Values{}

	if optsValue.Kind() != reflect.Struct {
		panic("options type is not a struct")
	}
	for _, field := range reflect.VisibleFields(optsType) {
		fieldValue := optsValue.FieldByIndex(field.Index)
		qTag := field.Tag.Get("q")
		if qTag == "" {
			panic(fmt.Sprintf(`expected %q to have a "q:"-tag`, field.Name))
		}
		key, format, required := parseQTag(qTag)

		// if field not set, skip
		if isZero(fieldValue) {
			if required {
				// if the field is required, it can't have a zero-value
				return url.Values{}, fmt.Errorf("required query parameter [%s] not set", field.Name)
			}
			continue
		}
	loop:
		switch fieldValue.Kind() {
		case reflect.Ptr: //nolint: govet // won't inline this...
			fieldValue = fieldValue.Elem()
			goto loop
		case reflect.String:
			params.Add(key, fieldValue.String())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			params.Add(key, strconv.FormatInt(fieldValue.Int(), 10))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			params.Add(key, strconv.FormatUint(fieldValue.Uint(), 10))
		case reflect.Bool:
			params.Add(key, strconv.FormatBool(fieldValue.Bool()))
		case reflect.Float32, reflect.Float64:
			params.Add(key, strconv.FormatFloat(fieldValue.Float(), 'g', -1, fieldValue.Type().Bits()))
		case reflect.Slice:
			var values []string
			for i := range fieldValue.Len() {
				values = append(values, serializeValue(fieldValue.Index(i), format))
			}
			params[key] = append(params[key], values...)
		case reflect.Struct:
			if fieldValue.Type() == reflect.TypeFor[time.Time]() {
				params.Add(key, serializeValue(fieldValue, format))
			} else if m := fieldValue.MethodByName("AsPointer"); m.IsValid() {
				// Option[T] — unwrap via AsPointer
				results := m.Call(nil)
				if len(results) == 1 && results[0].Kind() == reflect.Ptr && !results[0].IsNil() { //nolint:govet // won't inline this...
					params.Add(key, serializeValue(results[0].Elem(), format))
				}
			} else {
				// defense in depth: already handled by isZero function
				panic("for structs only time.Time and implementers of isZeroer are supported")
			}
		case reflect.Map:
			if fieldValue.Type().Key().Kind() == reflect.String && fieldValue.Type().Elem().Kind() == reflect.String {
				keys := make([]string, 0, fieldValue.Len())
				for _, k := range fieldValue.MapKeys() {
					keys = append(keys, k.String())
				}
				slices.Sort(keys)
				for _, k := range keys {
					value := fieldValue.MapIndex(reflect.ValueOf(k)).String()
					params.Add(key, k+":"+value)
				}
			}
		}
	}
	return params, nil
}

// serializeValue converts a reflect.Value to its string representation for query parameters.
func serializeValue(v reflect.Value, format string) string {
	// Dereference pointers.
	for v.Kind() == reflect.Ptr { //nolint:govet // won't inline this...
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, v.Type().Bits())
	case reflect.Struct:
		if v.Type() == reflect.TypeFor[time.Time]() {
			t := v.Interface().(time.Time)
			switch format {
			case "":
				return t.Format(time.RFC3339)
			case unixFormat:
				return strconv.FormatInt(t.Unix(), 10)
			default:
				tf, exists := timeFormats[format]
				if !exists {
					panic("unknown time format " + format)
				}
				return t.Format(tf)
			}
		}
		// For Option[T] and similar wrapper types: try to unwrap via AsPointer() method.
		if m := v.MethodByName("AsPointer"); m.IsValid() {
			results := m.Call(nil)
			if len(results) == 1 && results[0].Kind() == reflect.Ptr && !results[0].IsNil() { //nolint:govet // won't inline this...
				return serializeValue(results[0].Elem(), format)
			}
		}
	}
	return fmt.Sprintf("%v", v.Interface())
}
