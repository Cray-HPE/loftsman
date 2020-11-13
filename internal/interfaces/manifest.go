package interfaces

import (
	"github.com/Cray-HPE/loftsman/internal/logger"
)

// ManifestReleaseError is a general-use object for recording manifest release errors
type ManifestReleaseError struct {
	Chart     string
	Version   string
	Namespace string
	Error     error
}

// Manifest is the interface for all manifest schema versions
type Manifest interface {
	GetName() string
	Create(initializeCharts []string) (string, error)
	Load(manifestContent string) error
	SetLogger(log *logger.Logger)
	SetTempDirectory(tempDirectory string)
	Release(kubernetes Kubernetes, helm Helm) []*ManifestReleaseError
}
