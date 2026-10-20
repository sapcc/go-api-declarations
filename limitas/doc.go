// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

// Package limitas contains the specification for the v2 API of [Limes] (due to be renamed to Limitas),
// an OpenStack service for managing quotas, commitments as well as usage and capacity data.
//
// # Ontology
//
// All data tracked by Limes follows the Keystone model of projects grouped into domains.
// Please refer to the [Keystone documentation] for details.
//
// Furthermore, data is structured into areas, then services, then categories, then resources and rates as follows:
//
//   - A service is any backend service that can track resources or rates.
//     Examples include Nova (the OpenStack compute service) or Cinder (the OpenStack block storage service).
//   - Resources describe objects that are provisioned at some point and then kept around until they are later deleted.
//     Examples of resources include VMs in a compute service, volumes in a storage service, or floating IPs in a network service.
//     (This does not mean that each individual floating IP is a resource. The entire concept of "floating IPs" is the resource.)
//     Resource usage and capacity is always measured at a specific point in time, like for the Prometheus metric type "gauge".
//   - Rates are measurements that only ever increase over time, similar to the Prometheus metric type "counter".
//     For example, if a compute service has the resource "VMs", it might have rates like "VM creations" or "VM deletions".
//     Rates describe countable events like in this example, or measurable transfers like "bytes received" or "bytes transferred" on network links.
//   - Each service is grouped into an area, and each resource/rate within a service is grouped into a category.
//     These groupings are purely informational, and can be used to filter reports or to organize UIs displaying data from the Limes API.
//
// TODO: detailed explanations from LIQUID
// TODO: API structure from staging
// TODO: refer to subpackages for endpoint docs
//
// [Limes]: https://github.com/sapcc/limes
// [Keystone documentation]: https://docs.openstack.org/keystone/latest/admin/identity-concepts.html
package limitas

// Area identifies a set of services.
// Each service must belong to exactly one area.
// This classification can be used to filter reports (see [GenericReportOpts]), or to organize UIs displaying data for several services.
//
// This type is used to distinguish service types from other types of string values in function signatures.
type Area string

// ServiceType identifies a backend service that can have resources or rates.
//
// This type is used to distinguish service types from other types of string values in function signatures.
type ServiceType string
