// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package liquid

import (
	"math/big"
	"slices"
	"strings"
	"testing"

	. "github.com/majewsky/gg/option"

	"github.com/sapcc/go-api-declarations/internal/errorset"
)

var serviceInfo = ServiceInfo{
	Version: 73,
	Categories: map[CategoryName]CategoryInfo{
		"cat1": {
			DisplayName: "Category 1",
		},
		"cat2": {
			DisplayName: "Category 2",
		},
	},
	Resources: map[ResourceName]ResourceInfo{
		"foo": {
			Category:    Some(CategoryName("cat1")),
			Unit:        UnitNone,
			Topology:    AZAwareTopology,
			HasCapacity: true,
			HasQuota:    true,
		},
		"bar": {
			Category:    Some(CategoryName("cat1")),
			Unit:        UnitNone,
			Topology:    FlatTopology,
			HasCapacity: true,
			HasQuota:    true,
		},
		"baz": {
			Category:    Some(CategoryName("cat2")),
			Unit:        UnitNone,
			Topology:    FlatTopology,
			HasCapacity: false,
			HasQuota:    false,
		},
		"qux": {
			Category:    Some(CategoryName("cat2")),
			Unit:        UnitNone,
			Topology:    AZSeparatedTopology,
			HasCapacity: true,
			HasQuota:    true,
		},
		"quux": {
			Unit:        UnitNone,
			Topology:    AZSeparatedTopology,
			HasCapacity: true,
			HasQuota:    true,
		},
	},
	Rates: map[RateName]RateInfo{
		"corge": {
			Category: Some(CategoryName("cat1")),
			Unit:     UnitNone,
			HasUsage: true,
			Topology: AZAwareTopology,
		},
		"grault": {
			Category: Some(CategoryName("cat1")),
			Unit:     UnitNone,
			HasUsage: true,
			Topology: FlatTopology,
		},
		"garply": {
			Category: Some(CategoryName("cat2")),
			Unit:     UnitNone,
			HasUsage: true,
			Topology: AZAwareTopology,
		},
		"waldo": {
			Unit:     UnitNone,
			HasUsage: true,
			Topology: AZAwareTopology,
		},
	},
	CapacityMetricFamilies: map[MetricName]MetricFamilyInfo{
		"capacityMetric1": {
			Type:      MetricTypeGauge,
			LabelKeys: []string{"lk1", "lk2"},
		},
		"capacityMetric2": {
			Type:      MetricTypeGauge,
			LabelKeys: []string{"lk1", "lk2"},
		},
	},
	UsageMetricFamilies: map[MetricName]MetricFamilyInfo{
		"usageMetric1": {
			Type:      MetricTypeGauge,
			LabelKeys: []string{"lk1", "lk2"},
		},
		"usageMetric2": {
			Type:      MetricTypeGauge,
			LabelKeys: []string{"lk1", "lk2"},
		},
	},
}

