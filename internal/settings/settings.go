package settings

import (
	"errors"
	"fmt"
	"os"

	"github.com/Cray-HPE/go-lib/shell"
	"github.com/Cray-HPE/loftsman/internal/interfaces"
)

// Settings are all dynamic settings and data to be used in Loftsman operations
type Settings struct {
	RunID          string
	TempDirectory  string
	JSONLog        *JSONLog
	Namespace      string // the namespace where loftsman will keep internal-use resources
	Manifest       *Manifest
	ChartsSource   *interfaces.HelmChartsSource
	Kubernetes     *Kubernetes
	HelmExecConfig *interfaces.HelmExecConfig
}

// JSONLog are settings related to the written JSON log file
type JSONLog struct {
	Path string
	File *os.File
}

// Manifest are those specific to operations using, producing, validating manifests
type Manifest struct {
	Name       string // name of the manifest, used in avast operations
	Path       string // local path to the loftsman manifest to use for a ship
	Content    []byte // the bytes of the manifest file
	ChartNames string // comma-delimited list of charts provided when creating a new manifest file
}

// Kubernetes are settings and data related to Kubernetes API communication
type Kubernetes struct {
	KubeconfigPath string // absolute path to the k8s config path to use
	KubeContext    string // name of the kubeconfig context to use
}

// ValidateChartsSource will make sure related charts source settings are set up correctly
func (s *Settings) ValidateChartsSource() error {
	var err error
	if s.ChartsSource.Repo != "" && s.ChartsSource.Path != "" {
		return errors.New("both charts-repo and charts-path are set, you should use one or the other")
	}
	if s.ChartsSource.Path != "" {
		if _, err = os.Stat(s.ChartsSource.Path); os.IsNotExist(err) {
			return fmt.Errorf("charts-path %s not found", s.ChartsSource.Path)
		}
	}
	return nil
}

// ValidateManifestPath will ensure our manifest path setting is valid
func (s *Settings) ValidateManifestPath() error {
	var err error
	if _, err = os.Stat(s.Manifest.Path); os.IsNotExist(err) {
		return fmt.Errorf("manifest path %s not found", s.Manifest.Path)
	}
	return nil
}

// New gets a settings object with defaults
func New() *Settings {
	return &Settings{
		JSONLog: &JSONLog{
			Path: "",
		},
		Namespace:    "loftsman",
		ChartsSource: &interfaces.HelmChartsSource{},
		Manifest:     &Manifest{},
		Kubernetes:   &Kubernetes{},
		HelmExecConfig: &interfaces.HelmExecConfig{
			Binary: "helm",
			Shell:  &shell.Shell{},
		},
	}
}
