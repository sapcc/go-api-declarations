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

	th "github.com/sapcc/go-api-declarations/internal/testhelper"
	"github.com/sapcc/go-api-declarations/limes"
)

var clusterServicesMockJSON = `
	[
		{
			"type": "compute",
			"area": "compute",
			"resources": [
				{
					"name": "cores",
					"capacity": 500,
					"per_availability_zone": [
						{
							"name": "az-one",
							"capacity": 250,
							"usage": 70
						},
						{
							"name": "az-two",
							"capacity": 250,
							"usage": 30
						}
					],
					"domains_quota": 200,
					"usage": 100
				},
				{
					"name": "ram",
					"unit": "MiB",
					"capacity": 204800,
					"domains_quota": 102400,
					"usage": 40800
				}
			],
			"max_scraped_at": 1539024049,
			"min_scraped_at": 1539023764
		}
	]
`

var clusterResourcesMockJSON = `
	[
		{
			"name": "cores",
			"capacity": 500,
			"per_availability_zone": [
				{
					"name": "az-one",
					"capacity": 250,
					"usage": 70
				},
				{
					"name": "az-two",
					"capacity": 250,
					"usage": 30
				}
			],
			"domains_quota": 200,
			"usage": 100
		},
		{
			"name": "ram",
			"unit": "MiB",
			"capacity": 204800,
			"domains_quota": 102400,
			"usage": 40800
		}
	]
`

var clusterMockResources = &ClusterResourceReports{
	"cores": &ClusterResourceReport{
		ResourceInfo: ResourceInfo{
			Name: "cores",
		},
		Capacity: &coresCap,
		CapacityPerAZ: ClusterAvailabilityZoneReports{
			"az-one": {
				Name:     "az-one",
				Capacity: 250,
				Usage:    70,
			},
			"az-two": {
				Name:     "az-two",
				Capacity: 250,
				Usage:    30,
			},
		},
		DomainsQuota: p2u64(200),
		Usage:        100,
	},
	"ram": &ClusterResourceReport{
		ResourceInfo: ResourceInfo{
			Name: "ram",
			Unit: limes.UnitMebibytes,
		},
		Capacity:     &ramCap,
		DomainsQuota: p2u64(102400),
		Usage:        40800,
	},
}

var coresCap uint64 = 500
var ramCap uint64 = 204800

var clusterMockServices = &ClusterServiceReports{
	"compute": &ClusterServiceReport{
		ServiceInfo: limes.ServiceInfo{
			Type: "compute",
			Area: "compute",
		},
		Resources:    *clusterMockResources,
		MaxScrapedAt: p2time(1539024049),
		MinScrapedAt: p2time(1539023764),
	},
}

func p2u64(val uint64) *uint64 {
	return &val
}

func TestClusterServicesMarshal(t *testing.T) {
	th.CheckJSONEquals(t, clusterServicesMockJSON, clusterMockServices)
}

func TestClusterServicesUnmarshal(t *testing.T) {
	actual := &ClusterServiceReports{}
	err := actual.UnmarshalJSON([]byte(clusterServicesMockJSON))
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, clusterMockServices, actual)
}

func TestClusterResourcesMarshal(t *testing.T) {
	th.CheckJSONEquals(t, clusterResourcesMockJSON, clusterMockResources)
}

func TestClusterResourcesUnmarshal(t *testing.T) {
	actual := &ClusterResourceReports{}
	err := actual.UnmarshalJSON([]byte(clusterResourcesMockJSON))
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, clusterMockResources, actual)
}
