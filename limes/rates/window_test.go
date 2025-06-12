// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package limesrates

import "testing"

func TestWindowSerializeRoundtrip(t *testing.T) {
	tests := map[string]string{
		"1ms":    "1ms",
		"500ms":  "500ms",
		"1000ms": "1s",
		"1s":     "1s",
		"42s":    "42s",
		"90s":    "90s",
		"120s":   "2m",
		"1m":     "1m",
		"42m":    "42m",
		"90m":    "90m",
		"120m":   "2h",
		"1h":     "1h",
		"120h":   "120h",
	}

	for input, expected := range tests {
		parsed, err := ParseWindow(input)
		if err != nil {
			t.Error(err.Error())
			continue
		}
		actual := parsed.String()
		if actual != expected {
			t.Errorf("for input %q: expected %q, got %q", input, expected, actual)
		}
	}
}
