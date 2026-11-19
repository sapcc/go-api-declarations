// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

// Package deployevent contains data structures for the event messages that our
// CI generates for Helm deployments (i.e. "helm install" and "helm upgrade")
// and Terraform runs (e.g. "terragrunt apply").
package deployevent

import (
	"encoding/json"
	"time"
)

// Event describes a deployment (i.e. install or upgrade) of one or more Helm releases.
type Event struct {
	// NOTE: "recorded_at" should be "recorded-at", and "helm-release" should be
	// "helm-releases". The inconsistent naming needs to stay like this now for
	// backwards compatibility.
	Region     string             `json:"region"`
	RecordedAt *time.Time         `json:"recorded_at"`
	GitRepos   map[string]GitRepo `json:"git,omitempty"`
	// Exactly one of the following fields must be filled.
	Pipeline         *Pipeline         `json:"pipeline,omitempty"`
	GreenhousePlugin *GreenhousePlugin `json:"greenhouse_plugin,omitempty"`
	MCP              *MCP              `json:"mcp,omitempty"`
	// Exactly one of the following fields must be filled.
	HelmReleases    []*HelmRelease             `json:"helm-release,omitempty"`
	TerraformRuns   []*TerraformRun            `json:"terraform-runs,omitempty"`
	ADDeployment    *ActiveDirectoryDeployment `json:"active-directory-deployment,omitempty"`
	VolumeMigration *VolumeMigration           `json:"volume_migration,omitempty"`
}

// GitRepo appears in type Event. It describes the state of a Git repository
// that was checked out for a specific deployment.
type GitRepo struct {
	AuthoredAt  *time.Time `json:"authored-at,omitempty"`
	Branch      string     `json:"branch"`
	CommittedAt *time.Time `json:"committed-at,omitempty"`
	CommitID    string     `json:"commit-id"`
	RemoteURL   string     `json:"remote-url"`
}

// TerraformRun appears in type Event. It describes a Terraform run that was
// executed and its outcome.
type TerraformRun struct {
	Outcome Outcome `json:"outcome"`

	// StartedAt is not set for OutcomeNotDeployed.
	StartedAt *time.Time `json:"started-at"`
	// FinishedAt is not set for OutcomeNotDeployed and OutcomeHelmUpgradeFailed.
	FinishedAt      *time.Time `json:"finished-at,omitempty"`
	DurationSeconds *uint64    `json:"duration,omitempty"`

	TerraformVersion string                  `json:"terraform-version"`
	ChangeSummary    *TerraformChangeSummary `json:"change-summary,omitempty"`
	ErrorMessage     string                  `json:"error-message,omitempty"`
}

// TerraformChangeSummary appears in TerraformRun. It describes how many
// resources were added, destroyed or changed by a Terraform run.
type TerraformChangeSummary struct {
	Added     int    `json:"added"`
	Changed   int    `json:"changed"`
	Removed   int    `json:"removed"`
	Operation string `json:"operation"`
}

// HelmRelease appears in type Event. It describes a Helm release that was
// installed or upgraded as part of a specific deployment.
type HelmRelease struct {
	Name    string  `json:"name"`
	Outcome Outcome `json:"outcome"`

	// ChartID contains "${name}-${version}" for charts pulled from Chartmuseum.
	// ChartPath contains the path to that chart inside helm-charts.git for charts
	// coming from helm-charts.git directly. Exactly one of those must be set.
	ChartID   string `json:"chart-id"`
	ChartPath string `json:"chart-path"`
	Cluster   string `json:"cluster"`
	// ImageVersion is only set for releases that take an image version produced by an earlier pipeline job.
	ImageVersion string `json:"image-version,omitempty"`
	Namespace    string `json:"kubernetes-namespace"`
	// DeployedImages is a list of all Docker image references that were found in the deployed Helm manifest.
	DeployedImages []string `json:"deployed-images"`

	// StartedAt is not set for OutcomeNotDeployed.
	StartedAt *time.Time `json:"started-at"`
	// FinishedAt is not set for OutcomeNotDeployed and OutcomeHelmUpgradeFailed.
	FinishedAt      *time.Time `json:"finished-at,omitempty"`
	DurationSeconds *uint64    `json:"duration,omitempty"`
}

