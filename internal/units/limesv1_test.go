// SPDX-FileCopyrightText: 2017-2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package units

import (
	"testing"

	th "github.com/sapcc/go-api-declarations/internal/testhelper"
)

func TestParseInUnit(t *testing.T) {
	// This executes all the examples from the docstring of func ParseInUnit().
	// Since ParseInUnit() calls ConvertTo(), this also provides coverage for ConvertTo().

	value, err := ParseInUnit(UnitMebibytes, "10 MiB")
	th.AssertNoErr(t, err)
	th.CheckEquals(t, 10, value)

	value, err = ParseInUnit(UnitMebibytes, "10 GiB")
	th.AssertNoErr(t, err)
	th.CheckEquals(t, 10240, value)

	_, err = ParseInUnit(UnitMebibytes, "10 KiB")
	th.AssertErr(t, `value "10 KiB" cannot be represented as integer number of MiB`, err)

	_, err = ParseInUnit(UnitMebibytes, "10")
	th.AssertErr(t, `cannot convert value "10" to MiB because units are incompatible`, err)

	value, err = ParseInUnit(UnitNone, "42")
	th.AssertNoErr(t, err)
	th.CheckEquals(t, 42, value)

	_, err = ParseInUnit(UnitNone, "42 MiB")
	th.AssertErr(t, `cannot convert value "42 MiB" to <count> because units are incompatible`, err)
}

func TestValueWithUnitToString(t *testing.T) {
	// check behavior LimesV1ValueWithUnit.String() esp. with non-standard units and integer overflows

	v := LimesV1ValueWithUnit{
		Value: 128,
		Unit:  UnitKibibytes,
	}
	th.CheckEquals(t, "128 KiB", v.String())

	v = LimesV1ValueWithUnit{
		Value: 128,
		Unit:  mustMultiply(UnitKibibytes, 32),
	}
	th.CheckEquals(t, "4 MiB", v.String()) // uses nice formatting, i.e. neither "128 x 32 KiB" nor "4096 KiB"

	v = LimesV1ValueWithUnit{
		// this value is equal to 2^75 bytes and overflows type Amount
		Value: 32768,
		Unit:  UnitExbibytes,
	}
	th.CheckEquals(t, "32768 EiB", v.String()) // printed via fallback logic

	v = LimesV1ValueWithUnit{
		// same value, but this time the fallback printing logic is more obvious
		Value: 16384,
		Unit:  mustMultiply(UnitExbibytes, 2),
	}
	th.CheckEquals(t, "16384 x 2 EiB", v.String())
}
