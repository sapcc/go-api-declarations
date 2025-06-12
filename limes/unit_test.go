// SPDX-FileCopyrightText: 2017 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package limes

import "testing"

func assertConvertSuccess(t *testing.T, from, expected ValueWithUnit) {
	actual, err := from.ConvertTo(expected.Unit)
	switch {
	case err != nil:
		t.Errorf("unexpected error when converting %s to %s: %v", from.String(), string(expected.Unit), err)
	case actual != expected:
		t.Errorf("error when converting %s: expected %s, got %s", from.String(), expected.String(), actual.String())
	}
}

func assertConvertError(t *testing.T, from ValueWithUnit, to Unit, expectedError string) {
	_, err := from.ConvertTo(to)
	switch {
	case err == nil:
		t.Errorf("expected error when converting %s to %s, but found err == nil", from.String(), string(to))
	case err.Error() != expectedError:
		t.Errorf("unexpected error when converting %s to %s", from.String(), string(to))
		t.Logf("  expected: %s", expectedError)
		t.Logf("    actual: %s", err.Error())
	}
}

func Test_ValueWithUnit_ConvertTo(t *testing.T) {
	// happy cases
	assertConvertSuccess(t, ValueWithUnit{5, UnitMebibytes}, ValueWithUnit{5 << 20, UnitBytes})
	assertConvertSuccess(t, ValueWithUnit{5 << 20, UnitBytes}, ValueWithUnit{5, UnitMebibytes})
	assertConvertSuccess(t, ValueWithUnit{42, UnitBytes}, ValueWithUnit{42, UnitBytes})

	// failure cases
	assertConvertError(t, ValueWithUnit{5, UnitMebibytes}, UnitNone,
		"cannot convert value from MiB to <count> because units are incompatible",
	)
	assertConvertError(t, ValueWithUnit{42, UnitBytes}, UnitMebibytes,
		"value of 42 B cannot be represented as integer number of MiB",
	)
}
