// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package limesresources

import (
	"reflect"
	"testing"
	"time"
)

func TestParseCommitmentDurationOK(t *testing.T) {
	tests := map[string]CommitmentDuration{
		"1 year":                     {Years: 1},
		"2 years":                    {Years: 2},
		"3year,5month":               {Years: 3, Months: 5},
		"   1 days  ,\t2 seconds\n,": {Days: 1, Short: 2 * time.Second},
	}

	for input, expected := range tests {
		actual, err := ParseCommitmentDuration(input)
		if err != nil {
			t.Errorf("expected %q to parse, but got error: %s", input, err.Error())
		} else if !reflect.DeepEqual(actual, expected) {
			t.Errorf("expected %q to parse into %#v, but got %#v", input, expected, actual)
		}
	}
}

func TestParseCommitmentDurationError(t *testing.T) {
	tests := map[string]string{
		"":                `could not parse CommitmentDuration "": empty duration`,
		"0 days":          `could not parse CommitmentDuration "0 days": empty duration`,
		",,,,3 blobs":     `could not parse CommitmentDuration ",,,,3 blobs": malformed field "3 blobs"`,
		"a year,3 months": `could not parse CommitmentDuration "a year,3 months": malformed field "a year"`,
	}

	for input, expected := range tests {
		_, err := ParseCommitmentDuration(input)
		actual := ""
		if err != nil {
			actual = err.Error()
		}
		if actual != expected {
			t.Errorf("expected parse of %q to fail with %q, but got %q", input, expected, actual)
		}
	}
}

func TestSerializeCommitmentDuration(t *testing.T) {
	tests := map[string]string{
		"1 year":                     "1 year",
		"2 years":                    "2 years",
		"3year,5month":               "3 years, 5 months",
		"   1 days  ,\t2 seconds\n,": "1 day, 2 seconds",
	}

	for input, expected := range tests {
		parsed, err := ParseCommitmentDuration(input)
		if err != nil {
			t.Errorf("expected %q to parse, but got error: %s", input, err.Error())
		} else {
			actual := parsed.String()
			if actual != expected {
				t.Errorf("expected canonical representation of %q to be %q, but got %q", input, expected, actual)
			}
		}
	}
}
