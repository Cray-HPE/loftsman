// Package manifest is for manifest-related resources
package manifest

import (
	"strings"
	"testing"
)

func TestValidateV1Beta1ValidWithMinimalChart(t *testing.T) {
	manifest := `---
apiVersion: manifests/v1beta1
metadata:
  name: test-manifest
spec:
  charts:
    - name: chart1
      namespace: default
      version: 0.0.1
`
	_, err := Validate(manifest)
	if err != nil {
		t.Errorf("Got unexpected error from manifest.TestValidateV1Beta1ValidWithMinimalChart(): %s", err)
	}
}

func TestValidateV1Beta1MissingChartVersion(t *testing.T) {
	manifest := `---
apiVersion: manifests/v1beta1
metadata:
  name: test-manifest
spec:
  charts:
    - name: chart1
      namespace: default
`
	_, err := Validate(manifest)
	if err == nil || !strings.Contains(err.Error(), "version is required") {
		t.Errorf("Didn't get expected error from manifest.TestValidateV1Beta1MissingChartVersion(), instead got: %s", err)
	}
}

func TestValidateV1Beta1InvalidYAML(t *testing.T) {
	manifest := `---
invalidyaml
	`
	_, err := Validate(manifest)
	if err == nil || err.Error() != "could not parse the manifest as yaml to retrieve the apiVersion" {
		t.Errorf("Didn't get expected error from manifest.TestValidateV1Beta1InvalidYAML(), instead got: %s", err)
	}
}

func TestValidateV1Beta1MissingMetadata(t *testing.T) {
	manifest := `---
apiVersion: manifests/v1beta1
spec:
  charts: []
`
	_, err := Validate(manifest)
	if err == nil || !strings.Contains(err.Error(), "manifest validation errors") {
		t.Errorf("Didn't get expected error from manifest.TestValidateV1Beta1MissingMetadata(), instead got: %s", err)
	}
}

func TestValidateV1Beta1MissingSpec(t *testing.T) {
	manifest := `---
apiVersion: manifests/v1beta1
metadata:
  name: test-manifest
`
	_, err := Validate(manifest)
	if err == nil || !strings.Contains(err.Error(), "manifest validation errors") {
		t.Errorf("Didn't get expected error from manifest.TestValidateV1Beta1MissingSpec(), instead got: %s", err)
	}
}

func TestValidateV1Beta1MissingCharts(t *testing.T) {
	manifest := `---
apiVersion: manifests/v1beta1
metadata:
  name: test-manifest
spec: {}
`
	_, err := Validate(manifest)
	if err == nil || !strings.Contains(err.Error(), "manifest validation errors") {
		t.Errorf("Didn't get expected error from manifest.TestValidateV1Beta1MissingCharts(), instead got: %s", err)
	}
}

func TestValidateV1Beta1MissingSourceOnChartWhileUsingChartsSource(t *testing.T) {
	manifest := `---
apiVersion: manifests/v1beta1
metadata:
  name: test-manifest
spec:
  sources:
    charts:
    - type: directory
      name: local
      location: ./charts
  charts:
  - name: chart1
    namespace: default
    version: 1.0.0
`
	_, err := Validate(manifest)
	if err == nil || !strings.Contains(err.Error(), "manifest validation errors") {
		t.Errorf("Didn't get expected error from manifest.TestValidateV1Beta1MissingSourceOnChartWhileUsingChartsSource(), instead got: %s", err)
	}
}

func TestValidateV1Beta1IncompleteChartSource(t *testing.T) {
	manifest := `---
apiVersion: manifests/v1beta1
metadata:
  name: test-manifest
spec:
  sources:
    charts:
    - type: directory
      name: local
  charts:
  - name: chart1
    source: local
    namespace: default
    version: 1.0.0
`
	_, err := Validate(manifest)
	if err == nil || !strings.Contains(err.Error(), "manifest validation errors") {
		t.Errorf("Didn't get expected error from manifest.TestValidateV1Beta1IncompleteChartSource(), instead got: %s", err)
	}
}

func TestValidateV1Beta1InvalidChartSourceType(t *testing.T) {
	manifest := `---
apiVersion: manifests/v1beta1
metadata:
  name: test-manifest
spec:
  sources:
    charts:
    - type: invalid
      name: local
      location: ./charts
  charts:
  - name: chart1
    source: local
    namespace: default
    version: 1.0.0
`
	_, err := Validate(manifest)
	if err == nil || !strings.Contains(err.Error(), "manifest validation errors") {
		t.Errorf("Didn't get expected error from manifest.TestValidateV1Beta1InvalidChartSourceType(), instead got: %s", err)
	}
}

func TestValidateV1Beta1ValidRepoChartSource(t *testing.T) {
	manifest := `---
apiVersion: manifests/v1beta1
metadata:
  name: test-manifest
spec:
  sources:
    charts:
    - type: repo
      name: remote
      location: https://repo
      credentialsSecret:
        name: secret
        namespace: default
        usernameKey: username
        passwordKey: password
  charts:
  - name: chart1
    source: remote
    namespace: default
    version: 1.0.0
`
	_, err := Validate(manifest)
	if err != nil {
		t.Errorf("Got unexpected error from manifest.TestValidateV1Beta1ValidRepoChartSource(): %s", err)
	}
}
