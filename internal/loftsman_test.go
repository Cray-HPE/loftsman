package internal

import (
	"errors"
	"strings"
	"testing"

	custommocks "github.com/Cray-HPE/loftsman/mocks/custom-mocks"
)

func TestInitialize(t *testing.T) {
	loftsman := getTestLoftsman("")
	err := loftsman.Initialize("ship")
	if err != nil {
		t.Errorf("Got unexpected error from loftsman.TestInitialize(): %s", err)
	}
}

func TestInitializeManifestPath(t *testing.T) {
	loftsman := getTestLoftsman("")
	loftsman.manifest = nil
	loftsman.Settings.Manifest.Path = "./.test-fixtures/manifest-v1beta1.yaml"
	err := loftsman.Initialize("ship")
	if err != nil {
		t.Errorf("Got unexpected error from loftsman.TestInitializeManifestPath(): %s", err)
	}
}

func TestInitializeManifestPathDoesNotExist(t *testing.T) {
	loftsman := getTestLoftsman("")
	loftsman.manifest = nil
	loftsman.Settings.Manifest.Path = "/path/does/not/exist/manifest.yaml"
	err := loftsman.Initialize("ship")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("Didn't get expected error from loftsman.TestInitializeManifestPathDoesNotExist(), instead got: %s", err)
	}
}

func TestInitializeManifestPathInvalidManifest(t *testing.T) {
	loftsman := getTestLoftsman("")
	loftsman.manifest = nil
	loftsman.Settings.Manifest.Path = "./.test-fixtures/invalid-manifest-yaml.yaml"
	err := loftsman.Initialize("ship")
	if err == nil || !strings.Contains(err.Error(), "could not parse") {
		t.Errorf("Didn't get expected error from loftsman.TestInitializeManifestPathInvalidManifest(), instead got: %s", err)
	}
}

func Test_fail(t *testing.T) {
	loftsman := getTestLoftsman("tests")
	loftsman.fail(errors.New("fail"))
}

func TestShipChartsPath(t *testing.T) {
	loftsman := getTestLoftsman("ship")
	loftsman.Settings.ChartsSource.Path = "./helm/.test-fixtures/charts"
	err := loftsman.Ship()
	if err != nil {
		t.Errorf("Got unexpected error from loftsman.TestShipChartsPath(): %s", err)
	}
}

func TestShipChartsRepo(t *testing.T) {
	loftsman := getTestLoftsman("ship")
	loftsman.Settings.ChartsSource.Repo = "https://charts.io/charts"
	err := loftsman.Ship()
	if err != nil {
		t.Errorf("Got unexpected error from loftsman.TestShipChartsRepo(): %s", err)
	}
}

func TestShipChartsRepoWithCreds(t *testing.T) {
	loftsman := getTestLoftsman("ship")
	loftsman.Settings.ChartsSource.Repo = "https://charts.io/charts"
	loftsman.Settings.ChartsSource.RepoUsername = "user"
	loftsman.Settings.ChartsSource.RepoPassword = "pass"
	err := loftsman.Ship()
	if err != nil {
		t.Errorf("Got unexpected error from loftsman.TestShipChartsRepoWithCreds(): %s", err)
	}
}

func TestShipWhenAnotherRunning(t *testing.T) {
	loftsman := getTestLoftsman("ship")
	loftsman.kubernetes = custommocks.GetKubernetesMock(true)
	loftsman.Settings.ChartsSource.Path = "./helm/.test-fixtures/charts"
	err := loftsman.Ship()
	if err == nil || !strings.Contains(err.Error(), "another loftsman ship in progress") {
		t.Errorf("Didn't get expected error from loftsman.TestShipWhenAnotherRunning(), instead got: %s", err)
	}
}

func TestShipFailure(t *testing.T) {
	setReleaseErrors("ERROR")
	defer resetReleaseErrors()
	loftsman := getTestLoftsman("ship")
	loftsman.Settings.ChartsSource.Path = "./helm/.test-fixtures/charts"
	err := loftsman.Ship()
	if err == nil || !strings.Contains(err.Error(), "Some charts did not release successfully") {
		t.Errorf("Didn't get expected error from loftsman.TestShipFailure(), instead got: %s", err)
	}
}

func TestManifestCreate(t *testing.T) {
	loftsman := getTestLoftsman("manifest create")
	err := loftsman.ManifestCreate()
	if err != nil {
		t.Errorf("Got unexpected error from loftsman.TestManifestCreate(): %s", err)
	}
}

func TestManifestValidateV1Beta1(t *testing.T) {
	loftsman := getTestLoftsman("manifest validate")
	err := loftsman.ManifestValidate("./.test-fixtures/manifest-v1beta1.yaml")
	if err != nil {
		t.Errorf("Got unexpected error from loftsman.TestManifestValidate(): %s", err)
	}
}

func TestAvastMissingManifestSettings(t *testing.T) {
	loftsman := getTestLoftsman("avast")
	loftsman.Settings.Manifest.Path = ""
	loftsman.Settings.Manifest.Name = ""
	err := loftsman.Avast()
	if err == nil || !strings.Contains(err.Error(), "Unable to determine manifest name") {
		t.Errorf("Didn't get expected error from loftsman.TestAvastMissingManifestSettings(), instead got: %s", err)
	}
}

func TestAvastActiveConfigmapNotFound(t *testing.T) {
	loftsman := getTestLoftsman("avast")
	loftsman.Settings.Manifest.Name = "test-manifest"
	loftsman.reader = strings.NewReader("yes")
	err := loftsman.Avast()
	if err == nil || !strings.Contains(err.Error(), "Couldn't find an active ship in progress") {
		t.Errorf("Didn't get expected error from loftsman.TestAvastActiveConfigmapNotFound(), instead got: %s", err)
	}
}

func TestAvastActiveConfigmapFound(t *testing.T) {
	loftsman := getTestLoftsman("avast")
	loftsman.kubernetes = custommocks.GetKubernetesMock(true)
	loftsman.Settings.Manifest.Name = "test-manifest"
	loftsman.reader = strings.NewReader("yes")
	err := loftsman.Avast()
	if err != nil {
		t.Errorf("Got unexpected error from loftsman.TestAvastActiveConfigmapFound(): %s", err)
	}
}

func TestAvastManifestPathActiveConfigmapFound(t *testing.T) {
	loftsman := getTestLoftsman("avast")
	loftsman.kubernetes = custommocks.GetKubernetesMock(true)
	loftsman.Settings.Manifest.Path = "./.test-fixtures/manifest-v1beta1.yaml"
	loftsman.reader = strings.NewReader("yes")
	err := loftsman.Avast()
	if err != nil {
		t.Errorf("Got unexpected error from loftsman.TestAvastManifestPathActiveConfigmapFound(): %s", err)
	}
}

func TestAvastConfirmNo(t *testing.T) {
	loftsman := getTestLoftsman("avast")
	loftsman.kubernetes = custommocks.GetKubernetesMock(true)
	loftsman.Settings.Manifest.Path = "./.test-fixtures/manifest-v1beta1.yaml"
	loftsman.reader = strings.NewReader("no")
	err := loftsman.Avast()
	if err != nil {
		t.Errorf("Got unexpected error from loftsman.TestAvastConfirmNo(): %s", err)
	}
}
