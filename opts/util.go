// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package opts

import (
	"maps"
	"reflect"
	"slices"
	"strings"
	"time"
)

var (
	// timeFormats lists all allowed formats, except "Unix" which is handled separately.
	timeFormats = map[string]string{
		"RFC3339Nano": time.RFC3339Nano,
		"RFC3339":     time.RFC3339,
		"DateTime":    time.DateTime,
		"DateOnly":    time.DateOnly,
	}
	supportedHumanReadableFormats = strings.Join(slices.Sorted(maps.Keys(timeFormats)), ", ")
)

const (
	unixFormat           = "Unix"
	commaSeparatedFormat = "comma-separated"
)

// isZero checks if a value is the zero value for its type. For scalar values,
// it uses reflect. For structs it checks each field, except structs that
// implement the isZeroer interface. For arrays it checks each element. For
// pointer, funcs, maps and slices reflect.IsNil can be used.
func isZero(v reflect.Value) bool {
	// Check for types implementing IsZero() bool (e.g. Option[T], time.Time).
	// Guard against nil pointers whose element type has pointer-receiver IsZero.
	type isZeroer interface{ IsZero() bool }
	if v.Kind() == reflect.Ptr { //nolint:govet // won't inline this...
		if v.IsNil() {
			return true
		}
		// Dereference the pointer and check the element.
		return isZero(v.Elem())
	}
	if v.CanInterface() {
		if z, ok := v.Interface().(isZeroer); ok {
			return z.IsZero()
		}
	}

	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := range v.Len() {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for _, structField := range v.Fields() {
			z = z && isZero(structField)
		}
		return z
	default:
		z := reflect.Zero(v.Type())
		return v.Interface() == z.Interface()
	}
}