// ActiveDirectoryDeployment appears in type Event. It describes a deployment of Active
// Directory to one of our Windows servers.
type ActiveDirectoryDeployment struct {
	Landscape string  `json:"landscape"` // e.g. "dev" or "prod"
	Hostname  string  `json:"host"`
	Outcome   Outcome `json:"outcome"`

	// StartedAt is not set for OutcomeNotDeployed.
	StartedAt *time.Time `json:"started-at"`
	// FinishedAt is not set for OutcomeNotDeployed and OutcomeADDeploymentFailed.
	FinishedAt      *time.Time `json:"finished-at,omitempty"`
	DurationSeconds *uint64    `json:"duration,omitempty"`
}

// VolumeMigration appears in type Event. It describes a Migration of virtual storage
// volume from one physical hardware to another.
type VolumeMigration struct {
	Outcome         Outcome             `json:"outcome"`
	Summary         string              `json:"summary"`
	Service         string              `json:"service"`
	Volume          string              `json:"volume"`
	SourceCluster   StorageCluster      `json:"source_cluster"`
	TargetCluster   StorageCluster      `json:"target_cluster"`
	StartedAt       time.Time           `json:"started_at"`
	FinishedAt      time.Time           `json:"finished_at"`
	Execution       ExecutionParams     `json:"execution"`
	RelatedIncident string              `json:"related_incident"`
	Requester       VolumeMigrationUser `json:"requester"`
	Assessment      json.RawMessage     `json:"assessment"`
}

// VolumeMigrationUser appears in type VolumeMigration.
type VolumeMigrationUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ExecutionParams appears in type VolumeMigration.
type ExecutionParams struct {
	Method         string `json:"method"`
	HumanInTheLoop bool   `json:"human_in_the_loop"`
}

// StorageCluster appears in type VolumeMigration.
type StorageCluster struct {
	Name                               string        `json:"name"`
	AvailabilityZone                   string        `json:"availability_zone"`
	Datacenter                         string        `json:"datacenter"`
	OverallUsedTiB                     float64       `json:"used_tib"`
	ClusterAverageAggregateUsedPercent float64       `json:"cluster_avg_aggr_used_pct"`
	Model                              string        `json:"model"`
	HardwareState                      string        `json:"hardware_state"`
	SVMCount                           int           `json:"svm_count"`
	Nodes                              []StorageNode `json:"nodes"`
}

// StorageNode appears in type StorageCluster.
type StorageNode struct {
	Name                           string  `json:"name"`
	Aggregate                      string  `json:"aggregate"`
	AggregateUsedPercent           float64 `json:"aggr_used_pct"`
	VolumeCount                    int     `json:"volume_count"`
	AverageCPUUsedPercent          float64 `json:"avg_cpu_pct"`
	MaxCPUUsedPercent              float64 `json:"max_cpu_pct"`
	AverageLatencyMicroseconds     int     `json:"avg_latency_us"`
	MaxLatencyMicroseconds         int     `json:"max_latency_us"`
	AverageThroughputMbitPerSecond float64 `json:"avg_throughput_mb_s"`
	MaxThroughputMbitPerSecond     float64 `json:"max_throughput_mb_s"`
}

// Outcome appears in type HelmRelease and TerraformRun. It describes the final
// state of a release.
type Outcome string