func TestValidateServiceInfo(t *testing.T) {
	invalidServiceInfo := ServiceInfo{
		Categories: map[CategoryName]CategoryInfo{
			"default": {DisplayName: "Default"}, // Category name "default" is reserved
			"valid":   {DisplayName: "Valid"},
			"extra":   {DisplayName: "Extra"}, // This category is not used by any resource or rate which is forbidden
			"":        {DisplayName: "Empty"}, // Invalid category
			"empty":   {DisplayName: ""},      // Invalid category
		},
		Resources: map[ResourceName]ResourceInfo{
			"foo":         {Category: Some(CategoryName("empty"))}, // Topology is missing
			"bar":         {Category: Some(CategoryName("valid")), Topology: "InvalidTopology"},
			"baz":         {Category: Some(CategoryName("valid")), Topology: AZSeparatedTopology},
			"foo+private": {Category: Some(CategoryName("valid")), Topology: FlatTopology},               // Invalid name
			"qux1":        {Category: Some(CategoryName("")), Topology: FlatTopology},                    // Invalid category
			"qux2":        {Category: Some(CategoryName("someUnknownCategory")), Topology: FlatTopology}, // Unknown category
		},
		Rates: map[RateName]RateInfo{
			"corge":      {Category: Some(CategoryName("empty")), HasUsage: true}, // Topology is missing
			"grault":     {Category: Some(CategoryName("valid")), HasUsage: true, Topology: "InvalidTopology"},
			"garply":     {Category: Some(CategoryName("valid")), HasUsage: false, Topology: AZSeparatedTopology}, // HasUsage = false is not allowed
			"waldo":      {Category: Some(CategoryName("valid")), HasUsage: true, Topology: AZSeparatedTopology},
			"foo/create": {Category: Some(CategoryName("valid")), HasUsage: true, Topology: FlatTopology},               // Invalid name
			"bla1":       {Category: Some(CategoryName("")), HasUsage: true, Topology: FlatTopology},                    // Invalid category
			"bla2":       {Category: Some(CategoryName("someUnknownCategory")), HasUsage: true, Topology: FlatTopology}, // Unknown category
		},
	}
	expectedErrStrings := []string{
		`.Resources["foo"] has invalid topology ""`,
		`.Resources["bar"] has invalid topology "InvalidTopology"`,
		`.Resources["foo+private"] has invalid name (must match /^[a-zA-Z][a-zA-Z0-9._-]*$/)`,
		`.Rates["corge"] has invalid topology ""`,
		`.Rates["grault"] has invalid topology "InvalidTopology"`,
		`.Rates["garply"] declared with HasUsage = false, but must be true`,
		`.Rates["foo/create"] has invalid name (must match /^[a-zA-Z][a-zA-Z0-9._-]*$/)`,
		`.Resources["qux1"] has invalid category ""`,
		`.Resources["qux2"] has category "someUnknownCategory", which is not declared in .Categories`,
		`.Rates["bla1"] has invalid category ""`,
		`.Rates["bla2"] has category "someUnknownCategory", which is not declared in .Categories`,
		`.Categories["default"] has reserved identifier "default"`,
		`.Categories[""] has invalid identifier`,
		`.Categories["extra"] is not referenced by any resource or rate`,
		`.Categories["empty"] has invalid DisplayName`,
	}
	errs := validateServiceInfoImpl(invalidServiceInfo)
	assertErrorSet(t, errs, expectedErrStrings)

	errs = validateServiceInfoImpl(serviceInfo)
	if !errs.IsEmpty() {
		t.Errorf("expected no errors for a valid serviceInfo but got: %s", errs.Join(", "))
	}
}

