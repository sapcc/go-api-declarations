// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package liquid

import (
	"encoding/json"
	"testing"

	th "github.com/sapcc/go-api-declarations/internal/testhelper"

	. "github.com/majewsky/gg/option"
)

func TestCloneServiceInfo(t *testing.T) {
	// this dummy info sets all possible fields in order to test cloning of all levels
	info := ServiceInfo{
		Version:     42,
		DisplayName: "Test Service",
		Categories: map[CategoryName]CategoryInfo{
			"cat1": {DisplayName: "Category 1"},
			"cat2": {DisplayName: "Category 2"},
		},
		Resources: map[ResourceName]ResourceInfo{
			"capacity": {
				DisplayName:         "Capacity",
				Category:            Some(CategoryName("cat1")),
				Unit:                UnitBytes,
				Topology:            AZAwareTopology,
				HasCapacity:         true,
				NeedsResourceDemand: true,
				HasQuota:            true,
				HandlesCommitments:  true,
				Attributes:          json.RawMessage(`{"foo":"bar"}`),
			},
		},
		Rates: map[RateName]RateInfo{
			"thing_creations": {
				DisplayName: "Thing Creations",
				Category:    Some(CategoryName("cat2")),
				Unit:        UnitNone,
				Topology:    FlatTopology,
				HasUsage:    true,
			},
		},
		CapacityMetricFamilies: map[MetricName]MetricFamilyInfo{
			"foo_count": {
				Type:      MetricTypeCounter,
				Help:      "Counts foo things.",
				LabelKeys: []string{"flux_polarization_setting"},
			},
		},
		UsageMetricFamilies: map[MetricName]MetricFamilyInfo{
			"bar_count": {
				Type:      MetricTypeCounter,
				Help:      "Counts bar things.",
				LabelKeys: []string{"phase_shift"},
			},
		},
		UsageReportNeedsProjectMetadata:        true,
		QuotaUpdateNeedsProjectMetadata:        true,
		CommitmentHandlingNeedsProjectMetadata: true,
	}

	clonedInfo := info.Clone()
	th.CheckDeepEquals(t, info, clonedInfo)
	th.CheckFullySeparate(t, info, clonedInfo)
}
