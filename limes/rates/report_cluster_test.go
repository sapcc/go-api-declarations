// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

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

func TestClusterServicesOnlyRatesMarshal(t *testing.T) {
	th.CheckJSONEquals(t, clusterServicesOnlyRatesMockJSON, clusterServicesOnlyRates)
}

func TestClusterServicesOnlyRatesUnmarshal(t *testing.T) {
	actual := &ClusterServiceReports{}
	err := actual.UnmarshalJSON([]byte(clusterServicesOnlyRatesMockJSON))
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, clusterServicesOnlyRates, actual)
}