func TestValidateCapacityReport(t *testing.T) {
	serviceCapacityRequest := ServiceCapacityRequest{
		AllAZs: []AvailabilityZone{"az-one", "az-two"},
	}
	wrongVersionReport := ServiceCapacityReport{
		InfoVersion: 409,
	}
	errs := validateCapacityReportImpl(wrongVersionReport, serviceCapacityRequest, serviceInfo)
	assertErrorSet(t, errs, []string{`received ServiceCapacityReport is invalid: expected .InfoVersion = 73, but got 409`})

	invalidServiceCapacityReport := ServiceCapacityReport{
		InfoVersion: 73,
		Resources: map[ResourceName]*ResourceCapacityReport{
			// foo is missing
			"bar": {
				PerAZ: map[AvailabilityZone]*AZResourceCapacityReport{
					"az-one": {Capacity: 42}, // AZ aware reporting on resource with flat topology
					"az-two": {Capacity: 42},
				},
			},
			"baz": {
				PerAZ: map[AvailabilityZone]*AZResourceCapacityReport{
					"any": {Capacity: 42}, // Report for resource with HasCapacity=false
				},
			},
			"qux": {
				PerAZ: map[AvailabilityZone]*AZResourceCapacityReport{
					"any": {Capacity: 42}, // Flat reporting for AZ aware resource
				},
			},
			"quux": {
				PerAZ: map[AvailabilityZone]*AZResourceCapacityReport{
					"az-one": {Capacity: 42}, // Partial AZ aware reporting, az-two is missing
				},
			},
			"unknown": {
				PerAZ: map[AvailabilityZone]*AZResourceCapacityReport{
					"any": {Capacity: 42}, // Report for resource which is not in ServiceInfo
				},
			},
		},
		Metrics: map[MetricName][]Metric{
			// capacityMetric1 is missing
			"capacityMetric2": {Metric{Value: 42}}, // Missing label values
			"unknownMetric":   {Metric{Value: 42}}, // MetricFamily that is not declared in ServiceInfo
		},
	}
	expectedErrStrings := []string{
		`missing value for .Resources["foo"] (resource was declared with HasCapacity = true)`,
		`.Resources["bar"].PerAZ has entries for []liquid.AvailabilityZone{"az-one", "az-two"}, which is invalid for topology "flat" (expected entries for []liquid.AvailabilityZone{"any"})`,
		`unexpected value for .Resources["baz"] (resource was declared with HasCapacity = false)`,
		`.Resources["qux"].PerAZ has entries for []liquid.AvailabilityZone{"any"}, which is invalid for topology "az-separated" (expected entries for []liquid.AvailabilityZone{"az-one", "az-two"})`,
		`.Resources["quux"].PerAZ has entries for []liquid.AvailabilityZone{"az-one"}, which is invalid for topology "az-separated" (expected entries for []liquid.AvailabilityZone{"az-one", "az-two"})`,
		`unexpected value for .Resources["unknown"] (resource was not declared)`,
		`missing value for .Metrics["capacityMetric1"] (declared in .CapacityMetricFamilies)`,
		`unexpected value for .Metrics["unknownMetric"] (not declared in .CapacityMetricFamilies)`,
		`malformed value for .Metrics["capacityMetric2"][0].LabelValues (expected 2, but got 0 entries)`,
	}
	errs = validateCapacityReportImpl(invalidServiceCapacityReport, serviceCapacityRequest, serviceInfo)
	assertErrorSet(t, errs, expectedErrStrings)

	serviceCapacityReport := ServiceCapacityReport{
		InfoVersion: 73,
		Resources: map[ResourceName]*ResourceCapacityReport{
			"foo": {
				PerAZ: map[AvailabilityZone]*AZResourceCapacityReport{
					"az-one": {Capacity: 42},
					"az-two": {Capacity: 42},
				},
			},
			"bar": {
				PerAZ: map[AvailabilityZone]*AZResourceCapacityReport{
					"any": {Capacity: 42},
				},
			},
			"qux": {
				PerAZ: map[AvailabilityZone]*AZResourceCapacityReport{
					"az-one": {Capacity: 42},
					"az-two": {Capacity: 42},
				},
			},
			"quux": {
				PerAZ: map[AvailabilityZone]*AZResourceCapacityReport{
					"az-one": {Capacity: 42},
					"az-two": {Capacity: 42},
				},
			},
		},
		Metrics: map[MetricName][]Metric{
			"capacityMetric1": {Metric{Value: 42, LabelValues: []string{"val1", "val2"}}},
			"capacityMetric2": {Metric{Value: 42, LabelValues: []string{"val1", "val2"}}},
		},
	}
	errs = validateCapacityReportImpl(serviceCapacityReport, serviceCapacityRequest, serviceInfo)
	if !errs.IsEmpty() {
		t.Errorf("expected no errors for a valid ServiceCapacityReport but got: %s", errs.Join(", "))
	}
}

