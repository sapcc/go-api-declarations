// SPDX-FileCopyrightText: 2018 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package limesresources

import (
	"testing"

	"go.xyrillian.de/gg/assert"

	th "github.com/sapcc/go-api-declarations/internal/testhelper"
	"github.com/sapcc/go-api-declarations/limes"
)

var domainServicesMockJSON = `
	[
		{
			"type": "compute",
			"area": "compute",
			"resources": [
				{
					"name": "cores",
					"quota": 50,
					"projects_quota": 20,
					"usage": 10
				},
				{
					"name": "ram",
					"unit": "MiB",
					"quota": 20480,
					"projects_quota": 10240,
					"usage": 4080
				}
			],
			"max_scraped_at": 1538955269,
			"min_scraped_at": 1538955116
		}
	]
`

var domainResourcesMockJSON = `
	[
		{
			"name": "cores",
			"quota": 50,
			"projects_quota": 20,
			"usage": 10
		},
		{
			"name": "ram",
			"unit": "MiB",
			"quota": 20480,
			"projects_quota": 10240,
			"usage": 4080
		}
	]
`

var domainMockResources = &DomainResourceReports{
	"cores": &DomainResourceReport{
		ResourceInfo: ResourceInfo{
			Name: "cores",
		},
		DomainQuota:   new(uint64(50)),
		ProjectsQuota: new(uint64(20)),
		Usage:         10,
	},
	"ram": &DomainResourceReport{
		ResourceInfo: ResourceInfo{
			Name: "ram",
			Unit: limes.UnitMebibytes,
		},
		DomainQuota:   new(uint64(20480)),
		ProjectsQuota: new(uint64(10240)),
		Usage:         4080,
	},
}

var domainMockServices = &DomainServiceReports{
	"compute": &DomainServiceReport{
		ServiceInfo: limes.ServiceInfo{
			Type: "compute",
			Area: "compute",
		},
		Resources:    *domainMockResources,
		MaxScrapedAt: p2time(1538955269),
		MinScrapedAt: p2time(1538955116),
	},
}

func TestDomainServicesMarshal(t *testing.T) {
	th.CheckJSONEquals(t, domainServicesMockJSON, domainMockServices)
}

func TestDomainServicesUnmarshal(t *testing.T) {
	actual := &DomainServiceReports{}
	err := actual.UnmarshalJSON([]byte(domainServicesMockJSON))
	assert.ErrEqual(t, err, nil)
	assert.Equal(t, actual, domainMockServices)
}

func TestDomainResourcesMarshal(t *testing.T) {
	th.CheckJSONEquals(t, domainResourcesMockJSON, domainMockResources)
}

func TestDomainResourcesUnmarshal(t *testing.T) {
	actual := &DomainResourceReports{}
	err := actual.UnmarshalJSON([]byte(domainResourcesMockJSON))
	assert.ErrEqual(t, err, nil)
	assert.Equal(t, actual, domainMockResources)
}
