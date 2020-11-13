package interfaces

import (
	"github.com/Cray-HPE/go-lib/shell"
)

// HelmExecConfig are settings/config related to running shell helm commands
type HelmExecConfig struct {
	Shell          shell.Interface
	Binary         string
	KubeconfigPath string
	KubeContext    string
}

// HelmChartsSource is an object storing config for where our Helm charts exist
type HelmChartsSource struct {
	Repo         string
	RepoUsername string
	RepoPassword string
	Path         string
}

// HelmAvailableChartVersion is a single version available for a chart
type HelmAvailableChartVersion struct {
	Path    string
	Version string
}

// HelmReleaseStatus represents a minimal representation of helm release status YAML output
type HelmReleaseStatus struct {
	Info     *HelmReleaseStatusInfo `yaml:"info"`
	Revision int                    `yaml:"version"`
}

// HelmReleaseStatusInfo represents a minimal representation of helm release status YAML info status
type HelmReleaseStatusInfo struct {
	Status string `yaml:"status"`
}

// Helm is an interface for a helm command object instance
type Helm interface {
	Initialize(execConfig *HelmExecConfig, chartsSource *HelmChartsSource) error
	Exec(subCommand string) (string, error)
	GetAvailableChartVersions(chartName string) ([]*HelmAvailableChartVersion, error)
	GetReleaseStatus(chartName string, chartNamespace string) (*HelmReleaseStatus, error)
}
