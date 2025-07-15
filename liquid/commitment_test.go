// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package liquid

import (
	"testing"
	"time"

	. "github.com/majewsky/gg/option"

	th "github.com/sapcc/go-api-declarations/internal/testhelper"
)

const (
	dummyUUID1 = "30c343c8-7540-451a-bff5-fed9c35f8a43"
	dummyUUID2 = "c0191273-2126-4cb9-a3cd-8288300af14e"
	dummyUUID3 = "9ba49f82-3a78-4535-96a8-336415e233c5"
)

var (
	dummyNow = time.Date(2025, 7, 1, 12, 0, 0, 0, time.UTC)
)

func TestCommitmentChangeRequestRequiresConfirmation(t *testing.T) {
	// various shorthands for this test suite
	allAliveStatuses := []CommitmentStatus{
		CommitmentStatusPlanned,
		CommitmentStatusPending,
		CommitmentStatusGuaranteed,
		CommitmentStatusConfirmed,
		// "superseded" and "expired" are not here because those statuses do not allow further actions,
		// so they cannot appear as the OldStatus of a commitment in a changeset
	}
	ifGuaranteed := func(status CommitmentStatus, value uint64) uint64 {
		if status == CommitmentStatusGuaranteed {
			return value
		} else {
			return 0
		}
	}
	ifConfirmed := func(status CommitmentStatus, value uint64) uint64 {
		if status == CommitmentStatusConfirmed {
			return value
		} else {
			return 0
		}
	}
	makeRequest := func(byProject map[ProjectUUID]ProjectCommitmentChangeset) CommitmentChangeRequest {
		return CommitmentChangeRequest{
			AZ:          "az-one",
			InfoVersion: 1,
			ByProject:   byProject,
		}
	}

	// creation of an commitment with confirmation at a later date DOES NOT require confirmation...
	c := makeRequest(map[ProjectUUID]ProjectCommitmentChangeset{
		"proj-one": {
			ByResource: map[ResourceName]ResourceCommitmentChangeset{
				"capacity": {
					TotalConfirmedBefore:  50,
					TotalConfirmedAfter:   50,
					TotalGuaranteedBefore: 15,
					TotalGuaranteedAfter:  15,
					Commitments: []Commitment{{
						UUID:      dummyUUID1,
						NewStatus: Some(CommitmentStatusPlanned),
						Amount:    10,
						ConfirmBy: Some(dummyNow.Add(1 * time.Hour)),
						ExpiresAt: dummyNow.Add(24 * time.Hour),
					}},
				},
			},
		},
	})
	th.CheckDeepEquals(t, c.RequiresConfirmation(), false)

	// ...but creation of an immediately-confirmed commitment DOES require confirmation
	c = makeRequest(map[ProjectUUID]ProjectCommitmentChangeset{
		"proj-one": {
			ByResource: map[ResourceName]ResourceCommitmentChangeset{
				"capacity": {
					TotalConfirmedBefore:  50,
					TotalConfirmedAfter:   60,
					TotalGuaranteedBefore: 15,
					TotalGuaranteedAfter:  15,
					Commitments: []Commitment{{
						UUID:      dummyUUID1,
						NewStatus: Some(CommitmentStatusConfirmed),
						Amount:    10,
						ExpiresAt: dummyNow.Add(24 * time.Hour),
					}},
				},
			},
		},
	})
	th.CheckDeepEquals(t, c.RequiresConfirmation(), true)

	// moving a commitment from "planned" to "pending" on its ConfirmBy time DOES NOT require confirmation...
	c = makeRequest(map[ProjectUUID]ProjectCommitmentChangeset{
		"proj-one": {
			ByResource: map[ResourceName]ResourceCommitmentChangeset{
				"capacity": {
					TotalConfirmedBefore:  50,
					TotalConfirmedAfter:   50,
					TotalGuaranteedBefore: 15,
					TotalGuaranteedAfter:  15,
					Commitments: []Commitment{{
						UUID:      dummyUUID1,
						OldStatus: Some(CommitmentStatusPlanned),
						NewStatus: Some(CommitmentStatusPending),
						Amount:    10,
						ConfirmBy: Some(dummyNow),
						ExpiresAt: dummyNow.Add(24 * time.Hour),
					}},
				},
			},
		},
	})
	th.CheckDeepEquals(t, c.RequiresConfirmation(), false)

	// ...but moving from either "planned" or "pending" to "confirmed" DOES require confirmation
	for _, oldStatus := range []CommitmentStatus{CommitmentStatusPlanned, CommitmentStatusPending} {
		c = makeRequest(map[ProjectUUID]ProjectCommitmentChangeset{
			"proj-one": {
				ByResource: map[ResourceName]ResourceCommitmentChangeset{
					"capacity": {
						TotalConfirmedBefore: 50,
						TotalConfirmedAfter:  60,
						Commitments: []Commitment{{
							UUID:      dummyUUID1,
							OldStatus: Some(oldStatus),
							NewStatus: Some(CommitmentStatusConfirmed),
							Amount:    10,
							ConfirmBy: Some(dummyNow),
							ExpiresAt: dummyNow.Add(24 * time.Hour),
						}},
					},
				},
			},
		})
		t.Logf("checking confirmation from status %q", oldStatus)
		th.CheckDeepEquals(t, c.RequiresConfirmation(), true)
	}

	// creating a commitment in "guaranteed" DOES require confirmation...
	c = makeRequest(map[ProjectUUID]ProjectCommitmentChangeset{
		"proj-one": {
			ByResource: map[ResourceName]ResourceCommitmentChangeset{
				"capacity": {
					TotalConfirmedBefore:  50,
					TotalConfirmedAfter:   50,
					TotalGuaranteedBefore: 15,
					TotalGuaranteedAfter:  25,
					Commitments: []Commitment{{
						UUID:      dummyUUID1,
						NewStatus: Some(CommitmentStatusGuaranteed),
						Amount:    10,
						ConfirmBy: Some(dummyNow.Add(1 * time.Hour)),
						ExpiresAt: dummyNow.Add(24 * time.Hour),
					}},
				},
			},
		},
	})
	th.CheckDeepEquals(t, c.RequiresConfirmation(), true)

	// ...but then moving it into "confirmed" later DOES NOT require additional confirmation
	c = makeRequest(map[ProjectUUID]ProjectCommitmentChangeset{
		"proj-one": {
			ByResource: map[ResourceName]ResourceCommitmentChangeset{
				"capacity": {
					TotalConfirmedBefore:  50,
					TotalConfirmedAfter:   60,
					TotalGuaranteedBefore: 25,
					TotalGuaranteedAfter:  15,
					Commitments: []Commitment{{
						UUID:      dummyUUID1,
						OldStatus: Some(CommitmentStatusGuaranteed),
						NewStatus: Some(CommitmentStatusConfirmed),
						Amount:    10,
						ConfirmBy: Some(dummyNow),
						ExpiresAt: dummyNow.Add(24 * time.Hour),
					}},
				},
			},
		},
	})
	th.CheckDeepEquals(t, c.RequiresConfirmation(), false)

	// splitting a commitment DOES NOT require confirmation, regardless of status
	for _, status := range allAliveStatuses {
		c = makeRequest(map[ProjectUUID]ProjectCommitmentChangeset{
			"proj-one": {
				ByResource: map[ResourceName]ResourceCommitmentChangeset{
					"capacity": {
						TotalConfirmedBefore: 50,
						TotalConfirmedAfter:  50,
						Commitments: []Commitment{
							{
								UUID:      dummyUUID1,
								OldStatus: Some(status),
								NewStatus: Some(CommitmentStatusSuperseded),
								Amount:    10,
								ExpiresAt: dummyNow.Add(24 * time.Hour),
							},
							{
								UUID:      dummyUUID2,
								NewStatus: Some(status),
								Amount:    7,
								ExpiresAt: dummyNow.Add(24 * time.Hour),
							},
							{
								UUID:      dummyUUID3,
								NewStatus: Some(status),
								Amount:    3,
								ExpiresAt: dummyNow.Add(24 * time.Hour),
							},
						},
					},
				},
			},
		})
		t.Logf("checking split of status %q", status)
		th.CheckDeepEquals(t, c.RequiresConfirmation(), false)
	}

	// moving a commitment from one project to another requires confirmation only in status "guaranteed" or "confirmed"
	// (in those statuses, the underlying reservation may be tied to existing usage or specific properties of the source project,
	// so there could not be enough space for a reservation on the target project)
	for _, status := range allAliveStatuses {
		c = makeRequest(map[ProjectUUID]ProjectCommitmentChangeset{
			"proj-one": {
				ByResource: map[ResourceName]ResourceCommitmentChangeset{
					"capacity": {
						TotalConfirmedBefore:  50 + ifConfirmed(status, 10),
						TotalConfirmedAfter:   50,
						TotalGuaranteedBefore: 15 + ifGuaranteed(status, 10),
						TotalGuaranteedAfter:  15,
						Commitments: []Commitment{{
							UUID:      dummyUUID1,
							OldStatus: Some(status),
							Amount:    10,
							ExpiresAt: dummyNow.Add(24 * time.Hour),
						}},
					},
				},
			},
			"proj-two": {
				ByResource: map[ResourceName]ResourceCommitmentChangeset{
					"capacity": {
						TotalConfirmedBefore:  25,
						TotalConfirmedAfter:   25 + ifConfirmed(status, 10),
						TotalGuaranteedBefore: 5,
						TotalGuaranteedAfter:  5 + ifGuaranteed(status, 10),
						Commitments: []Commitment{{
							UUID:      dummyUUID1,
							NewStatus: Some(status),
							Amount:    10,
							ExpiresAt: dummyNow.Add(24 * time.Hour),
						}},
					},
				},
			},
		})
		t.Logf("checking move in status %q", status)
		th.CheckDeepEquals(t, c.RequiresConfirmation(), status == CommitmentStatusGuaranteed || status == CommitmentStatusConfirmed)
	}

	// converting a commitment between compatible resources requires confirmation only in status "guaranteed" or "confirmed"
	// (in those statuses, the underlying reservation may be tied to existing usage or specific properties of the old resource,
	// so there could not be enough space for a reservation on the new resource)
	for _, status := range allAliveStatuses {
		c = makeRequest(map[ProjectUUID]ProjectCommitmentChangeset{
			"proj-one": {
				ByResource: map[ResourceName]ResourceCommitmentChangeset{
					"capacity-dense": {
						TotalConfirmedBefore:  50 + ifConfirmed(status, 10),
						TotalConfirmedAfter:   50,
						TotalGuaranteedBefore: 15 + ifGuaranteed(status, 10),
						TotalGuaranteedAfter:  15,
						Commitments: []Commitment{{
							UUID:      dummyUUID1,
							OldStatus: Some(status),
							NewStatus: Some(CommitmentStatusSuperseded),
							Amount:    10,
							ExpiresAt: dummyNow.Add(24 * time.Hour),
						}},
					},
					"capacity-sparse": {
						TotalConfirmedBefore:  25,
						TotalConfirmedAfter:   25 + ifConfirmed(status, 20),
						TotalGuaranteedBefore: 5,
						TotalGuaranteedAfter:  5 + ifGuaranteed(status, 20),
						Commitments: []Commitment{{
							UUID:      dummyUUID2,
							NewStatus: Some(status),
							Amount:    20,
							ExpiresAt: dummyNow.Add(24 * time.Hour),
						}},
					},
				},
			},
		})
		t.Logf("checking conversion in status %q", status)
		th.CheckDeepEquals(t, c.RequiresConfirmation(), status == CommitmentStatusGuaranteed || status == CommitmentStatusConfirmed)
	}

	// transitioning into status "expired" is the only type of change to the relevant totals numbers
	// that does not require confirmation
	for _, status := range allAliveStatuses {
		c = makeRequest(map[ProjectUUID]ProjectCommitmentChangeset{
			"proj-one": {
				ByResource: map[ResourceName]ResourceCommitmentChangeset{
					"capacity-dense": {
						TotalConfirmedBefore:  50 + ifConfirmed(status, 10),
						TotalConfirmedAfter:   50,
						TotalGuaranteedBefore: 15 + ifGuaranteed(status, 10),
						TotalGuaranteedAfter:  15,
						Commitments: []Commitment{{
							UUID:      dummyUUID1,
							OldStatus: Some(status),
							NewStatus: Some(CommitmentStatusExpired),
							Amount:    10,
							ExpiresAt: dummyNow.Add(-1 * time.Minute),
						}},
					},
				},
			},
		})
		t.Logf("checking expiry in status %q", status)
		th.CheckDeepEquals(t, c.RequiresConfirmation(), false)
	}

	// cloud admins can delete commitments entirely, but this requires confirmation if in status "guaranteed" or "confirmed"
	for _, status := range allAliveStatuses {
		c = makeRequest(map[ProjectUUID]ProjectCommitmentChangeset{
			"proj-one": {
				ByResource: map[ResourceName]ResourceCommitmentChangeset{
					"capacity-dense": {
						TotalConfirmedBefore:  50 + ifConfirmed(status, 10),
						TotalConfirmedAfter:   50,
						TotalGuaranteedBefore: 15 + ifGuaranteed(status, 10),
						TotalGuaranteedAfter:  15,
						Commitments: []Commitment{{
							UUID:      dummyUUID1,
							OldStatus: Some(status),
							Amount:    10,
							ExpiresAt: dummyNow.Add(24 * time.Hour),
						}},
					},
				},
			},
		})
		t.Logf("checking deletion in status %q", status)
		th.CheckDeepEquals(t, c.RequiresConfirmation(), status == CommitmentStatusGuaranteed || status == CommitmentStatusConfirmed)
	}
}
