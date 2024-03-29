// Package kubernetes is for interactivity with the Kubernetes cluster API
package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8s "k8s.io/client-go/kubernetes"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"

	// Imported to support the fact that alternative client-go auth mechansims are not included
	// by default in the client-go default usage itself
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// Imported to support the fact that alternative client-go auth mechansims are not included
	// by default in the client-go default usage itself
	_ "k8s.io/client-go/plugin/pkg/client/auth/exec"
	// Imported to support the fact that alternative client-go auth mechansims are not included
	// by default in the client-go default usage itself
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// Imported to support the fact that alternative client-go auth mechansims are not included
	// by default in the client-go default usage itself
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// Imported to support the fact that alternative client-go auth mechansims are not included
	// by default in the client-go default usage itself
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
	// Imported to support the fact that alternative client-go auth mechansims are not included
	// by default in the client-go default usage itself
)

// Kubernetes is our k8s client object, implements internal/interfaces/kubernetes.go
type Kubernetes struct {
	client *k8s.Clientset
}

// Initialize will set up our object for connection to the desired cluster, and test that connection
func (k *Kubernetes) Initialize(kubeconfigPath string, kubeContext string) error {
	var err error
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}
	if kubeContext != "" {
		overrides.CurrentContext = kubeContext
	}
	if kubeconfigPath != "" {
		rules.ExplicitPath = kubeconfigPath
	}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
	if err != nil {
		return fmt.Errorf("could not get Kubernetes config for context %q: %s", kubeContext, err)
	}
	k.client, err = k8s.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("could not set up Kubernetes client: %s", err)
	}
	if _, err := k.client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{}); err != nil {
		return fmt.Errorf("error attempting to list namespaces in the cluster, are you sure you have your kubeconfig connected to an active cluster? %s", err)
	}
	return nil
}

// IsRetryError is used to determine whether or not we have an error that suggests we should retry
// TODO: I think eventually we'd pull down on the types here, and likely even make use of something like
//       the kerrors.SuggestsClientDelay, see
//       https://github.com/kubernetes/apimachinery/blob/master/pkg/api/errors/errors.go
//       This currently addresses a more unstable k8s cluster though
func (k *Kubernetes) IsRetryError(err error) bool {
	if kerrors.IsTooManyRequests(err) ||
		kerrors.IsServerTimeout(err) ||
		kerrors.IsTimeout(err) ||
		kerrors.IsServiceUnavailable(err) ||
		kerrors.IsConflict(err) ||
		kerrors.IsNotFound(err) {
		return true
	}
	return false
}

// EnsureNamespace will try to create a namespace, ignore the error if it already exists
func (k *Kubernetes) EnsureNamespace(name string) error {
	var err error
	err = retry.OnError(retry.DefaultBackoff, k.IsRetryError, func() error {
		_, err = k.client.CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}, metav1.CreateOptions{})
		if kerrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	})
	return err
}

func (k *Kubernetes) getCommonLabels() map[string]string {
	labels := make(map[string]string)
	labels["app.kubernetes.io/managed-by"] = "loftsman"
	return labels
}

// FindConfigMap will seek out an existing configmap in the namespace matching a name and with a particular
// key/value set in data
func (k *Kubernetes) FindConfigMap(name string, namespace string, withKey string, withValue string) (*v1.ConfigMap, error) {
	var err error
	var result *v1.ConfigMap
	err = retry.OnError(retry.DefaultBackoff, k.IsRetryError, func() error {
		list, err := k.client.CoreV1().ConfigMaps(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}
		for _, item := range list.Items {
			if item.ObjectMeta.Name == name {
				if value, ok := item.Data[withKey]; ok {
					if value == withValue {
						result = &item
						return nil
					}
				}
			}
		}
		result = nil
		return nil
	})

	return result, err
}

// InitializeShipConfigMap will ensure a configmap exists by name, in a namespace, with data. If an existing configmap
// is found and it is presisting previous data, then remove any previous data in the new version of the configmap
func (k *Kubernetes) InitializeShipConfigMap(name string, namespace string, data map[string]string) (*v1.ConfigMap, error) {
	var err error
	var result *v1.ConfigMap
	logConfigMapName := fmt.Sprintf("%s-ship-log", name)
	err = retry.OnError(retry.DefaultBackoff, k.IsRetryError, func() error {
		result, err = k.client.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if kerrors.IsNotFound(err) {
			annotations := make(map[string]string)
			annotations["loftsman.io/ship-log-configmap"] = logConfigMapName
			result, err = k.client.CoreV1().ConfigMaps(namespace).Create(context.Background(), &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:        name,
					Namespace:   namespace,
					Labels:      k.getCommonLabels(),
					Annotations: annotations,
				},
				Data: data,
			}, metav1.CreateOptions{})
			return err
		}
		if err != nil {
			return err
		}

		// Remove legacy annoations and fields
		patchData := map[string]interface{}{
			"metadata": map[string]interface{}{
				"annotations": map[string]interface{}{
					"loftsman.io/previous-data": nil,
					"loftsman.io/ship-log-configmap": logConfigMapName,
				},
			},
			"data": map[string]interface{}{
				"loftsman.log": nil,
			},
		}
		patchDataEncoded, err := json.Marshal(patchData)
		if err != nil {
			return err
		}

		result, err = k.client.CoreV1().ConfigMaps(namespace).Patch(context.Background(), name,
			types.MergePatchType, []byte(patchDataEncoded), metav1.PatchOptions{})

		return err
	})
	return result, err
}

// InitializeLogConfigMap will ensure a configmap exists by name, in a namespace, with data. If an existing configmap
// is found then it will not be modifed
func (k *Kubernetes) InitializeLogConfigMap(name string, namespace string, data map[string]string) (*v1.ConfigMap, error) {
	var err error
	var result *v1.ConfigMap
	err = retry.OnError(retry.DefaultBackoff, k.IsRetryError, func() error {
		result, err = k.client.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if kerrors.IsNotFound(err) {
			annotations := make(map[string]string)
			result, err = k.client.CoreV1().ConfigMaps(namespace).Create(context.Background(), &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:        name,
					Namespace:   namespace,
					Labels:      k.getCommonLabels(),
					Annotations: annotations,
				},
				Data: data,
			}, metav1.CreateOptions{})
			return err
		}
		if err != nil {
			return err
		}

		return err
	})
	return result, err
}

// PatchConfigMap will patch an existing configmap with the StrategicMergePatchType
func (k *Kubernetes) PatchConfigMap(name string, namespace string, data map[string]string) (*v1.ConfigMap, error) {
	var err error
	var result *v1.ConfigMap
	patchData, err := json.Marshal(v1.ConfigMap{
		Data: data,
	})
	if err != nil {
		return result, err
	}
	err = retry.OnError(retry.DefaultBackoff, k.IsRetryError, func() error {
		result, err = k.client.CoreV1().ConfigMaps(namespace).Patch(context.Background(), name,
			types.StrategicMergePatchType, []byte(patchData), metav1.PatchOptions{})
		return err
	})
	return result, err
}

// GetSecretKeyValue will retrieve a particular data key from a secret
func (k *Kubernetes) GetSecretKeyValue(secretName string, namespace string, dataKey string) (string, error) {
	var err error
	var secret *v1.Secret
	var result string
	err = retry.OnError(retry.DefaultBackoff, k.IsRetryError, func() error {
		secret, err = k.client.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		result = strings.TrimSpace(string(secret.Data[dataKey]))
		return nil
	})
	return result, err
}
