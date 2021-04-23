package kubernetes

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
)

var configMapList = `{
  "metadata": {},
  "items": [
    {
      "metadata": {
				"name": "found"
			},
      "data": {
        "status": "active"
      }
    }
  ]
}`

func TestInitialize(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", `=~http://loftsman-tests`, httpmock.NewStringResponder(200, `{}`))
	k := &Kubernetes{}
	err := k.Initialize("./.test-fixtures/kubeconfig.yaml", "default")
	if err != nil {
		t.Errorf("Got unexpected error from kubernetes.TestInitialize(): %s", err)
	}
}

func TestInitializeFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", `=~http://loftsman-tests`, httpmock.NewStringResponder(500, `{}`))
	k := &Kubernetes{}
	err := k.Initialize("./.test-fixtures/kubeconfig.yaml", "default")
	if err == nil || !strings.Contains(err.Error(), "error attempting to list namespaces") {
		t.Errorf("Didn't get expected error from kubernetes.TestInitializeFail(), instead got: %s", err)
	}
}

func TestIsRetryError(t *testing.T) {
	k := &Kubernetes{}
	result := k.IsRetryError(errors.New("error"))
	if result == true {
		t.Error("Got unexpected result true for kubernetes.TestIsRetryError")
	}
}

func TestEnsureNamespace(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", `=~http://loftsman-tests`, httpmock.NewStringResponder(200, `{}`))
	httpmock.RegisterResponder("POST", `=~http://loftsman-tests`, httpmock.NewStringResponder(200, `{}`))
	k := &Kubernetes{}
	_ = k.Initialize("./.test-fixtures/kubeconfig.yaml", "default")
	err := k.EnsureNamespace("ns")
	if err != nil {
		t.Errorf("Got unexpected error from kubernetes.TestEnsureNamespace(): %s", err)
	}
}

func TestFindConfigMapMatched(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", `=~http://loftsman-tests`, httpmock.NewStringResponder(200, configMapList))
	k := &Kubernetes{}
	_ = k.Initialize("./.test-fixtures/kubeconfig.yaml", "default")
	configMap, err := k.FindConfigMap("found", "default", "status", "active")
	if err != nil {
		t.Errorf("Got unexpected error from kubernetes.TestFindConfigMapMatched(): %s", err)
		return
	}
	configMapType := reflect.TypeOf(configMap)
	if configMapType.String() != "*v1.ConfigMap" {
		t.Errorf("Expected result in kubernetes.TestFindConfigMapMatched() to be a found configmap of type *v1.ConfigMap, but got: %s", configMapType)
	}
	if configMap.Data["status"] != "active" {
		t.Errorf("Expected to find status=active in configmap data in kubernetes.kubernetes.TestFindConfigMapMatched(), but didn', here's the configmap: %v", configMap)
	}
}

func TestFindConfigMapNotFound(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", `=~http://loftsman-tests`, httpmock.NewStringResponder(200, configMapList))
	k := &Kubernetes{}
	_ = k.Initialize("./.test-fixtures/kubeconfig.yaml", "default")
	configMap, err := k.FindConfigMap("not-found", "default", "status", "succes")
	if err != nil {
		t.Errorf("Got unexpected error from kubernetes.TestFindConfigMapNotFound(): %s", err)
		return
	}
	if configMap != nil {
		t.Errorf("Expected to get a nil result in kubernetes.TestFindConfigMapNotFound(), but got: %v", configMap)
	}
}

func TestInitializeConfigMapNew(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", `=~http://loftsman-tests`, httpmock.NewStringResponder(404, `{}`))
	httpmock.RegisterResponder("POST", `=~http://loftsman-tests`, httpmock.NewStringResponder(200, `{}`))
	k := &Kubernetes{}
	_ = k.Initialize("./.test-fixtures/kubeconfig.yaml", "default")
	data := make(map[string]string)
	data["one"] = "1"
	data["two"] = "2"
	_, err := k.InitializeConfigMap("loftsman-tests", "default", data)
	if err != nil {
		t.Errorf("Got unexpected error from kubernetes.TestInitializeConfigMapNew(): %s", err)
	}
}

func TestInitializeConfigMapExists(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", `=~http://loftsman-tests`, httpmock.NewStringResponder(200, `{"metadata": {"annotations": {}}}`))
	httpmock.RegisterResponder("POST", `=~http://loftsman-tests`, httpmock.NewStringResponder(200, `{}`))
	httpmock.RegisterResponder("PATCH", `=~http://loftsman-tests`, httpmock.NewStringResponder(200, `{}`))
	k := &Kubernetes{}
	_ = k.Initialize("./.test-fixtures/kubeconfig.yaml", "default")
	data := make(map[string]string)
	data["one"] = "1"
	data["two"] = "2"
	_, err := k.InitializeConfigMap("loftsman-tests", "default", data)
	if err != nil {
		t.Errorf("Got unexpected error from kubernetes.TestInitializeConfigMapExists(): %s", err)
	}
}

func TestPatchConfigMap(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", `=~http://loftsman-tests`, httpmock.NewStringResponder(200, `{}`))
	httpmock.RegisterResponder("PATCH", `=~http://loftsman-tests`, httpmock.NewStringResponder(200, `{}`))
	k := &Kubernetes{}
	_ = k.Initialize("./.test-fixtures/kubeconfig.yaml", "default")
	data := make(map[string]string)
	data["one"] = "1"
	data["two"] = "2"
	_, err := k.PatchConfigMap("loftsman-tests", "default", data)
	if err != nil {
		t.Errorf("Got unexpected error from kubernetes.TestPatchConfigMap(): %s", err)
	}
}

func TestGetSecretKeyValue(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", `=~http://loftsman-tests`, httpmock.NewStringResponder(200, `{"metadata": {"name": "secret-name", "namespace": "default"}, "data": {"test-key": "dGVzdC12YWx1ZQo="}}`))
	k := &Kubernetes{}
	_ = k.Initialize("./.test-fixtures/kubeconfig.yaml", "default")
	value, err := k.GetSecretKeyValue("secret-name", "default", "test-key")
	if err != nil {
		t.Errorf("Got unexpected error from kubernetes.TestGetSecretKeyValue(): %s", err)
	}
	if value != "test-value" {
		t.Errorf("Got unexpected value from kubernetes.TestGetSecretKeyValue(): %s", value)
	}
}
