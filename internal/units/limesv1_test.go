// SPDX-FileCopyrightText: 2017-2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package units

import (
	"testing"

	"go.xyrillian.de/gg/assert"
)

func TestParseInUnit(t *testing.T) {
	// This executes all the examples from the docstring of func ParseInUnit().
	// Since ParseInUnit() calls ConvertTo(), this also provides coverage for ConvertTo().

	value, err := ParseInUnit(UnitMebibytes, "10 MiB")
	assert.ErrEqual(t, err, nil)
	assert.Equal(t, value, 10)

	value, err = ParseInUnit(UnitMebibytes, "10 GiB")
	assert.ErrEqual(t, err, nil)
	assert.Equal(t, value, 10240)

	_, err = ParseInUnit(UnitMebibytes, "10 KiB")
	assert.ErrEqual(t, err, `value "10 KiB" cannot be represented as integer number of MiB`)

	_, err = ParseInUnit(UnitMebibytes, "10")
	assert.ErrEqual(t, err, `cannot convert value "10" to MiB because units are incompatible`)

	value, err = ParseInUnit(UnitNone, "42")
	assert.ErrEqual(t, err, nil)
	assert.Equal(t, value, 42)

	_, err = ParseInUnit(UnitNone, "42 MiB")
	assert.ErrEqual(t, err, `cannot convert value "42 MiB" to <count> because units are incompatible`)
}

func TestValueWithUnitToString(t *testing.T) {
	// check behavior LimesV1ValueWithUnit.String() esp. with non-standard units and integer overflows

	v := LimesV1ValueWithUnit{
		Value: 128,
		Unit:  UnitKibibytes,
	}
	assert.Equal(t, v.String(), "128 KiB")

	v = LimesV1ValueWithUnit{
		Value: 128,
		Unit:  mustMultiply(UnitKibibytes, 32),
	}
	assert.Equal(t, v.String(), "4 MiB") // uses nice formatting, i.e. neither "128 x 32 KiB" nor "4096 KiB"

	v = LimesV1ValueWithUnit{
		// this value is equal to 2^75 bytes and overflows type Amount
		Value: 32768,
		Unit:  UnitExbibytes,
	}
	assert.Equal(t, v.String(), "32768 EiB") // printed via fallback logic

	v = LimesV1ValueWithUnit{
		// same value, but this time the fallback printing logic is more obvious
		Value: 16384,
		Unit:  mustMultiply(UnitExbibytes, 2),
	}
	assert.Equal(t, v.String(), "16384 x 2 EiB")
}
