// Package v1beta1 is manifest resources for the v1beta1 schema
package v1beta1

import (
	"github.com/Cray-HPE/loftsman/internal/logger"
)

const (
	// APIVersion is the string representation of the api version of this schema
	APIVersion = "manifests/v1beta1"
	// ChartSourceTypeDirectory is the identifier for spec.source.charts[].type where charts exist in a local directory
	ChartSourceTypeDirectory = "directory"
	// ChartSourceTypeRepo is the identifier for spec.source.charts[].type where charts exist in a chart repository
	ChartSourceTypeRepo = "repo"
)

// Manifest is the v1beta1 manifest object, implements internal/interfaces/manifest.go
type Manifest struct {
	logger        *logger.Logger
	tempDirectory string
	APIVersion    string    `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"`
	Metadata      *Metadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Spec          *Spec     `yaml:"spec,omitempty" json:"spec,omitempty"`
}

// Metadata stores the meta info about the manifest
type Metadata struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
}

// Spec is the root of definitions and instructions for the manifest
type Spec struct {
	Sources *Sources `yaml:"sources,omitempty" json:"sources,omitempty"`
	All     *Chart   `yaml:"all,omitempty" json:"all,omitempty"` // All is really a subset of *Chart, but we can restrict accepted parts via our schema
	Charts  []*Chart `yaml:"charts,omitempty" json:"charts,omitempty"`
}

// Sources contains info about artifact sources to use during loftsman shipping
type Sources struct {
	Charts []*ChartSource `yaml:"charts" json:"charts"`
}

// ChartSource is a local or remote location that can serve one or more packaged charts
type ChartSource struct {
	Type     string `yaml:"type,omitempty" json:"type,omitempty"`
	Name     string `yaml:"name,omitempty" json:"name,omitempty"`
	Location string `yaml:"location,omitempty" json:"location,omitempty"`
	// all properties below here are not relevant to every ChartSource.Type, but will either
	// just be used when needed/ignored otherwise for chart source types where they're irrelevant
	CredentialsSecret *ChartSourceCredentialsSecret `yaml:"credentialsSecret,omitempty" json:"credentialsSecret,omitempty"`
}

// ChartSourceCredentialsSecret is a reference to a Kubernetes secret storing credentials for accessing
// a chart source
type ChartSourceCredentialsSecret struct {
	Name        string `yaml:"name,omitempty" json:"name,omitempty"`
	Namespace   string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	UsernameKey string `yaml:"usernameKey,omitempty" json:"usernameKey,omitempty"`
	PasswordKey string `yaml:"passwordKey,omitempty" json:"passwordKey,omitempty"`
}

// Chart is a single chart to install/upgrade
type Chart struct {
	Name        string      `yaml:"name,omitempty" json:"name,omitempty"`
	Source      string      `yaml:"source,omitempty" json:"source,omitempty"`
	ReleaseName string      `yaml:"releaseName,omitempty" json:"releaseName,omitempty"`
	Namespace   string      `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Version     string      `yaml:"version,omitempty" json:"version,omitempty"`
	Values      interface{} `yaml:"values,omitempty" json:"-"` // json:"-" here is to ignore generic type validation, otherwise we'd get: json: unsupported type: map[interface {}]interface {}
	Timeout     string      `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}
