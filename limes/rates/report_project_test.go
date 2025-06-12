// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package limesrates

import (
	"testing"
	"time"

	th "github.com/sapcc/go-api-declarations/internal/testhelper"
	"github.com/sapcc/go-api-declarations/limes"
)

var projectServicesRateLimitMockJSON = `
	[
		{
			"type": "shared",
			"area": "shared",
			"rates": [
				{
					"name": "services/swift/account/container/object:create",
					"limit": 1000,
					"window": "1s"
				},
				{
					"name": "services/swift/account/container/object:delete",
					"limit": 1000,
					"window": "1s"
				},
				{
					"name": "services/swift/account:create",
					"limit": 10,
					"window": "1m"
				}
			],
			"scraped_at": 22
		}
	]
`

var projectServicesRateLimitDeviatingFromDefaultsMockJSON = `
	[
		{
			"type": "shared",
			"area": "shared",
			"rates": [
				{
					"name": "services/swift/account/container/object:create",
					"limit": 1000,
					"window": "1s",
					"default_limit": 500,
					"default_window": "1s"
				},
				{
					"name": "services/swift/account/container/object:delete",
					"limit": 1000,
					"window": "1s",
					"default_limit": 500,
					"default_window": "1s"
				},
				{
					"name": "services/swift/account:create",
					"limit": 10,
					"window": "1m",
					"default_limit": 5,
					"default_window": "1m"
				}
			],
			"scraped_at": 22
		}
	]
`

var projectMockServicesRateLimit = &ProjectServiceReports{
	"shared": &ProjectServiceReport{
		ServiceInfo: limes.ServiceInfo{
			Type: "shared",
			Area: "shared",
		},
		Rates: ProjectRateReports{
			"services/swift/account/container/object:create": {
				RateInfo: RateInfo{Name: "services/swift/account/container/object:create"},
				Limit:    1000,
				Window:   p2window(1 * WindowSeconds),
			},
			"services/swift/account/container/object:delete": {
				RateInfo: RateInfo{Name: "services/swift/account/container/object:delete"},
				Limit:    1000,
				Window:   p2window(1 * WindowSeconds),
			},
			"services/swift/account:create": {
				RateInfo: RateInfo{Name: "services/swift/account:create"},
				Limit:    10,
				Window:   p2window(1 * WindowMinutes),
			},
		},
		ScrapedAt: p2time(22),
	},
}

var projectMockServicesRateLimitDeviatingFromDefaults = &ProjectServiceReports{
	"shared": &ProjectServiceReport{
		ServiceInfo: limes.ServiceInfo{
			Type: "shared",
			Area: "shared",
		},
		Rates: ProjectRateReports{
			"services/swift/account/container/object:create": {
				RateInfo:      RateInfo{Name: "services/swift/account/container/object:create"},
				Limit:         1000,
				Window:        p2window(1 * WindowSeconds),
				DefaultLimit:  500,
				DefaultWindow: p2window(1 * WindowSeconds),
			},
			"services/swift/account/container/object:delete": {
				RateInfo:      RateInfo{Name: "services/swift/account/container/object:delete"},
				Limit:         1000,
				Window:        p2window(1 * WindowSeconds),
				DefaultLimit:  500,
				DefaultWindow: p2window(1 * WindowSeconds),
			},
			"services/swift/account:create": {
				RateInfo:      RateInfo{Name: "services/swift/account:create"},
				Limit:         10,
				Window:        p2window(1 * WindowMinutes),
				DefaultLimit:  5,
				DefaultWindow: p2window(1 * WindowMinutes),
			},
		},
		ScrapedAt: p2time(22),
	},
}

func TestProjectServicesRateLimitMarshall(t *testing.T) {
	th.CheckJSONEquals(t, projectServicesRateLimitMockJSON, projectMockServicesRateLimit)
}

func TestProjectServicesRateLimitUnmarshall(t *testing.T) {
	actual := &ProjectServiceReports{}
	err := actual.UnmarshalJSON([]byte(projectServicesRateLimitMockJSON))
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, projectMockServicesRateLimit, actual)
}

func TestProjectServicesRateLimitDeviatingFromDefaultsMarshall(t *testing.T) {
	th.CheckJSONEquals(t, projectServicesRateLimitDeviatingFromDefaultsMockJSON, projectMockServicesRateLimitDeviatingFromDefaults)
}

func TestProjectServicesRateLimitDeviatingFromDefaultsUnmarshall(t *testing.T) {
	actual := &ProjectServiceReports{}
	err := actual.UnmarshalJSON([]byte(projectServicesRateLimitDeviatingFromDefaultsMockJSON))
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, projectMockServicesRateLimitDeviatingFromDefaults, actual)
}

func p2time(timestamp int64) *limes.UnixEncodedTime {
	t := limes.UnixEncodedTime{Time: time.Unix(timestamp, 0).UTC()}
	return &t
}

func p2window(val Window) *Window {
	return &val
}
