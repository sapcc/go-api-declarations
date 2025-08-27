// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package liquid

import (
	"encoding/json"
	"math/big"
	"testing"

	. "github.com/majewsky/gg/option"

	th "github.com/sapcc/go-api-declarations/internal/testhelper"
)

func TestCloneServiceUsageRequest(t *testing.T) {
	// this dummy request sets all possible fields in order to test cloning of all levels
	request := ServiceUsageRequest{
		AllAZs: []AvailabilityZone{"az-one"},
		ProjectMetadata: Some(ProjectMetadata{
			UUID: "uuid-for-dresden",
			Name: "dresden",
			Domain: DomainMetadata{
				UUID: "uuid-for-germany",
				Name: "germany",
			},
		}),
		SerializedState: json.RawMessage(`{"state":"d41d8cd98f00b204e9800998ecf8427e"}`),
	}

	clonedRequest := request.Clone()
	th.CheckDeepEquals(t, request, clonedRequest)
	th.CheckFullySeparate(t, request, clonedRequest)
}

func TestCloneServiceUsageReport(t *testing.T) {
	// this dummy report sets all possible fields in order to test cloning of all levels
	report := ServiceUsageReport{
		InfoVersion: 42,
		Resources: map[ResourceName]*ResourceUsageReport{
			"capacity": {
				Forbidden: true,
				Quota:     Some[int64](10),
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"az-one": {
						Usage:         5,
						PhysicalUsage: Some[uint64](2),
						Quota:         Some[int64](10),
						Subresources: []Subresource{{
							ID:         "id-1",
							Name:       "first",
							Usage:      Some[uint64](5),
							Attributes: json.RawMessage(`{"foo":"bar"}`),
						}},
					},
				},
			},
		},
		Rates: map[RateName]*RateUsageReport{
			"thing_creations": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					AvailabilityZoneAny: {
						Usage: Some(big.NewInt(23)),
					},
				},
			},
		},
		Metrics: map[MetricName][]Metric{
			"bar_count": {{
				Value:       23,
				LabelValues: []string{"qux"},
			}},
		},
		SerializedState: json.RawMessage(`{"state":"d41d8cd98f00b204e9800998ecf8427e"}`),
	}

	clonedReport := report.Clone()
	th.CheckDeepEquals(t, report, clonedReport)
	th.CheckFullySeparate(t, report, clonedReport)
}
