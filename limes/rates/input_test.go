// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package limesrates

import (
	"testing"

	th "github.com/sapcc/go-api-declarations/internal/testhelper"
)

var rateLimits = RateRequest{
	"object-store": ServiceRequest{
		"object/account/container:create": RateLimitRequest{Limit: 1000, Window: 1 * WindowSeconds},
		"object/account/container:delete": RateLimitRequest{Limit: 100, Window: 1 * WindowSeconds},
	},
}

var rateLimitJSON = `
	[
		{
			"type": "object-store",
			"rates": [
				{
					"name": "object/account/container:create",
					"limit": 1000,
					"window": "1s"
				},
				{
					"name": "object/account/container:delete",
					"limit": 100,
					"window": "1s"
				}
			]
		}
	]
`

func TestQuotaRateLimitMarshal(t *testing.T) {
	th.CheckJSONEquals(t, rateLimitJSON, rateLimits)
}

func TestRateLimitRequestUnmarshal(t *testing.T) {
	var actual RateRequest
	err := actual.UnmarshalJSON([]byte(rateLimitJSON))
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, rateLimits, actual)
}
