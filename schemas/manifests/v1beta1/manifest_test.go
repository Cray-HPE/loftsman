package v1beta1

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Cray-HPE/loftsman/internal/interfaces"
	"github.com/Cray-HPE/loftsman/internal/logger"
	custommocks "github.com/Cray-HPE/loftsman/mocks/custom-mocks"
)

func getTestManifest() *Manifest {
	m := &Manifest{
		APIVersion: "manifests/v1beta1",
		Metadata: &Metadata{
			Name: "test-manifest",
		},
		Spec: &Spec{
			Charts: []*Chart{},
		},
	}
	logFile, _ := os.Create(filepath.Join(os.TempDir(), "loftsman-tests-manifest-v1beta1.log"))
	m.SetLogger(logger.New(logFile, "loftsman-tests-manifest-v1beta1"))
	tempDir := filepath.Join(os.TempDir(), "loftsman-tests-manifest-v1beta1")
	os.Mkdir(tempDir, 0755)
	m.SetTempDirectory(tempDir)
	return m
}

func errsToString(errs []*interfaces.ManifestReleaseError) string {
	result := "\n"
	for _, err := range errs {
		result = fmt.Sprintf("%s\n%s", result, err)
	}
	return result
}

func TestLoad(t *testing.T) {
	manifest := &Manifest{}
	manifestContent := `---
apiVersion: manifests/v1beta1
metadata:
  name: test-manifest
spec:
  charts: []`
	err := manifest.Load(manifestContent)
	if err != nil {
		t.Errorf("Got unexpected error from manifest.v1beta1.TestLoad(): %s", err)
	}
}

func TestCreateNoCharts(t *testing.T) {
	manifest := &Manifest{}
	created, err := manifest.Create([]string{})
	if err != nil {
		t.Errorf("Got unexpected error from manifest.v1beta1.TestCreateNoCharts: %s", err)
	}
	if created == "" {
		t.Errorf("Got unexpected empty string from manifest.v1beta1.TestCreateNoCharts: %s", err)
	}
}

func TestCreateSomeCharts(t *testing.T) {
	manifest := &Manifest{}
	created, err := manifest.Create([]string{"one", "two", "three"})
	if err != nil {
		t.Errorf("Got unexpected error from manifest.v1beta1.TestCreateSomeCharts: %s", err)
	}
	if created == "" {
		t.Errorf("Got unexpected empty string from manifest.v1beta1.TestCreateSomeCharts: %s", err)
	}
	if !strings.Contains(created, "name: one") || !strings.Contains(created, "name: two") || !strings.Contains(created, "name: three") {
		t.Errorf("Didn't find expected chart names in created output from manifest.v1beta1.TestCreateSomeCharts, got output: %s", created)
	}
}

func TestReleaseNoCharts(t *testing.T) {
	manifest := getTestManifest()
	errs := manifest.Release(custommocks.GetKubernetesMock(false), custommocks.GetHelmMock([]*interfaces.HelmAvailableChartVersion{}))
	if len(errs) != 0 {
		t.Errorf("Got unexpected errors from manifest.v1beta1.TestReleaseNoCharts(): %s", errsToString(errs))
	}
}

func TestReleaseSingleMinimalChart(t *testing.T) {
	availableChartVersions := []*interfaces.HelmAvailableChartVersion{
		&interfaces.HelmAvailableChartVersion{
			Version: "0.0.1",
			Path:    "/tmp/test-chart-0.0.1.tgz",
		},
	}
	manifest := getTestManifest()
	manifest.Spec.Charts = []*Chart{
		&Chart{
			Name:      "test-chart",
			Namespace: "default",
			Version:   "0.0.1",
		},
	}
	errs := manifest.Release(custommocks.GetKubernetesMock(false), custommocks.GetHelmMock(availableChartVersions))
	if len(errs) != 0 {
		t.Errorf("Got unexpected errors from manifest.v1beta1.TestReleaseSingleMinimalChart(): %s", errsToString(errs))
	}
}

func TestReleaseOverFailedRelease(t *testing.T) {
	availableChartVersions := []*interfaces.HelmAvailableChartVersion{
		&interfaces.HelmAvailableChartVersion{
			Version: "0.0.1",
			Path:    "/tmp/fixed-failed-release-0.0.1.tgz",
		},
	}
	manifest := getTestManifest()
	manifest.Spec.Charts = []*Chart{
		&Chart{
			Name:      "fixed-failed-release",
			Namespace: "default",
			Version:   "0.0.1",
		},
	}
	errs := manifest.Release(custommocks.GetKubernetesMock(false), custommocks.GetHelmMock(availableChartVersions))
	if len(errs) != 0 {
		t.Errorf("Got unexpected errors from manifest.v1beta1.TestReleaseOverFailedRelease(): %s", errsToString(errs))
	}
}

func TestReleaseOverFailedCouldntClean(t *testing.T) {
	availableChartVersions := []*interfaces.HelmAvailableChartVersion{
		&interfaces.HelmAvailableChartVersion{
			Version: "0.0.1",
			Path:    "/tmp/failed-remove-0.0.1.tgz",
		},
	}
	manifest := getTestManifest()
	manifest.Spec.Charts = []*Chart{
		&Chart{
			Name:      "failed-remove",
			Namespace: "default",
			Version:   "0.0.1",
		},
	}
	errs := manifest.Release(custommocks.GetKubernetesMock(false), custommocks.GetHelmMock(availableChartVersions))
	if len(errs) == 0 {
		t.Error("Didn't get expected errors from manifest.v1beta1.TestReleaseOverFailedCouldntClean()")
	}
}

