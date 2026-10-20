// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package limitas

import (
	. "go.xyrillian.de/gg/option"

	"github.com/sapcc/go-api-declarations/liquid"
)

// GenericReportOpts contains query parameter options shared between resources and rates reports.
// It appears in the ResourceReportOpts and RateReportOpts types in the subpackages.
type GenericReportOpts struct {
	// If given, only report on services in this area.
	Area Option[Area] `q:"area"`
	// If given, only report on this service.
	ServiceType Option[ServiceType] `q:"service"`
	// If given, only report on resources in this category.
	Category Option[liquid.CategoryName] `q:"category"`
	// If true, include an InfoReport in the result.
	WithInfo bool `q:"with,value:info"`
}
