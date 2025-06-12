// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package castellum

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

// vendored minimal version of https://github.com/sapcc/go-bits/blob/master/assert/assert.go
// to avoid cyclic dependency
func checkDeepEqual(t *testing.T, variable string, actual, expected any) {
	t.Helper()
	if reflect.DeepEqual(actual, expected) {
		return
	}

	//NOTE: We HAVE TO use %#v here, even if it's verbose. Every other generic
	// formatting directive will not correctly distinguish all values, and thus
	// possibly render empty diffs on failure. For example,
	//
	//	fmt.Sprintf("%+v\n", []string{})    == "[]\n"
	//	fmt.Sprintf("%+v\n", []string(nil)) == "[]\n"
	//
	t.Error("assert.DeepEqual failed for " + variable)
	t.Logf("\texpected = %#v\n", expected)
	t.Logf("\t  actual = %#v\n", actual)
}

func TestUsageValuesEncodingDecoding(t *testing.T) {
	testCases := []struct {
		UsageValues  UsageValues
		SQLEncoding  string
		JSONEncoding string
	}{
		{
			UsageValues:  UsageValues{SingularUsageMetric: 42.1},
			SQLEncoding:  `{"singular":42.1}`,
			JSONEncoding: `42.1`,
		},
		{
			UsageValues:  UsageValues{"foo": 0},
			SQLEncoding:  `{"foo":0}`,
			JSONEncoding: `{"foo":0}`,
		},
		{
			UsageValues:  UsageValues{"foo": 0, "bar": 23},
			SQLEncoding:  `{"bar":23,"foo":0}`,
			JSONEncoding: `{"bar":23,"foo":0}`,
		},
		{
			UsageValues:  UsageValues{"foo": 0, SingularUsageMetric: 42.1},
			SQLEncoding:  `{"foo":0,"singular":42.1}`,
			JSONEncoding: `{"foo":0,"singular":42.1}`,
		},
	}

	for idx, tc := range testCases {
		indexed := func(task string) string {
			return fmt.Sprintf("%s no. %d/%d", task, idx+1, len(testCases))
		}

		// check encoding into SQL
		actualSQLEncoding, err := tc.UsageValues.Value()
		if err == nil {
			checkDeepEqual(t, indexed("SQLEncoding"), actualSQLEncoding, driver.Value(tc.SQLEncoding))
		} else {
			t.Errorf("SQL encoding of %#v failed: %v", tc.UsageValues, err.Error())
		}

		// check decoding from SQL
		var actualDecoded UsageValues
		err = actualDecoded.Scan(tc.SQLEncoding)
		if err == nil {
			checkDeepEqual(t, indexed("SQLDecoded"), actualDecoded, tc.UsageValues)
		} else {
			t.Errorf("SQL decoding of %q failed: %v", tc.SQLEncoding, err.Error())
		}

		// check encoding into JSON
		actualJSONEncoding, err := json.Marshal(tc.UsageValues)
		if err == nil {
			checkDeepEqual(t, indexed("JSONEncoding"), string(actualJSONEncoding), tc.JSONEncoding)
		} else {
			t.Errorf("JSON encoding of %#v failed: %v", tc.UsageValues, err.Error())
		}

		// check decoding from JSON
		actualDecoded = UsageValues{}
		err = json.Unmarshal([]byte(tc.JSONEncoding), &actualDecoded)
		if err == nil {
			checkDeepEqual(t, indexed("JSONDecoded"), actualDecoded, tc.UsageValues)
		} else {
			t.Errorf("JSON decoding of %q failed: %v", tc.JSONEncoding, err.Error())
		}
	}
}
