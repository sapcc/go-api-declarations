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
