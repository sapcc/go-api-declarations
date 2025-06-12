// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package liquid

import "testing"

func TestOvercommitFactor(t *testing.T) {
	check := func(f OvercommitFactor, raw, effective uint64) {
		actualEffective := f.ApplyTo(raw)
		if actualEffective != effective {
			t.Errorf("expected (%g).ApplyTo(%d) = %d, but got %d", f, raw, effective, actualEffective)
		}
		actualRaw := f.ApplyInReverseTo(effective)
		if actualRaw != raw {
			t.Errorf("expected (%g).ApplyInReverseTo(%d) = %d, but got %d", f, effective, raw, actualRaw)
		}
	}

	check(0, 42, 42)
	check(1, 42, 42)
	check(1.2, 42, 50)

	// ApplyTo is pretty straightforward, but I'd like some more test coverage for ApplyInReverseTo
	for _, factor := range []OvercommitFactor{0, 1, 1.1, 1.2, 1.5, 2, 2.5, 3, 4} {
		for raw := range uint64(100) {
			check(factor, raw, factor.ApplyTo(raw))
		}
	}
}
