/*******************************************************************************
*
* Copyright 2022 SAP SE
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You should have received a copy of the License along with this
* program. If not, you may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*
*******************************************************************************/

package limes

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMarshalTime(t *testing.T) {
	tst := time.Unix(23, 0).UTC().Add(300 * time.Millisecond) //test that marshalling ignores the subsecond part
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
