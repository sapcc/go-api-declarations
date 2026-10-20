// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package limitas

// DomainMetadata contains the metadata for a domain.
// It appears in the DomainReport types of the subpackages.
type DomainMetadata struct {
	// The UUID field contains what is called "domain_id" in OpenStack.
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// ProjectMetadata contains the metadata for a project.
// It appears in the ProjectReport types of the subpackages.
type ProjectMetadata struct {
	// The UUID field contains what is called "project_id" in OpenStack.
	UUID string `json:"uuid"`
	Name string `json:"name"`
	// The ParentUUID is the ID of the domain in case the project is not a subproject,
	// else it is the UUID of another project. This hierarchy can have multiple levels.
	ParentUUID string         `json:"parent_uuid"`
	DomainInfo DomainMetadata `json:"domain"`
}
