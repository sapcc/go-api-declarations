// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package liquid

import (
	"encoding/json"
	"testing"

	. "github.com/majewsky/gg/option"

	th "github.com/sapcc/go-api-declarations/internal/testhelper"
)

func TestCloneServiceCapacityRequest(t *testing.T) {
	// this dummy request sets all possible fields in order to test cloning of all levels
	request := ServiceCapacityRequest{
		AllAZs: []AvailabilityZone{"az-one"},
		DemandByResource: map[ResourceName]ResourceDemand{
			"capacity": {
				OvercommitFactor: 1.1,
				PerAZ: map[AvailabilityZone]ResourceDemandInAZ{
					"az-one": {
						Usage:              100,
						UnusedCommitments:  20,
						PendingCommitments: 10,
					},
				},
			},
		},
	}

	clonedRequest := request.Clone()
	th.CheckDeepEquals(t, request, clonedRequest)
	th.CheckFullySeparate(t, request, clonedRequest)
}

func TestCloneServiceCapacityReport(t *testing.T) {
	// this dummy report sets all possible fields in order to test cloning of all levels
	report := ServiceCapacityReport{
		InfoVersion: 42,
		Resources: map[ResourceName]*ResourceCapacityReport{
			"capacity": {
				PerAZ: map[AvailabilityZone]*AZResourceCapacityReport{
					"az-one": {
						Capacity: 1000,
						Usage:    Some[uint64](500),
						Subcapacities: []Subcapacity{{
							ID:         "id-1",
							Name:       "first",
							Capacity:   1000,
							Usage:      Some[uint64](500),
							Attributes: json.RawMessage(`{"foo":"bar"}`),
						}},
					},
				},
			},
		},
		Metrics: map[MetricName][]Metric{
			"foo_count": {{
				Value:       23,
				LabelValues: []string{"qux"},
			}},
		},
	}

	clonedReport := report.Clone()
	th.CheckDeepEquals(t, report, clonedReport)
	th.CheckFullySeparate(t, report, clonedReport)
}
