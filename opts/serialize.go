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

// BuildQueryString is a function to be used by request methods with opts structs.
// It's inspired by [gophercloud.BuildQueryString] with partially stricter behavior.
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
// will be converted into
//
//	?x_bar=AAA&lorem_ipsum=BBB
//
// On configuration errors (e.g. non-struct opts, opts with non-q-tagged fields)
// the function panics. On user errors (e.g. missing required field) an error
// is returned. On success, url.Values are returned according to the opts.
//
// This function understands and expects the same values for the "q" tag as [BuildQueryString].
// See documentation over there for details.
//
// [gophercloud.BuildQueryString]: https://pkg.go.dev/github.com/gophercloud/gophercloud/v2#BuildQueryString
//
// [option.Option]: https://pkg.go.dev/go.xyrillian.de/gg/option#Option
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
		key, timeFormat, required := parseQTag(qTag)
		// all known formats are currently timeFormats
		if _, ok := TimeFormats[timeFormat]; timeFormat != "" && timeFormat != UnixFormat && !ok {
			panic(fmt.Sprintf("unsupported time format %q; accepted: %s", timeFormat, supportedHumanReadableFormats))
		}

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
			values := make([]string, fieldValue.Len())
			for i := range fieldValue.Len() {
				values[i] = serializeValue(fieldValue.Index(i), timeFormat)
			}
			params[key] = append(params[key], values...)
		case reflect.Struct:
			if fieldValue.Type() == reflect.TypeFor[time.Time]() {
				if timeFormat == "" {
					panic("time format is missing")
				}
				params.Add(key, serializeValue(fieldValue, timeFormat))
			} else if m := fieldValue.MethodByName("AsPointer"); m.IsValid() {
				// Option[T] — unwrap via AsPointer
				results := m.Call(nil)
				if len(results) == 1 && results[0].Kind() == reflect.Ptr && !results[0].IsNil() { //nolint:govet // won't inline this...
					params.Add(key, serializeValue(results[0].Elem(), timeFormat))
				}
			} else {
				// defense in depth: already handled by isZero function
				panic("for structs only time.Time and implementers of isZeroer are supported")
			}
		case reflect.Map:
			keys := fieldValue.MapKeys()
			slices.SortFunc(keys, func(a, b reflect.Value) int {
				return strings.Compare(serializeValue(a, ""), serializeValue(b, ""))
			})
			for _, k := range keys {
				params.Add(key, serializeValue(k, "")+":"+serializeValue(fieldValue.MapIndex(k), ""))
			}
		}
	}
	return params, nil
}

// serializeValue converts a reflect.Value to its string representation for query parameters.
func serializeValue(v reflect.Value, timeFormat string) string {
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
			switch timeFormat {
			case UnixFormat:
				return strconv.FormatInt(t.Unix(), 10)
			default:
				tf, exists := TimeFormats[timeFormat]
				// defense in depth: should be checked outside of this function (when no value is set)
				if !exists {
					panic("unknown time format " + timeFormat)
				}
				return t.Format(tf)
			}
		}
		// For Option[T] and similar wrapper types: try to unwrap via AsPointer() method.
		if m := v.MethodByName("AsPointer"); m.IsValid() {
			results := m.Call(nil)
			if len(results) == 1 && results[0].Kind() == reflect.Ptr && !results[0].IsNil() { //nolint:govet // won't inline this...
				return serializeValue(results[0].Elem(), timeFormat)
			}
		}
	}
	return fmt.Sprintf("%v", v.Interface())
}
