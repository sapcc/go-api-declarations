// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package testhelper

import (
	"reflect"
	"strings"
	"testing"
)

// CheckFullySeparate tests whether the left-hand side and right-hand side values
// are fully separate from each other. This test fails if manipulating any part
// of one value can have an effect on the other value through sharing a
// contained pointer (or map or slice).
func CheckFullySeparate[V any](t *testing.T, lhs, rhs V) {
	t.Helper()
	refs := make(map[uintptr]any)
	walkByReference(reflect.ValueOf(lhs), func(v reflect.Value) {
		refs[uintptr(v.UnsafePointer())] = v.Interface()
	})

	walkByReference(reflect.ValueOf(rhs), func(v reflect.Value) {
		value, exists := refs[uintptr(v.UnsafePointer())]
		if exists {
			t.Errorf("CheckFullySeparate found a reused value: %#v", value)
		}
	})
}

// Walks through `v` and calls `action` once for every pointer-ish thing in it
// that can be modified in-place (i.e. every chan, func, map, pointer or slice).
func walkByReference(v reflect.Value, action func(reflect.Value)) {
	switch v.Kind() {
	case reflect.Chan, reflect.Func:
		action(v)

	case reflect.Pointer:
		action(v)
		fallthrough
	case reflect.Interface:
		// NOTE: cannot call `action` for reflect.Interface because this kind does not allow v.UnsafePointer()
		//       (but that's okay because we will call `action` on v.Elem() in the next recursion level if necessary)
		walkByReference(v.Elem(), action)

	case reflect.Slice:
		action(v)
		fallthrough
	case reflect.Array:
		// NOTE: cannot call `action` for reflect.Array because this kind does not allow v.UnsafePointer()
		//       (but that's okay because a fixed-size array member of e.g. a struct does not represent its own allocation)
		for idx := range v.Len() {
			walkByReference(v.Index(idx), action)
		}

	case reflect.Map:
		action(v)
		for _, key := range v.MapKeys() {
			walkByReference(key, action)
			walkByReference(v.MapIndex(key), action)
		}

	case reflect.Struct:
		// trying to walk through struct types with unexported fields would break;
		// use domain knowledge instead for the few relevant cases
		t := v.Type()
		switch {
		case t.PkgPath() == "github.com/majewsky/gg/option" && strings.HasPrefix(t.Name(), "Option["):
			// Option[T] can be traversed with the Unpack() method
			retvals := v.MethodByName("Unpack").Call(nil) // value, ok := optionalValue.Unpack()
			if retvals[1].Interface().(bool) == true {    // if ok {
				walkByReference(retvals[0], action) // walkByReference(value, action)
			}
			return
		case t.PkgPath() == "time" && t.Name() == "Time":
			// time.Time does not have any pointer-ish values inside of it
			return
		case t.PkgPath() == "math/big" && t.Name() == "Int":
			// *big.Int DOES have pointer-ish values inside of it (specifically, a growable slice of words),
			// but its API ensures that two *big.Int values with distinct pointers also hold distinct slices;
			// since we already checked the outer pointer before coming here, we do not need to check further
			return
		}

		for idx := range v.NumField() {
			walkByReference(v.Field(idx), action)
		}
	}
}
