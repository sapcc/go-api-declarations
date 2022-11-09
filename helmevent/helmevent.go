/*******************************************************************************
*
* Copyright 2022 SAP SE
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

// Package helmevent contains data structures describing the event messages that
// our CI generates for Helm deployments (i.e. "helm install" and "helm upgrade".
package helmevent

import (
	"time"
)

// Event describes a deployment (i.e. install or upgrade) of one or more Helm releases.
type Event struct {
	Region string `json:"region"`
	//NOTE: This should be "recorded-at". The inconsistent naming needs to stay like this now for backwards compatibility.
	RecordedAt   *time.Time         `json:"recorded_at"`
	GitRepos     map[string]GitRepo `json:"git"`
	HelmReleases []*HelmRelease     `json:"helm-release"`
	Pipeline     Pipeline           `json:"pipeline"`
}

// GitRepo appears in type Event. It describes the state of a Git repository
// that was checked out for a specific deployment.
type GitRepo struct {
	AuthoredAt  *time.Time `json:"authored-at"`
	Branch      string     `json:"branch"`
	CommittedAt *time.Time `json:"committed-at"`
	CommitID    string     `json:"commit-id"`
	RemoteURL   string     `json:"remote-url"`
}

// HelmRelease appears in type Event. It describes a Helm release that was
// installed or upgraded as part of a specific deployment.
type HelmRelease struct {
	Name    string  `json:"name"`
	Outcome Outcome `json:"outcome"`

	//ChartID contains "${name}-${version}" for charts pulled from Chartmuseum.
	//ChartPath contains the path to that chart inside helm-charts.git for charts
	//coming from helm-charts.git directly. Exactly one of those must be set.
	ChartID   string `json:"chart-id"`
	ChartPath string `json:"chart-path"`
	Cluster   string `json:"cluster"`
	//ImageVersion is only set for releases that take an image version produced by an earlier pipeline job.
	ImageVersion string `json:"image-version,omitempty"`
	Namespace    string `json:"kubernetes-namespace"`
	//DeployedImages is a list of all Docker image references that were found in the deployed Helm manifest.
	DeployedImages []string `json:"deployed-images"`

	//StartedAt is not set for OutcomeNotDeployed.
	StartedAt *time.Time `json:"started-at"`
	//FinishedAt is not set for OutcomeNotDeployed and OutcomeHelmUpgradeFailed.
	FinishedAt      *time.Time `json:"finished-at,omitempty"`
	DurationSeconds *uint64    `json:"duration,omitempty"`
}

// Outcome appears in type HelmRelease. It describes the final state of a Helm release.
type Outcome string

const (
	//OutcomeNotDeployed describes a Helm release that was not deployed because
	//of an unexpected error before `helm upgrade`.
	OutcomeNotDeployed Outcome = "not-deployed"
	//OutcomeSucceeded describes a Helm release that succeeded.
	OutcomeSucceeded Outcome = "succeeded"
	//OutcomeHelmUpgradeFailed describes a Helm release that failed during
	//`helm upgrade` or because some deployed pods did not come up correctly.
	OutcomeHelmUpgradeFailed Outcome = "helm-upgrade-failed"
	//OutcomeE2ETestFailed describes a Helm release that was deployed, but a
	//subsequent end-to-end test failed.
	OutcomeE2ETestFailed Outcome = "e2e-test-failed"
	//OutcomePartiallyDeployed is returned by Event.CombinedOutcome() when the event
	//in question contains some releases that are "succeeded" and some that are
	//"not-deployed". This value is not acceptable for an individual Helm release.
	OutcomePartiallyDeployed Outcome = "partially-deployed"
)

// IsKnownInputValue returns whether this value is acceptable for an individual
// Helm release.
func (o Outcome) IsKnownInputValue() bool {
	switch o {
	case OutcomeNotDeployed, OutcomeSucceeded, OutcomeHelmUpgradeFailed, OutcomeE2ETestFailed:
		return true
	case OutcomePartiallyDeployed:
		return false //not acceptable on an individual release, can only appear as result of Event.CombinedOutcome()
	default:
		return false
	}
}

// Pipeline appears in type Event. It describes the Concourse pipeline in which
// the given deployment was performed.
type Pipeline struct {
	BuildNumber  string `json:"build-number"`
	BuildURL     string `json:"build-url"`
	JobName      string `json:"job"`
	PipelineName string `json:"name"`
	TeamName     string `json:"team"`
	CreatedBy    string `json:"created-by"`
}

// CombinedOutcome merges the Outcome values of all HelmReleases in this Event
// into a single summary value.
func (event Event) CombinedOutcome() Outcome {
	hasSucceeded := false
	hasUndeployed := false
	for _, hr := range event.HelmReleases {
		switch hr.Outcome {
		case OutcomeHelmUpgradeFailed, OutcomeE2ETestFailed:
			//specific failure forces the entire result to be that failure
			return hr.Outcome
		case OutcomeSucceeded:
			hasSucceeded = true
		case OutcomeNotDeployed:
			hasUndeployed = true
		}
	}

	switch {
	case hasSucceeded && hasUndeployed:
		return OutcomePartiallyDeployed
	case hasSucceeded:
		return OutcomeSucceeded
	default:
		return OutcomeNotDeployed
	}
}

// CombinedStartDate merges the StartedAt values of all HelmReleases in this
// Event and returns the earliest start date.
func (event Event) CombinedStartDate() *time.Time {
	t := event.RecordedAt
	for _, hr := range event.HelmReleases {
		if hr.StartedAt == nil {
			continue
		}
		if t.After(*hr.StartedAt) {
			t = hr.StartedAt
		}
	}
	return t
}
