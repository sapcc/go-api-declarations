// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package limes

import (
	"encoding/json"
	"time"
)

// UnixEncodedTime is a time.Time that marshals into JSON as a UNIX timestamp.
//
// This is a single-member struct instead of a newtype because the former
// enables directly calling time.Time methods on this type, e.g. t.String()
// instead of time.Time(t).String().
type UnixEncodedTime struct {
	time.Time
}

// MarshalJSON implements the json.Marshaler interface.
func (t UnixEncodedTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Unix())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *UnixEncodedTime) UnmarshalJSON(buf []byte) error {
	var tst int64
	err := json.Unmarshal(buf, &tst)
	if err == nil {
		t.Time = time.Unix(tst, 0).UTC()
	}
	return err
}

// RFC3339EncodedTime is a time.Time that marshals into JSON as RFC3339 timestamp.
type RFC3339EncodedTime struct {
	time.Time
}

// MarshalJSON implements the json.Marshaler interface.
func (t RFC3339EncodedTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Format(time.RFC3339))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *RFC3339EncodedTime) UnmarshalJSON(buf []byte) error {
	var s string
	err := json.Unmarshal(buf, &s)
	if err != nil {
		return err
	}
	t.Time, err = time.Parse(time.RFC3339, s)
	return err
}

// RFC3339NanoEncodedTime is a time.Time that marshals into JSON as RFC3339Nano timestamp.
type RFC3339NanoEncodedTime struct {
	time.Time
}

// MarshalJSON implements the json.Marshaler interface.
func (t RFC3339NanoEncodedTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Format(time.RFC3339Nano))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *RFC3339NanoEncodedTime) UnmarshalJSON(buf []byte) error {
	var s string
	err := json.Unmarshal(buf, &s)
	if err != nil {
		return err
	}
	t.Time, err = time.Parse(time.RFC3339Nano, s)
	return err
}