const (
	// OutcomeNotDeployed describes a Helm release that was not deployed because
	// of an unexpected error before `helm upgrade`.
	OutcomeNotDeployed Outcome = "not-deployed"
	// OutcomeSucceeded describes a Helm release that succeeded.
	OutcomeSucceeded Outcome = "succeeded"
	// OutcomeVolumeMigrationFailed describes a volume migration that failed
	OutcomeVolumeMigrationFailed Outcome = "volume-migration-failed"
	// OutcomeTerraformRunFailed describes a terraform run that failed
	OutcomeTerraformRunFailed Outcome = "terraform-run-failed"
	// OutcomeHelmUpgradeFailed describes a Helm release that failed during
	// `helm upgrade` or because some deployed pods did not come up correctly.
	OutcomeHelmUpgradeFailed Outcome = "helm-upgrade-failed"
	// OutcomeADDeploymentFailed describes an Active Directory deployment that
	// failed or did not run all the way through.
	OutcomeADDeploymentFailed Outcome = "active-directory-deployment-failed"
	// OutcomeE2ETestFailed describes a Helm release that was deployed, but a
	// subsequent end-to-end test failed.
	OutcomeE2ETestFailed Outcome = "e2e-test-failed"
	// OutcomePartiallyDeployed is returned by Event.CombinedOutcome() when the event
	// in question contains some releases that are "succeeded" and some that are
	// "not-deployed". This value is not acceptable for an individual Helm release.
	OutcomePartiallyDeployed Outcome = "partially-deployed"
)

// IsKnownInputValue returns whether this value is acceptable for an individual
// Helm release.
func (o Outcome) IsKnownInputValue() bool {
	switch o {
	case OutcomeNotDeployed, OutcomeSucceeded, OutcomeHelmUpgradeFailed, OutcomeE2ETestFailed, OutcomeTerraformRunFailed, OutcomeADDeploymentFailed, OutcomeVolumeMigrationFailed:
		return true
	case OutcomePartiallyDeployed:
		return false // not acceptable on an individual release, can only appear as result of Event.CombinedOutcome()
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

// GreenhousePlugin appears in type Event. It describes the Greenhouse plugin from which
// the given deployment was performed.
type GreenhousePlugin struct {
	OrganizationName string `json:"org"`
	Name             string `json:"name"`
	Version          string `json:"version"`
	Owner            string `json:"owned_by"`
}

// MCP appears in type Event. It describes the MCP in which the given action was performed.
type MCP struct {
	Channel   string `json:"channel"`
	SessionID string `json:"session_id"`
}

// CombinedOutcome merges the Outcome values of all Operations in this Event
// into a single summary value.
func (event Event) CombinedOutcome() Outcome {
	allOutcomes := make([]Outcome, 0, len(event.HelmReleases)+len(event.TerraformRuns))
	for _, hr := range event.HelmReleases {
		allOutcomes = append(allOutcomes, hr.Outcome)
	}
	for _, tr := range event.TerraformRuns {
		allOutcomes = append(allOutcomes, tr.Outcome)
	}
	if event.ADDeployment != nil {
		allOutcomes = append(allOutcomes, event.ADDeployment.Outcome)
	}
	if event.VolumeMigration != nil {
		allOutcomes = append(allOutcomes, event.VolumeMigration.Outcome)
	}

	hasSucceeded := false
	hasUndeployed := false
	for _, outcome := range allOutcomes {
		switch outcome {
		case OutcomeHelmUpgradeFailed, OutcomeE2ETestFailed, OutcomeTerraformRunFailed, OutcomeADDeploymentFailed, OutcomeVolumeMigrationFailed:
			// specific failure forces the entire result to be that failure
			return outcome
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
		if hr.StartedAt != nil && t.After(*hr.StartedAt) {
			t = hr.StartedAt
		}
	}
	for _, tr := range event.TerraformRuns {
		if tr.StartedAt != nil && t.After(*tr.StartedAt) {
			t = tr.StartedAt
		}
	}
	if event.ADDeployment != nil {
		ad := *event.ADDeployment
		if ad.StartedAt != nil && t.After(*ad.StartedAt) {
			t = ad.StartedAt
		}
	}
	return t
}
