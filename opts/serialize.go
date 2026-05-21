// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package opts

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

/*
BuildQueryString is an internal function to be used by request methods in
individual resource packages.

It accepts a tagged structure and expands it into a URL struct. Field names are
converted into query parameters based on a "q" tag. For example:

	type struct Something {
	   Bar string `q:"x_bar"`
	   Baz int    `q:"lorem_ipsum"`
	}
	instance := Something{
	   Bar: "AAA",
	   Baz: "BBB",
	}

will be converted into "?x_bar=AAA&lorem_ipsum=BBB".

The struct's fields may be strings, integers, slices, or boolean values. Fields
left at their type's zero value will be omitted from the query.

Slice are handled in one of two ways:

	type struct Something {
	   Bar []string `q:"bar"` // E.g. ?bar=1&bar=2
	   Baz []int    `q:"baz" format="comma-separated"` // E.g. ?baz=1,2
	}
*/
func BuildQueryString(opts any) (*url.URL, error) {
	optsValue := reflect.ValueOf(opts)
	if optsValue.Kind() == reflect.Ptr { //nolint: govet // won't inline this...
		optsValue = optsValue.Elem()
	}

	optsType := reflect.TypeOf(opts)
	if optsType.Kind() == reflect.Ptr { //nolint: govet // won't inline this...
		optsType = optsType.Elem()
	}

	params := url.Values{}

	if optsValue.Kind() == reflect.Struct {
		for i := range optsValue.NumField() {
			v := optsValue.Field(i)
			f := optsType.Field(i)
			qTag := f.Tag.Get("q")

			// if the field has a 'q' tag, it goes in the query string
			if qTag != "" {
				tags := strings.Split(qTag, ",")

				// if the field is set, add it to the slice of query pieces
				if !isZero(v) {
					format := f.Tag.Get("format")
				loop:
					switch v.Kind() {
					case reflect.Ptr: //nolint: govet // won't inline this...
						v = v.Elem()
						goto loop
					case reflect.String:
						params.Add(tags[0], v.String())
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						params.Add(tags[0], strconv.FormatInt(v.Int(), 10))
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						params.Add(tags[0], strconv.FormatUint(v.Uint(), 10))
					case reflect.Bool:
						params.Add(tags[0], strconv.FormatBool(v.Bool()))
					case reflect.Float32, reflect.Float64:
						params.Add(tags[0], strconv.FormatFloat(v.Float(), 'g', -1, v.Type().Bits()))
					case reflect.Slice:
						var values []string
						for i := range v.Len() {
							values = append(values, serializeValue(v.Index(i), format))
						}
						if format == "comma-separated" {
							params.Add(tags[0], strings.Join(values, ","))
						} else {
							params[tags[0]] = append(params[tags[0]], values...)
						}
					case reflect.Struct:
						if v.Type() == reflect.TypeFor[time.Time]() {
							params.Add(tags[0], serializeValue(v, format))
						} else if m := v.MethodByName("AsPointer"); m.IsValid() {
							// Option[T] — unwrap via AsPointer
							results := m.Call(nil)
							if len(results) == 1 && results[0].Kind() == reflect.Ptr && !results[0].IsNil() { //nolint:govet // won't inline this...
								params.Add(tags[0], serializeValue(results[0].Elem(), format))
							}
						} else {
							// Plain nested struct — serialize with dot-prefix
							serializeStruct(v, tags[0], params)
						}
					case reflect.Map:
						if v.Type().Key().Kind() == reflect.String && v.Type().Elem().Kind() == reflect.String {
							keys := make([]string, 0, v.Len())
							for _, k := range v.MapKeys() {
								keys = append(keys, k.String())
							}
							slices.Sort(keys)
							var s []string
							for _, k := range keys {
								value := v.MapIndex(reflect.ValueOf(k)).String()
								s = append(s, fmt.Sprintf("'%s':'%s'", k, value))
							}
							params.Add(tags[0], fmt.Sprintf("{%s}", strings.Join(s, ", ")))
						}
					}
				} else {
					// if the field has a 'required' tag, it can't have a zero-value
					if requiredTag := f.Tag.Get("required"); requiredTag == "true" {
						return &url.URL{}, fmt.Errorf("required query parameter [%s] not set", f.Name)
					}
				}
			}
		}

		return &url.URL{RawQuery: params.Encode()}, nil
	}
	// Return an error if the underlying type of 'opts' isn't a struct.
	return nil, errors.New("options type is not a struct")
}

// serializeStruct serializes a struct's q-tagged fields as dot-prefixed query parameters.
func serializeStruct(v reflect.Value, prefix string, params url.Values) {
	vType := v.Type()
	for i := range v.NumField() {
		fv := v.Field(i)
		f := vType.Field(i)
		qTag := f.Tag.Get("q")
		if qTag == "" {
			continue
		}
		key := prefix + "." + strings.SplitN(qTag, ",", 2)[0]
		if !isZero(fv) {
			format := f.Tag.Get("format")
			// Dereference pointer
			for fv.Kind() == reflect.Ptr { //nolint:govet // won't inline this...
				fv = fv.Elem()
			}
			params.Add(key, serializeValue(fv, format))
		}
	}
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
