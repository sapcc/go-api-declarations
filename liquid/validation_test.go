/*******************************************************************************
*
* Copyright 2025 SAP SE
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

package liquid

import (
	"math/big"
	"slices"
	"strings"
	"testing"

	"github.com/sapcc/go-api-declarations/internal/errorset"
)

var serviceInfo = ServiceInfo{
	Version: 73,
	Resources: map[ResourceName]ResourceInfo{
		"foo": {
			Unit:        UnitNone,
			Topology:    AZAwareTopology,
			HasCapacity: true,
			HasQuota:    true,
		},
		"bar": {
			Unit:        UnitNone,
			Topology:    FlatTopology,
			HasCapacity: true,
			HasQuota:    true,
		},
		"baz": {
			Unit:        UnitNone,
			Topology:    FlatTopology,
			HasCapacity: false,
			HasQuota:    false,
		},
		"qux": {
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
			Unit:     UnitNone,
			HasUsage: true,
			Topology: AZAwareTopology,
		},
		"grault": {
			Unit:     UnitNone,
			HasUsage: true,
			Topology: FlatTopology,
		},
		"garply": {
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
		Resources: map[ResourceName]ResourceInfo{
			"foo": {}, // Topology is missing
			"bar": {Topology: "InvalidTopology"},
			"baz": {Topology: AZSeparatedTopology},
		},
		Rates: map[RateName]RateInfo{
			"corge":  {HasUsage: true}, // Topology is missing
			"grault": {HasUsage: true, Topology: "InvalidTopology"},
			"garply": {HasUsage: false, Topology: AZSeparatedTopology}, // HasUsage = false is not allowed
			"waldo":  {HasUsage: true, Topology: AZSeparatedTopology},
		},
	}
	expectedErrStrings := []string{
		`.Resources["foo"] has invalid topology ""`,
		`.Resources["bar"] has invalid topology "InvalidTopology"`,
		`.Rates["corge"] has invalid topology ""`,
		`.Rates["grault"] has invalid topology "InvalidTopology"`,
		`.Rates["garply"] declared with HasUsage = false, but must be true`,
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
					"az-one": {Usage: 42, Quota: p2i64(100)}, // AZ aware reporting on resource with flat topology
					"az-two": {Usage: 42, Quota: p2i64(100)}, // Quota reporting on AZ level instead of resource level
				},
			},
			"baz": {
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"any": {Usage: 42}, // Report for resource with HasCapacity=false
				},
			},
			"qux": {
				Quota: p2i64(100), // Quota reporting on resource level instead of AZ level
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"any": {Usage: 42}, // Flat reporting for AZ aware resource
				},
			},
			"quux": {
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"az-one": {Quota: p2i64(100), Usage: 42}, // Partial AZ aware reporting, az-two is missing
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
					"az-one": {Usage: big.NewInt(5)}, // AZ aware reporting on rate with flat topology
					"az-two": {Usage: big.NewInt(5)},
				},
			},
			"garply": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"any": {Usage: big.NewInt(5)}, // Flat reporting for AZ aware rate
				},
			},
			"waldo": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"az-one": {Usage: big.NewInt(5)}, // Partial AZ aware reporting, az-two is missing
				},
			},
			"unknown": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"any": {Usage: big.NewInt(5)}, // Report for rate which is not in ServiceInfo
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
		`missing value for .Resources["foo"] (resource was declared with HasQuota = true)`,
		`.Resources["bar"].PerAZ has entries for []liquid.AvailabilityZone{"az-one", "az-two"}, which is invalid for topology "flat" (expected entries for []liquid.AvailabilityZone{"any"})`,
		`.Resources["bar"] has no quota reported on resource level, which is invalid for HasQuota = true and topology "flat"`,
		`unexpected value for .Resources["baz"] (resource was declared with HasQuota = false)`,
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

	serviceUsageReport := ServiceUsageReport{
		InfoVersion: 73,
		Resources: map[ResourceName]*ResourceUsageReport{
			"foo": {
				Quota: p2i64(100),
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"az-one": {Usage: 42},
					"az-two": {Usage: 42},
				},
			},
			"bar": {
				Quota: p2i64(100),
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"any": {Usage: 42},
				},
			},
			"qux": {
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"az-one": {Usage: 42, Quota: p2i64(100)},
					"az-two": {Usage: 42, Quota: p2i64(100)},
				},
			},
			"quux": {
				PerAZ: map[AvailabilityZone]*AZResourceUsageReport{
					"az-one": {Usage: 42, Quota: p2i64(100)},
					"az-two": {Usage: 42, Quota: p2i64(100)},
				},
			},
		},
		Rates: map[RateName]*RateUsageReport{
			"corge": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"az-one": {Usage: big.NewInt(5)},
					"az-two": {Usage: big.NewInt(5)},
				},
			},
			"grault": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"any": {Usage: big.NewInt(5)},
				},
			},
			"garply": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"az-one": {Usage: big.NewInt(5)},
					"az-two": {Usage: big.NewInt(5)},
				},
			},
			"waldo": {
				PerAZ: map[AvailabilityZone]*AZRateUsageReport{
					"az-one": {Usage: big.NewInt(5)},
					"az-two": {Usage: big.NewInt(5)},
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

// p2i64 makes a "pointer to int64".
func p2i64(val int64) *int64 {
	return &val
}
