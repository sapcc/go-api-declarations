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

//Package testhelper reimplements parts of github.com/gophercloud/gophercloud/testhelper
//to avoid a Gophercloud dependency in this module.
package testhelper

import (
	"encoding/json"
	"reflect"
	"testing"
)

func AssertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func CheckDeepEquals(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected value: %#v", expected)
		t.Errorf(" but got value: %#v", actual)
	}
}

func CheckJSONEquals(t *testing.T, expectedJSON string, actual interface{}) {
	t.Helper()
	actualJSON, err := json.Marshal(actual)
	if err != nil {
		t.Errorf("cannot marshal actual value to JSON: %s", err.Error())
		return
	}
	if !reflect.DeepEqual(decodeJSON(t, []byte(expectedJSON)), decodeJSON(t, actualJSON)) {
		t.Errorf("expected JSON: %s", expectedJSON)
		t.Errorf(" but got JSON: %s", actualJSON)
	}
}

func decodeJSON(t *testing.T, buf []byte) (data interface{}) {
	t.Helper()
	err := json.Unmarshal(buf, &data)
	if err != nil {
		t.Fatalf("cannot unmarshal JSON value: %s", err.Error())
	}
	return
}
