// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package castellum

import "encoding/json"

// Resource is the API representation of a resource.
type Resource struct {
	// fields that only appear in GET responses
	Checked    *Checked `json:"checked,omitempty"`
	AssetCount int64    `json:"asset_count"`

	// fields that are also allowed in PUT requests
	ConfigJSON        *json.RawMessage `json:"config,omitempty"`
	LowThreshold      *Threshold       `json:"low_threshold,omitempty"`
	HighThreshold     *Threshold       `json:"high_threshold,omitempty"`
	CriticalThreshold *Threshold       `json:"critical_threshold,omitempty"`
	SizeConstraints   *SizeConstraints `json:"size_constraints,omitempty"`
	SizeSteps         SizeSteps        `json:"size_steps"`
}

// Threshold appears in type Resource.
type Threshold struct {
	UsagePercent UsageValues `json:"usage_percent"`
	DelaySeconds uint32      `json:"delay_seconds,omitempty"`
}

// SizeSteps appears in type Resource.
type SizeSteps struct {
	Percent float64 `json:"percent,omitempty"`
	Single  bool    `json:"single,omitempty"`
}

// SizeConstraints appears in type Resource.
type SizeConstraints struct {
	Minimum               *uint64 `json:"minimum,omitempty"`
	Maximum               *uint64 `json:"maximum,omitempty"`
	MinimumFree           *uint64 `json:"minimum_free,omitempty"`
	MinimumFreeIsCritical bool    `json:"minimum_free_is_critical,omitempty"`
}
