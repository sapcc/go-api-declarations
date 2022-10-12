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

package limesrates

import (
	"testing"

	th "github.com/sapcc/go-api-declarations/internal/testhelper"
	"github.com/sapcc/go-api-declarations/limes"
)

var clusterServicesOnlyRatesMockJSON = `
	[
		{
			"type": "compute",
			"area": "compute",
			"rates": [
				{
					"name": "service/shared/objects:create",
					"limit": 5000,
					"window": "1s"
				}
			]
		}
	]
`

var clusterServicesOnlyRates = &ClusterServiceReports{
	"compute": &ClusterServiceReport{
		ServiceInfo: limes.ServiceInfo{
			Type: "compute",
			Area: "compute",
		},
		Rates: ClusterRateReports{
			"service/shared/objects:create": {
				RateInfo: RateInfo{Name: "service/shared/objects:create"},
				Limit:    5000,
				Window:   1 * WindowSeconds,
			},
		},
	},
}

func p2i64(val int64) *int64 {
	return &val
}

func TestClusterServicesOnlyRatesMarshal(t *testing.T) {
	th.CheckJSONEquals(t, clusterServicesOnlyRatesMockJSON, clusterServicesOnlyRates)
}

func TestClusterServicesOnlyRatesUnmarshal(t *testing.T) {
	actual := &ClusterServiceReports{}
	err := actual.UnmarshalJSON([]byte(clusterServicesOnlyRatesMockJSON))
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, clusterServicesOnlyRates, actual)
}
