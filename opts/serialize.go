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
		key, maybeTimeFormat, required := parseQTag(qTag)

		// check if format is set when required
		if typeNeedsTimeFormat(fieldValue.Type()) && maybeTimeFormat.IsNone() {
			panic(fmt.Sprintf(`time format is missing for field %q`, field.Name))
		}

		// if field not set, skip
		if canBeSkipped(fieldValue, required) {
			continue
		}
	loop:
		switch fieldValue.Kind() {
		case reflect.Ptr: //nolint: govet // won't inline this...
			fieldValue = fieldValue.Elem()
			goto loop
		// only handle non-single-values here, rest is done by serializeSingleValue()
		case reflect.Slice:
			values := make([]string, fieldValue.Len())
			for i := range fieldValue.Len() {
				values[i] = serializeSingleValue(fieldValue.Index(i), maybeTimeFormat)
			}
			params[key] = append(params[key], values...)
		case reflect.Struct:
			if fieldValue.Type() == reflect.TypeFor[time.Time]() {
				params.Add(key, serializeSingleValue(fieldValue, maybeTimeFormat))
			} else if m := fieldValue.MethodByName("AsPointer"); m.IsValid() {
				// Option[T] — unwrap via AsPointer
				results := m.Call(nil)
				if len(results) == 1 && results[0].Kind() == reflect.Ptr && !results[0].IsNil() { //nolint:govet // won't inline this...
					params.Add(key, serializeSingleValue(results[0].Elem(), maybeTimeFormat))
				}
			} else {
				// defense in depth: already handled by canBeSkipped function
				panic("for structs only implementers of isZeroer are supported")
			}
		case reflect.Map:
			keys := fieldValue.MapKeys()
			slices.SortFunc(keys, func(a, b reflect.Value) int {
				return strings.Compare(serializeSingleValue(a, maybeTimeFormat), serializeSingleValue(b, maybeTimeFormat))
			})
			for _, k := range keys {
				params.Add(key, serializeSingleValue(k, maybeTimeFormat)+":"+serializeSingleValue(fieldValue.MapIndex(k), maybeTimeFormat))
			}
		default:
			params.Add(key, serializeSingleValue(fieldValue, maybeTimeFormat))
		}
		if required && isOnlyEmptyStrings(params[key]) {
			// if the field is required, it cannot have no value (handles nil maps, slices, arrays)
			return url.Values{}, fmt.Errorf("required query parameter [%s] not set", field.Name)
		}
	}
	return params, nil
}

// serializeSingleValue converts a reflect.Value to its string representation for query parameters.
// Zero values are skipped.
func serializeSingleValue(v reflect.Value, timeFormat Option[string]) string {
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
		// handle time
		if v.Type() == reflect.TypeFor[time.Time]() {
			t := v.Interface().(time.Time)
			tf := timeFormat.UnwrapOrPanic("timeFormat should have been set")
			switch tf {
			case unixTimeFormat:
				return strconv.FormatInt(t.Unix(), 10)
			default:
				layout := nonUnixTimeFormats[tf]
				return t.Format(layout)
			}
		}
		// handle Option[T]: try to unwrap via AsPointer() method.
		if m := v.MethodByName("AsPointer"); m.IsValid() {
			results := m.Call(nil)
			if len(results) == 1 && results[0].Kind() == reflect.Ptr && !results[0].IsNil() { //nolint:govet // won't inline this...
				return serializeSingleValue(results[0].Elem(), timeFormat)
			}
		}
	}
	return fmt.Sprintf("%v", v.Interface())
}

// canBeSkipped checks if a value can be skipped for serialization into a query string.
// Required params are never skipped. Otherwise, isZero() of the value is checked.
// Special handling is applied to
// - structs (check isZero() interface)
// - slices, arrays, maps (considered zero when nil or all values are zero)
func canBeSkipped(v reflect.Value, required bool) bool {
	if required {
		return false
	}
	// check pointers
	if v.Kind() == reflect.Ptr { //nolint:govet // won't inline this...
		if v.IsNil() {
			// Guard against nil pointers whose element type has pointer-receiver IsZero.
			return true
		}
		// Dereference the pointer and check the element.
		return canBeSkipped(v.Elem(), false)
	}

	// check for types implementing IsZero() bool (e.g. Option[T], time.Time).
	if v.CanInterface() {
		type isZeroer interface{ IsZero() bool }
		if z, ok := v.Interface().(isZeroer); ok {
			return z.IsZero()
		}
	}

	switch v.Kind() {
	case reflect.Func:
		panic("functions are not supported")
	case reflect.Struct:
		panic("for structs only implementers of isZeroer are supported")
	default:
		return v.IsZero()
	}
}
