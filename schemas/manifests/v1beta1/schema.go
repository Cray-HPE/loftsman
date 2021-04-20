// Package v1beta1 is manifest resources for the v1 schema
package v1beta1

import (
	"github.com/Cray-HPE/loftsman/internal/logger"
)

const (
	// APIVersion is the string representation of the api version of this schema
	APIVersion = "manifests/v1beta1"
)

// Manifest is the v1 manifest object, implements internal/interfaces/manifest.go
type Manifest struct {
	logger        *logger.Logger
	tempDirectory string
	APIVersion    string    `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"`
	Metadata      *Metadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Spec          *Spec     `yaml:"spec,omitempty" json:"spec,omitempty"`
}

// Metadata is the v1 schema metadata object
type Metadata struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
}

// Spec is the v1 schema spec object
type Spec struct {
	ChartTimeout string   `yaml:"chartTimeout,omitempty" json:"chartTimeout,omitempty"`
	Charts       []*Chart `yaml:"charts,omitempty" json:"charts,omitempty"`
}

// Chart is a v1 schema Chart object
type Chart struct {
	Name      string      `yaml:"name,omitempty" json:"name,omitempty"`
	Namespace string      `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Version   string      `yaml:"version,omitempty" json:"version,omitempty"`
	Values    interface{} `yaml:"values,omitempty" json:"-"` // json:"-" here is to ignore generic type validation, otherwise we'd get: json: unsupported type: map[interface {}]interface {}
	Timeout   string      `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// Replace defines a single item in the list of a chart's "replaces" list
type Replace struct {
	Name      string `yaml:"name,omitempty" json:"name,omitempty"`
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
}
