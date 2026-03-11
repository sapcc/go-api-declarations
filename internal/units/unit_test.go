// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package units

import (
	"encoding/json"
	"fmt"
	"testing"

	th "github.com/sapcc/go-api-declarations/internal/testhelper"
)

func TestParseUnit(t *testing.T) {
	test := func(input string, expected Unit) {
		t.Run(input, func(t *testing.T) {
			// parsing of Unit strings happens implicitly in context-specific interfaces

			// test parsing from SQL string
			var u1 Unit
			err := u1.Scan(input)
			th.AssertNoErr(t, err)
			th.CheckDeepEquals(t, expected, u1)

			// test parsing from SQL bytestring
			var u2 Unit
			err = u2.Scan([]byte(input))
			th.AssertNoErr(t, err)
			th.CheckDeepEquals(t, expected, u2)

			// test parsing from JSON string
			buf, err := json.Marshal(input)
			th.AssertNoErr(t, err)
			var u3 Unit
			err = json.Unmarshal(buf, &u3)
			th.AssertNoErr(t, err)
			th.CheckDeepEquals(t, expected, u3)
		})
	}

	test("", UnitNone)
	test("KiB", UnitKibibytes)
	test("1000 B", mustMultiply(UnitBytes, 1000))
	test("1024 B", UnitKibibytes)
}

func TestSerializeUnit(t *testing.T) {
	test := func(unit Unit, expected string) {
		t.Run(fmt.Sprintf("%#v", unit), func(t *testing.T) {
			// test serialization through fmt.Stringer
			th.CheckDeepEquals(t, expected, unit.String())

			// test serialization as SQL string
			value, err := unit.Value()
			th.AssertNoErr(t, err)
			th.CheckDeepEquals(t, expected, value)

			// test serialization as JSON string
			buf, err := json.Marshal(unit)
			th.AssertNoErr(t, err)
			th.CheckDeepEquals(t, fmt.Sprintf("%q", expected), string(buf))
		})
	}

	test(UnitNone, "")
	test(UnitKibibytes, "KiB")
	test(mustMultiply(UnitBytes, 1000), "1000 B")
	test(mustMultiply(UnitBytes, 1024), "KiB")
	test(mustMultiply(UnitBytes, 2048), "2 KiB")
	test(mustMultiply(UnitBytes, 1<<34), "16 GiB") // test that it chooses the optimal unit (not KiB or MiB)
}

func TestUnitMultiplyBy(t *testing.T) {
	// Basic behavior of MultiplyBy is already implicitly covered in the tests above.
	// This test checks some interesting corner cases.

	// multiply by 1 produces exactly equal instances
	u, err := UnitBytes.MultiplyBy(1)
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, UnitBytes, u)

	u, err = UnitGibibytes.MultiplyBy(1)
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, UnitGibibytes, u)
}

func mustMultiply(unit Unit, factor uint64) Unit {
	result, err := unit.MultiplyBy(factor)
	if err != nil {
		panic(err.Error())
	}
	return result
}