func TestValidateUsageReport(t *testing.T) {
	serviceUsageRequest := ServiceUsageRequest{
		AllAZs: []AvailabilityZone{"az-one", "az-two"},
	}
	wrongVersionReport := ServiceUsageReport{
		InfoVersion: 409,
	}
	errs := validateUsageReportImpl(wrongVersionReport, serviceUsageRequest, serviceInfo)
	assertErrorSet(t, errs, []string{`received ServiceUsageReport is invalid: expected .InfoVersion = 73, but got 409`})

	invalidServiceUsageReport := ServiceUsageReport{
		InfoVersion: 73,
		Resources: map[ResourceName]*ResourceUsageReport{
			// foo is missing
			"bar": {
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"az-one": {Usage: 42, Quota: Some[int64](100)}, // AZ aware reporting on resource with flat topology
					"az-two": {Usage: 42, Quota: Some[int64](100)}, // Quota reporting on AZ level instead of resource level
				},
			},
			"baz": {
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"az-one": {Usage: 42, Quota: Some[int64](100)}, // Quota reporting on AZ level despite HasQuota = false
				},
			},
			"qux": {
				Quota: Some[int64](100), // Quota reporting on resource level instead of AZ level
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"any": {Usage: 42}, // Flat reporting for AZ aware resource
				},
			},
			"quux": {
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"az-one": {Quota: Some[int64](100), Usage: 42}, // Partial AZ aware reporting, az-two is missing
				},
			},
			"unknown": {
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"any": {Usage: 42}, // Report for resource which is not in ServiceInfo
				},
			},
		},
		Rates: map[RateName]*RateUsageReport{
			// corge is missing
			"grault": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"az-one": {Usage: Some(big.NewInt(5))}, // AZ aware reporting on rate with flat topology
					"az-two": {Usage: Some(big.NewInt(5))},
				},
			},
			"garply": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"any": {Usage: Some(big.NewInt(5))}, // Flat reporting for AZ aware rate
				},
			},
			"waldo": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"az-one": {Usage: Some(big.NewInt(5))}, // Partial AZ aware reporting, az-two is missing
				},
			},
			"unknown": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"any": {Usage: Some(big.NewInt(5))}, // Report for rate which is not in ServiceInfo
				},
			},
		},
		Metrics: map[MetricName][]Metric{
			// usageMetric1 is missing
			"usageMetric2":  {Metric{Value: 42}}, // Missing label values
			"unknownMetric": {Metric{Value: 42}}, // MetricFamily that is not declared in ServiceInfo
		},
	}
	expectedErrStrings := []string{
		`missing value for .Resources["foo"]`,
		`.Resources["bar"].PerAZ has entries for []liquid.AvailabilityZone{"az-one", "az-two"}, which is invalid for topology "flat" (expected entries for []liquid.AvailabilityZone{"any"})`,
		`.Resources["bar"] has no quota reported on resource level, which is invalid for HasQuota = true and topology "flat"`,
		`.Resources["baz"].PerAZ has entries for []liquid.AvailabilityZone{"az-one"}, which is invalid for topology "flat" (expected entries for []liquid.AvailabilityZone{"any"})`,
		`.Resources["baz"] has quota reported on AZ level, which is invalid for HasQuota = false`,
		`.Resources["qux"].PerAZ has entries for []liquid.AvailabilityZone{"any"}, which is invalid for topology "az-separated" (expected entries for []liquid.AvailabilityZone{"az-one", "az-two"})`,
		`.Resources["qux"] has quota reported on resource level, which is invalid for topology "az-separated"`,
		`.Resources["quux"].PerAZ has entries for []liquid.AvailabilityZone{"az-one"}, which is invalid for topology "az-separated" (expected entries for []liquid.AvailabilityZone{"az-one", "az-two"})`,
		`.Resources["quux"] with topology "az-separated" is missing quota reports on the following AZs: az-two`,
		`unexpected value for .Resources["unknown"] (resource was not declared)`,
		`missing value for .Rates["corge"]`,
		`.Rates["grault"].PerAZ has entries for []liquid.AvailabilityZone{"az-one", "az-two"}, which is invalid for topology "flat" (expected entries for []liquid.AvailabilityZone{"any"})`,
		`.Rates["garply"].PerAZ has entries for []liquid.AvailabilityZone{"any"}, which is invalid for topology "az-aware" (expected entries for []liquid.AvailabilityZone{"az-one", "az-two"})`,
		`.Rates["waldo"].PerAZ has entries for []liquid.AvailabilityZone{"az-one"}, which is invalid for topology "az-aware" (expected entries for []liquid.AvailabilityZone{"az-one", "az-two"})`,
		`unexpected value for .Rates["unknown"] (rate was not declared)`,
		`missing value for .Metrics["usageMetric1"] (declared in .UsageMetricFamilies)`,
		`unexpected value for .Metrics["unknownMetric"] (not declared in .UsageMetricFamilies)`,
		`malformed value for .Metrics["usageMetric2"][0].LabelValues (expected 2, but got 0 entries)`,
	}
	errs = validateUsageReportImpl(invalidServiceUsageReport, serviceUsageRequest, serviceInfo)
	assertErrorSet(t, errs, expectedErrStrings)

	// More test cases regarding quota reporting
	invalidServiceUsageReport2 := ServiceUsageReport{
		InfoVersion: 73,
		Resources: map[ResourceName]*ResourceUsageReport{
			"foo": {
				Quota: Some[int64](100),
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"az-one": {Usage: 42},
					"az-two": {Usage: 42, Quota: Some[int64](100)}, // Quota reporting on AZ level instead of resource level
				},
			},
			"bar": {
				Quota: Some[int64](100),
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"any": {Usage: 42},
				},
			},
			"baz": {
				Quota: Some[int64](100), // Quota reporting on resource level despite HasQuota = false
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"any": {Usage: 42, Quota: Some[int64](100)}, // Quota reporting on AZ level despite HasQuota = false
				},
			},
			"qux": {
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"unknown": {Usage: 42, Quota: Some[int64](100)}, // Quota reporting in AZ "unknown"
					"az-one":  {Usage: 42, Quota: Some[int64](100)},
					"az-two":  {Usage: 42, Quota: Some[int64](100)},
				},
			},
			"quux": {
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"az-one": {Usage: 42, Quota: Some[int64](100)},
					"az-two": {Usage: 42, Quota: Some[int64](100)},
				},
			},
		},
		Rates: map[RateName]*RateUsageReport{
			"corge": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"az-one": {Usage: Some(big.NewInt(5))},
					"az-two": {Usage: Some(big.NewInt(5))},
				},
			},
			"grault": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"any": {Usage: Some(big.NewInt(5))},
				},
			},
			"garply": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"az-one": {Usage: Some(big.NewInt(5))},
					"az-two": {Usage: Some(big.NewInt(5))},
				},
			},
			"waldo": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"az-one": {Usage: None[*big.Int]()},    // Usage missing for rate with HasUsage = true
					"az-two": {Usage: Some[*big.Int](nil)}, // Usage value not intact
				},
			},
		},
		Metrics: map[MetricName][]Metric{
			"usageMetric1": {Metric{Value: 42, LabelValues: []string{"val1", "val2"}}},
			"usageMetric2": {Metric{Value: 42, LabelValues: []string{"val1", "val2"}}},
		},
	}
	expectedErrStrings = []string{
		`.Resources["foo"] has quota reported on AZ level, which is invalid for topology "az-aware"`,
		`.Resources["baz"] has quota reported on resource level, which is invalid for HasQuota = false`,
		`.Resources["qux"] reports quota in AZ "unknown", which is invalid for topology "az-separated"`,
		`missing value for .Rates["waldo"].PerAZ["az-one"].Usage (rate was declared with HasUsage = true)`,
		`unexpected nil value in payload of .Rates["waldo"].PerAZ["az-two"].Usage`,
	}
	errs = validateUsageReportImpl(invalidServiceUsageReport2, serviceUsageRequest, serviceInfo)
	assertErrorSet(t, errs, expectedErrStrings)

	serviceUsageReport := ServiceUsageReport{
		InfoVersion: 73,
		Resources: map[ResourceName]*ResourceUsageReport{
			"foo": {
				Quota: Some[int64](100),
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"az-one": {Usage: 42},
					"az-two": {Usage: 42},
				},
			},
			"bar": {
				Quota: Some[int64](100),
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"any": {Usage: 42},
				},
			},
			"baz": {
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"any": {Usage: 42}, // Report for resource declared with HasQuota = false. This should be allowed since we also need reports for resources that only report usage
				},
			},
			"qux": {
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"az-one": {Usage: 42, Quota: Some[int64](100)},
					"az-two": {Usage: 42, Quota: Some[int64](100)},
				},
			},
			"quux": {
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"az-one": {Usage: 42, Quota: Some[int64](100)},
					"az-two": {Usage: 42, Quota: Some[int64](100)},
				},
			},
		},
		Rates: map[RateName]*RateUsageReport{
			"corge": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"az-one": {Usage: Some(big.NewInt(5))},
					"az-two": {Usage: Some(big.NewInt(5))},
				},
			},
			"grault": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"any": {Usage: Some(big.NewInt(5))},
				},
			},
			"garply": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"az-one": {Usage: Some(big.NewInt(5))},
					"az-two": {Usage: Some(big.NewInt(5))},
				},
			},
			"waldo": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"az-one": {Usage: Some(big.NewInt(5))},
					"az-two": {Usage: Some(big.NewInt(5))},
				},
			},
		},
		Metrics: map[MetricName][]Metric{
			"usageMetric1": {Metric{Value: 42, LabelValues: []string{"val1", "val2"}}},
			"usageMetric2": {Metric{Value: 42, LabelValues: []string{"val1", "val2"}}},
		},
	}
	errs = validateUsageReportImpl(serviceUsageReport, serviceUsageRequest, serviceInfo)
	if !errs.IsEmpty() {
		t.Errorf("expected no errors for a valid ServiceUsageReport but got: %s", errs.Join(", "))
	}
}

func assertErrorSet(t *testing.T, actualErrorSet errorset.ErrorSet, expectedErrStrings []string) {
	actualErrStrings := strings.Split(actualErrorSet.Join("\n"), "\n")
	for _, expectedErrStr := range expectedErrStrings {
		if !slices.Contains(actualErrStrings, expectedErrStr) {
			t.Errorf("expected error to be present: %s", expectedErrStr)
			t.Logf("actual errors = %s", actualErrorSet.Join(", "))
		}
	}
	if actualErrorSet.IsEmpty() {
		return // In this case actualErrStrings = [""]
	}
	for _, actualErrStr := range actualErrStrings {
		if !slices.Contains(expectedErrStrings, actualErrStr) {
			t.Errorf("unexpected error: %s", actualErrStr)
		}
	}
}
