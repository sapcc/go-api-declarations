/*******************************************************************************
*
* Copyright 2024 SAP SE
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

package limesresources

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/sapcc/go-api-declarations/limes"
)

func TestParseQuotaOverrides(t *testing.T) {
	// mock implementation of getUnit callback presenting two example resources
	getUnit := func(serviceType limes.ServiceType, resourceName ResourceName) (limes.Unit, error) {
		switch fmt.Sprintf("%s/%s", serviceType, resourceName) {
		case "unittest/capacity":
			return limes.UnitBytes, nil
		case "unittest/things":
			return limes.UnitNone, nil
		default:
			return limes.UnitUnspecified, fmt.Errorf("%s/%s is not a valid resource", serviceType, resourceName)
		}
	}

	// test successful parsing
	buf := []byte(`
		{
			"domain-one": {
				"project-one": {
					"unittest": {
						"things": 20,
						"capacity": "5 GiB"
					}
				}
			}
		}
	`)
	result, errs := ParseQuotaOverrides(buf, getUnit)
	assertDeepEqual(t, "errors", errorsToStrings(errs), []string{})
	assertDeepEqual(t, "result", result, map[string]map[string]map[limes.ServiceType]map[ResourceName]uint64{
		"domain-one": {
			"project-one": {
				"unittest": {
					"things":   20,
					"capacity": 5 << 30, // capacity is in unit "B"
				},
			},
		},
	})

	// test parsing errors
	buf = []byte(`
		{
			"domain-one": {
				"project1": {
					"unittest": {
						"capacity": [ 1, "GiB" ]
					}
				},
				"project2": {
					"unittest": {
						"things": "50 GiB"
					}
				},
				"project3": {
					"unittest": {
						"unknown-resource": 10
					}
				},
				"project4": {
					"unknown-service": {
						"items": 10
					}
				}
			}
		}
	`)
	buf = bytes.ReplaceAll(buf, []byte("\t"), []byte("  "))
	_, errs = ParseQuotaOverrides(buf, getUnit)
	assertDeepEqual(t, "errors", errorsToStrings(errs), []string{
		`expected string field for unittest/capacity, but got "[ 1, \"GiB\" ]"`,
		`expected uint64 value for unittest/things, but got "\"50 GiB\""`,
		`unittest/unknown-resource is not a valid resource`,
		`unknown-service/items is not a valid resource`,
	})
}

func errorsToStrings(errs []error) []string {
	result := make([]string, len(errs))
	for idx, err := range errs {
		result[idx] = err.Error()
	}
	sort.Strings(result) // map iteration order in ParseQuotaOverrides() is not deterministic
	return result
}

// Like assert.DeepEqual() from go-bits, which we cannot depend on because that would be a cyclic dependency.
func assertDeepEqual(t *testing.T, msg string, actual, expected any) {
	t.Helper()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("failed assertion for %s", msg)
		t.Logf("expected = %#v", expected)
		t.Logf("  actual = %#v", actual)
	}
}
