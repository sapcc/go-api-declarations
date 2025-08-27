// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package liquid

import (
	"testing"

	. "github.com/majewsky/gg/option"

	th "github.com/sapcc/go-api-declarations/internal/testhelper"
)

func TestCloneServiceQuotaRequest(t *testing.T) {
	// this dummy request sets all possible fields in order to test cloning of all levels
	request := ServiceQuotaRequest{
		Resources: map[ResourceName]ResourceQuotaRequest{
			"capacity": {
				PerAZ: map[AvailabilityZone]AZResourceQuotaRequest{
					"az-one": {
						Quota: 100,
					},
				},
			},
		},
		ProjectMetadata: Some(ProjectMetadata{
			UUID: "uuid-for-dresden",
			Name: "dresden",
			Domain: DomainMetadata{
				UUID: "uuid-for-germany",
				Name: "germany",
			},
		}),
	}

	clonedRequest := request.Clone()
	th.CheckDeepEquals(t, request, clonedRequest)
	th.CheckFullySeparate(t, request, clonedRequest)
}
