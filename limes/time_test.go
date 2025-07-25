// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package limes

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMarshalTime(t *testing.T) {
	tst := time.Unix(23, 0).UTC().Add(300 * time.Millisecond) // test that marshalling ignores the subsecond part
	u := UnixEncodedTime{Time: tst}

	buf, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err.Error())
	}

	actual := string(buf)
	expected := "23"
	if actual != expected {
		t.Fatalf("expected %#v to serialize as %q, but got %q", u, expected, actual)
	}
}

func TestUnmarshalTime(t *testing.T) {
	input := "23"

	var u UnixEncodedTime
	err := json.Unmarshal([]byte(input), &u)
	if err != nil {
		t.Fatal(err.Error())
	}

	expected := time.Unix(23, 0).UTC()
	actual := u.Time
	if actual != expected {
		t.Fatalf("expected %q to deserialize into %#v, but got %#v", input, expected, actual)
	}
}
