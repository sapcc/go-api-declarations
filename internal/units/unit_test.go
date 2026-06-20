// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package units

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"testing"

	"go.xyrillian.de/gg/assert"
)

func TestParseUnit(t *testing.T) {
	test := func(input string, expected Unit) {
		t.Run(input, func(t *testing.T) {
			// parsing of Unit strings happens implicitly in context-specific interfaces

			// test parsing from SQL string
			var u1 Unit
			err := u1.Scan(input)
			assert.ErrEqual(t, err, nil)
			assert.Equal(t, u1, expected)

			// test parsing from SQL bytestring
			var u2 Unit
			err = u2.Scan([]byte(input))
			assert.ErrEqual(t, err, nil)
			assert.Equal(t, u2, expected)

			// test parsing from JSON string
			buf, err := json.Marshal(input)
			assert.ErrEqual(t, err, nil)
			var u3 Unit
			err = json.Unmarshal(buf, &u3)
			assert.ErrEqual(t, err, nil)
			assert.Equal(t, u3, expected)
		})
	}

	test("", UnitNone)
	test("piece", UnitPiece)
	test("1000 piece", mustMultiply(UnitPiece, 1000))
	test("KiB", UnitKibibytes)
	test("1000 B", mustMultiply(UnitBytes, 1000))
	test("1024 B", UnitKibibytes)
}

func TestSerializeUnit(t *testing.T) {
	test := func(unit Unit, expected string) {
		t.Run(fmt.Sprintf("%#v", unit), func(t *testing.T) {
			// test serialization through fmt.Stringer
			assert.Equal(t, unit.String(), expected)

			// test serialization as SQL string
			value, err := unit.Value()
			assert.ErrEqual(t, err, nil)
			assert.Equal(t, value, driver.Value(expected))

			// test serialization as JSON string
			buf, err := json.Marshal(unit)
			assert.ErrEqual(t, err, nil)
			assert.Equal(t, string(buf), fmt.Sprintf("%q", expected))
		})
	}

	test(UnitNone, "")
	test(UnitPiece, "piece")
	test(mustMultiply(UnitPiece, 1000), "1000 piece")
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
	for _, base := range []Unit{UnitPiece, UnitBytes, UnitGibibytes} {
		u, err := base.MultiplyBy(1)
		assert.ErrEqual(t, err, nil)
		assert.Equal(t, u, base)
	}
}

func mustMultiply(unit Unit, factor uint64) Unit {
	result, err := unit.MultiplyBy(factor)
	if err != nil {
		panic(err.Error())
	}
	return result
}
