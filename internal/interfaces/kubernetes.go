package interfaces

import (
	"k8s.io/api/core/v1"
)

// Kubernetes is the interface for a k8s api object instance
type Kubernetes interface {
	Initialize(kubeconfigPath string, kubeContext string) error
	IsRetryError(err error) bool
	EnsureNamespace(name string) error
	FindConfigMap(name string, namespace string, withKey string, withValue string) (*v1.ConfigMap, error)
	InitializeConfigMap(name string, namespace string, data map[string]string) (*v1.ConfigMap, error)
	PatchConfigMap(name string, namespace string, data map[string]string) (*v1.ConfigMap, error)
}
