// Copyright 2018 SAP SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package limesresources

import (
	"testing"
	"time"

	th "github.com/sapcc/go-api-declarations/internal/testhelper"
	"github.com/sapcc/go-api-declarations/limes"
)

var projectServicesMockJSON = `
	[
		{
			"type": "shared",
			"area": "shared",
			"resources": [
				{
					"name": "capacity",
					"unit": "B",
					"per_az": {
						"az-one": {
							"quota": 6,
							"committed": {
								"1 year": 3
							},
							"usage": 2
						},
						"az-two": {
							"quota": 4,
							"usage": 0
						}
					},
					"quota": 10,
					"usable_quota": 11,
					"usage": 2
				},
				{
					"name": "things",
					"per_az": {
						"any": {
							"usage": 2
						}
					},
					"quota": 10,
					"usable_quota": 10,
					"usage": 2
				}
			],
			"scraped_at": 22
		}
	]
`

var projectResourcesMockJSON = `
	[
		{
			"name": "capacity",
			"unit": "B",
			"per_az": {
				"az-one": {
					"quota": 6,
					"committed": {
						"1 year": 3
					},
					"usage": 2
				},
				"az-two": {
					"quota": 4,
					"usage": 0
				}
			},
			"quota": 10,
			"usable_quota": 11,
			"usage": 2
		},
		{
			"name": "things",
			"quota": 10,
			"per_az": {
				"any": {
					"usage": 2
				}
			},
			"usable_quota": 10,
			"usage": 2
		}
	]
`

var projectMockResources = &ProjectResourceReports{
	"capacity": &ProjectResourceReport{
		ResourceInfo: ResourceInfo{
			Name: "capacity",
			Unit: limes.UnitBytes,
		},
		PerAZ: ProjectAZResourceReports{
			"az-one": {Quota: p2u64(6), Usage: 2, Committed: map[string]uint64{"1 year": 3}},
			"az-two": {Quota: p2u64(4), Usage: 0},
		},
		Quota:       p2u64(10),
		UsableQuota: p2u64(11),
		Usage:       2,
	},
	"things": &ProjectResourceReport{
		ResourceInfo: ResourceInfo{
			Name: "things",
		},
		PerAZ: ProjectAZResourceReports{
			limes.AvailabilityZoneAny: {Usage: 2},
		},
		Quota:       p2u64(10),
		UsableQuota: p2u64(10),
		Usage:       2,
	},
}

var projectMockServices = &ProjectServiceReports{
	"shared": &ProjectServiceReport{
		ServiceInfo: limes.ServiceInfo{
			Type: "shared",
			Area: "shared",
		},
		Resources: *projectMockResources,
		ScrapedAt: p2time(22),
	},
}

func TestProjectServicesMarshall(t *testing.T) {
	th.CheckJSONEquals(t, projectServicesMockJSON, projectMockServices)
}

func TestProjectServicesUnmarshall(t *testing.T) {
	actual := &ProjectServiceReports{}
	err := actual.UnmarshalJSON([]byte(projectServicesMockJSON))
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, projectMockServices, actual)
}

func TestProjectResourcesMarshall(t *testing.T) {
	th.CheckJSONEquals(t, projectResourcesMockJSON, projectMockResources)
}

func TestProjectResourcesUnmarshall(t *testing.T) {
	actual := &ProjectResourceReports{}
	err := actual.UnmarshalJSON([]byte(projectResourcesMockJSON))
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, projectMockResources, actual)
}

func p2time(timestamp int64) *limes.UnixEncodedTime {
	t := limes.UnixEncodedTime{Time: time.Unix(timestamp, 0).UTC()}
	return &t
}