func TestReleaseChartWithValues(t *testing.T) {
	availableChartVersions := []*interfaces.HelmAvailableChartVersion{
		&interfaces.HelmAvailableChartVersion{
			Version: "0.0.1",
			Path:    "/tmp/chart-with-values-0.0.1.tgz",
		},
	}
	manifest := getTestManifest()
	manifest.Spec.Charts = []*Chart{
		&Chart{
			Name:      "chart-with-values",
			Namespace: "default",
			Version:   "0.0.1",
			Values: map[string]interface{}{
				"one": "1",
				"two": "2",
			},
		},
	}
	errs := manifest.Release(custommocks.GetKubernetesMock(false), custommocks.GetHelmMock(availableChartVersions))
	if len(errs) != 0 {
		t.Errorf("Got unexpected errors from manifest.v1beta1.TestReleaseChartWithValues(): %s", errsToString(errs))
	}
}

func TestReleaseChartWithFullChart(t *testing.T) {
	availableChartVersions := []*interfaces.HelmAvailableChartVersion{
		&interfaces.HelmAvailableChartVersion{
			Version: "0.0.1",
			Path:    "/tmp/full-chart-0.0.1.tgz",
		},
	}
	manifest := getTestManifest()
	manifest.Spec.Charts = []*Chart{
		&Chart{
			Name:      "full-chart",
			Namespace: "default",
			Version:   "0.0.1",
			Values: map[string]interface{}{
				"one": "1",
				"two": "2",
			},
		},
	}
	errs := manifest.Release(custommocks.GetKubernetesMock(false), custommocks.GetHelmMock(availableChartVersions))
	if len(errs) != 0 {
		t.Errorf("Got unexpected errors from manifest.v1beta1.TestReleaseChartWithFullChart(): %s", errsToString(errs))
	}
}

func TestReleaseWithFailedInstallUpgrade(t *testing.T) {
	availableChartVersions := []*interfaces.HelmAvailableChartVersion{
		&interfaces.HelmAvailableChartVersion{
			Version: "0.0.1",
			Path:    "/tmp/failed-0.0.1.tgz",
		},
	}
	manifest := getTestManifest()
	manifest.Spec.Charts = []*Chart{
		&Chart{
			Name:      "failed",
			Namespace: "default",
			Version:   "0.0.1",
			Values: map[string]interface{}{
				"one": "1",
				"two": "2",
			},
		},
	}
	errs := manifest.Release(custommocks.GetKubernetesMock(false), custommocks.GetHelmMock(availableChartVersions))
	if len(errs) == 0 {
		t.Errorf("Didn't get expected error from manifest.v1beta1.TestReleaseWithFailedInstallUpgrade()")
	}
}

func TestReleaseChartDoesntExist(t *testing.T) {
	availableChartVersions := []*interfaces.HelmAvailableChartVersion{}
	manifest := getTestManifest()
	manifest.Spec.Charts = []*Chart{
		&Chart{
			Name:      "chart-with-values",
			Namespace: "default",
			Values: map[string]interface{}{
				"one": "1",
				"two": "2",
			},
		},
	}
	errs := manifest.Release(custommocks.GetKubernetesMock(false), custommocks.GetHelmMock(availableChartVersions))
	if len(errs) == 0 {
		t.Error("Didn't get expected error from manifest.v1beta1.TestReleaseChartDoesntExist()")
	}
}

func TestGlobalChartTimeout(t *testing.T) {
	availableChartVersions := []*interfaces.HelmAvailableChartVersion{
		&interfaces.HelmAvailableChartVersion{
			Version: "0.0.1",
			Path:    "/tmp/full-chart-0.0.1.tgz",
		},
	}
	manifest := getTestManifest()
	manifest.Spec.ChartTimeout = "10m0s"
	manifest.Spec.Charts = []*Chart{
		&Chart{
			Name:      "full-chart",
			Namespace: "default",
			Version:   "0.0.1",
			Values: map[string]interface{}{
				"one": "1",
				"two": "2",
			},
		},
	}
	errs := manifest.Release(custommocks.GetKubernetesMock(false), custommocks.GetHelmMock(availableChartVersions))
	if len(errs) != 0 {
		t.Errorf("Got unexpected errors from manifest.v1beta1.TestGlobalChartTimeout(): %s", errsToString(errs))
	}
}

func TestIndividualChartTimeout(t *testing.T) {
	availableChartVersions := []*interfaces.HelmAvailableChartVersion{
		&interfaces.HelmAvailableChartVersion{
			Version: "0.0.1",
			Path:    "/tmp/full-chart-0.0.1.tgz",
		},
	}
	manifest := getTestManifest()
	manifest.Spec.Charts = []*Chart{
		&Chart{
			Name:      "full-chart",
			Namespace: "default",
			Version:   "0.0.1",
			Values: map[string]interface{}{
				"one": "1",
				"two": "2",
			},
			Timeout: "10m0s",
		},
	}
	errs := manifest.Release(custommocks.GetKubernetesMock(false), custommocks.GetHelmMock(availableChartVersions))
	if len(errs) != 0 {
		t.Errorf("Got unexpected errors from manifest.v1beta1.TestIndividualChartTimeout(): %s", errsToString(errs))
	}
}
