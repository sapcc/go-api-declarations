// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

// Package testhelper reimplements parts of github.com/gophercloud/gophercloud/testhelper
// to avoid a Gophercloud dependency in this module.
package testhelper

import (
	"encoding/json"
	"reflect"
	"testing"
)

// AssertNoErr fails the test if the provided error is not nil.
func AssertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

// AssertErr fails the test if the provided error is nil.
func AssertErr(t *testing.T, expected string, actual error) {
	t.Helper()
	switch {
	case actual == nil:
		t.Errorf("expected error %q, but got no error", expected)
	case actual.Error() != expected:
		t.Errorf("expected error: %s", expected)
		t.Errorf(" but got error: %s", actual.Error())
	}
}

// CheckEquals fails the test if a simple equal check between expected and actual fails.
func CheckEquals[V comparable](t *testing.T, expected, actual V) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected value: %#v", expected)
		t.Errorf(" but got value: %#v", actual)
	}
}

// CheckDeepEquals fails the test if a reflect.DeepEqual check between expected and actual fails.
func CheckDeepEquals(t *testing.T, expected, actual any) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected value: %#v", expected)
		t.Errorf(" but got value: %#v", actual)
	}
}

// CheckJSONEquals fails the test if a reflect.DeepEqual check between the marshalled actual value
// and the expectedJSON fails.
func CheckJSONEquals(t *testing.T, expectedJSON string, actual any) {
	t.Helper()
	actualJSON, err := json.Marshal(actual)
	if err != nil {
		t.Errorf("cannot marshal actual value to JSON: %s", err.Error())
		return
	}
	if !reflect.DeepEqual(decodeJSON(t, []byte(expectedJSON)), decodeJSON(t, actualJSON)) {
		t.Errorf("expected JSON: %s", expectedJSON)
		t.Errorf(" but got JSON: %s", actualJSON)
	}
}

// decodeJSON is a convenience wrapper around json.Marshal and fails the test if an error occurs.
func decodeJSON(t *testing.T, buf []byte) (data any) {
	t.Helper()
	err := json.Unmarshal(buf, &data)
	if err != nil {
		t.Fatalf("cannot unmarshal JSON value: %s", err.Error())
	}
	return
}
