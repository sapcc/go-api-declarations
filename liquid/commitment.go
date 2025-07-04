// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package liquid

import (
	"time"

	. "github.com/majewsky/gg/option"
)

// CommitmentChangeRequest is the request payload format for POST /v1/change-commitments.
type CommitmentChangeRequest struct {
	AZ AvailabilityZone `json:"az"`

	// The same version number that was reported in the Version field of a GET /v1/info response.
	// The liquid shall reject this request if the version here differs from the value in the ServiceInfo currently held by the liquid.
	// This is used to ensure that Limes does not request commitment changes based on outdated resource metadata.
	InfoVersion uint64 `json:"infoVersion"`

	// On the first level, the commitment changeset is grouped by project.
	//
	// Changesets may span over multiple projects e.g. when moving commitments from one project to another.
	// In this case, the changeset will show the commitment as being deleted in the source project, and as being created in the target project.
	ByProject map[ProjectUUID]ProjectCommitmentChangeset `json:"byProject"`

	// Whether Limes allows the liquid to confirm (or deny) this changeset.
	// If true, Limes will only apply the changeset in its database if the response from the liquid is positive.
	// If false, Limes has already applied the changeset in its database, and is only notifying the liquid about what happened.
	//
	// Examples for ExpectsConfirmation = true include commitments moving into the "confirmed" status, or conversion of commitments between resources.
	// Examples for ExpectsConfirmation = false include commitments being split or moving into the "expired" status.
	ExpectsConfirmation bool `json:"expectsConfirmation"`
}

// CommitmentChangeResponse is the response payload format for POST /v1/change-commitments.
type CommitmentChangeResponse struct {
	// If ExpectsConfirmation was true, this field shall be empty if the changeset is confirmed, or contain a human-readable error message if the changeset was rejected.
	// If ExpectsConfirmation was false, Limes will ignore this field (or, at most, log it silently).
	RejectionReason string `json:"rejectionReason,omitempty"`
}

// ProjectCommitmentChangeset appears in type CommitmentChangeRequest.
// It contains all commitments that are part of a single atomic changeset that belong to a specific project in a specific AZ.
type ProjectCommitmentChangeset struct {
	// Metadata about the project from Keystone.
	// Only included if the ServiceInfo declared a need for it.
	ProjectMetadata Option[ProjectMetadata] `json:"projectMetadata,omitzero"`

	// On the second level, the commitment changeset is grouped by resource.
	//
	// Changesets may span over multiple resources when converting commitments for one resource into commitments for another resource.
	// In this case, the changeset will show the original commitment being deleted in one resource, and a new commitment being created in another.
	ByResource map[ResourceName]ResourceCommitmentChangeset `json:"byResource"`
}

// ResourceCommitmentChangeset appears in type CommitmentChangeRequest.
// It contains all commitments that are part of a single atomic changeset that belong to a given resource within a specific project and AZ.
type ResourceCommitmentChangeset struct {
	// The sum of all commitments for the given resource, project and AZ before and after applying the proposed commitment changeset.
	//
	// For example, if this changeset shows a commitment with Amount = 6 as being created, and one with Amount = 9 as being deleted,
	// and also there are several other commitments with a total Amount = 100 that the changeset does not touch,
	// then we will have TotalCommittedBefore = 109 and TotalCommittedAfter = 106.
	TotalCommittedBefore uint64 `json:"totalCommittedBefore"`
	TotalCommittedAfter  uint64 `json:"totalCommittedAfter"`

	// A commitment changeset may contain multiple commitments for a single resource within the same project.
	// For example, when a commitment is split into two parts, the changeset will show the original commitment being deleted and two new commitments being created.
	Commitments []Commitment `json:"commitments"`
}

// Commitment appears in type CommitmentChangeRequest.
//
// The commitment is located in a certain project and applies to a certain resource within a certain AZ.
// These metadata are implied by where the commitment is found within type CommitmentChangeRequest.
type Commitment struct {
	// The same UUID may appear multiple times within the same changeset for one specific circumstance:
	// If a commitment moves between projects, it will appear as being deleted in the source project and again as being created in the target project.
	UUID string `json:"uuid"`

	// These two status fields communicate one of three possibilities:
	//   - If OldStatus.IsNone() and NewStatus.IsSome(), the commitment is being created (or moved to this location).
	//   - If OldStatus.IsSome() and NewStatus.IsNone(), the commitment is being deleted (or moved away from this location).
	//   - If OldStatus.IsSome() and NewStatus.IsSome(), the commitment is only changing its status (e.g. from "active" to "expired" when ExpiresAt has passed).
	OldStatus Option[CommitmentStatus] `json:"oldStatus"`
	NewStatus Option[CommitmentStatus] `json:"newStatus"`

	Amount uint64 `json:"amount"`

	// For commitments in status "planned", this field contains the point in time in the future when the user wants for it to move into status "confirmed".
	// If confirmation is not possible by that point in time, the commitment will move into status "pending" until it can be confirmed.
	//
	// For all other status values, this field contains the point in time when the status transitioned from "active" to "pending",
	// or None() if the commitment was created for immediate confirmation and therefore started in status "pending".
	ConfirmBy Option[time.Time] `json:"confirmBy,omitzero"`

	// This field contains the point in time when the commitment moves into status "expired", unless it is deleted or moves into status "superseded" first.
	ExpiresAt time.Time `json:"expiresAt,omitzero"`
}

// CommitmentStatus is an enum containing the various lifecycle states of type Commitment.
type CommitmentStatus string

const (
	// CommitmentStatusPlanned means that the commitment has a ConfirmBy date in the future.
	// Planned commitments are used to notify the cloud about future resource demand.
	CommitmentStatusPlanned CommitmentStatus = "planned"
	// CommitmentStatusPending means that the commitment has a ConfirmBy date in the past, but the cloud has not confirmed it yet.
	// Pending commitments usually occur when there is not enough capacity to cover all current resource demands.
	CommitmentStatusPending CommitmentStatus = "pending"
	// CommitmentStatusConfirmed means that the commitment has been confirmed and is being honored by the cloud.
	// Confirmed commitments represent current resource demand that the cloud is able to guarantee.
	CommitmentStatusConfirmed CommitmentStatus = "confirmed"
	// CommitmentStatusSuperseded means that the commitment is no longer being honored by the cloud because it has been replaced by other commitments.
	// For example, when splitting a commitment into two halves, the new commitments will have the same status as the old commitment, and the old commitment will move into status "superseded".
	CommitmentStatusSuperseded CommitmentStatus = "superseded"
	// CommitmentStatusExpired means that the commitment is no longer being honored by the cloud because its lifetime has expired.
	// Expired commitments can be renewed by the user manually, but that involves creating a new commitment separately, such that ConfirmBy of the new commitment is equal to ExpiresAt of the old commitment.
	CommitmentStatusExpired CommitmentStatus = "expired"
)

// IsValid returns whether the given status is one of the predefined enum variants.
func (s CommitmentStatus) IsValid() bool {
	switch s {
	case CommitmentStatusPlanned, CommitmentStatusPending, CommitmentStatusConfirmed, CommitmentStatusSuperseded, CommitmentStatusExpired:
		return true
	default:
		return false
	}
}
