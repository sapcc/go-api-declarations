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

func AssertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func CheckDeepEquals(t *testing.T, expected, actual any) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected value: %#v", expected)
		t.Errorf(" but got value: %#v", actual)
	}
}

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

func decodeJSON(t *testing.T, buf []byte) (data any) {
	t.Helper()
	err := json.Unmarshal(buf, &data)
	if err != nil {
		t.Fatalf("cannot unmarshal JSON value: %s", err.Error())
	}
	return
}
