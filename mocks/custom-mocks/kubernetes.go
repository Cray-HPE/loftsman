package mocks

import (
	kubernetesmocks "github.com/Cray-HPE/loftsman/mocks/interfaces"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// TestSecretKeyValue is just a mock value always returns when using GetSecretKeyValue
	TestSecretKeyValue = "secret"
)

// GetKubernetesMock will return a common mock for the Kubernetes interface/object
func GetKubernetesMock(triggerFoundConfigMap bool) *kubernetesmocks.Kubernetes {
	k := &kubernetesmocks.Kubernetes{}
	k.On("Initialize", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	k.On("EnsureNamespace", mock.AnythingOfType("string")).Return(nil)
	k.On("FindConfigMap", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(func(name string, namespace string, key string, value string) *v1.ConfigMap {
		if triggerFoundConfigMap {
			data := make(map[string]string)
			data["exists"] = "exists-value"
			return &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "configmap",
				},
				Data: data,
			}
		}
		return nil
	}, nil)
	k.On("InitializeShipConfigMap", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return(&v1.ConfigMap{}, nil)
	k.On("InitializeLogConfigMap", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return(&v1.ConfigMap{}, nil)
	k.On("PatchConfigMap", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return(&v1.ConfigMap{}, nil)
	k.On("GetSecretKeyValue", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(TestSecretKeyValue, nil)
	return k
}
