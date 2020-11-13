package internal

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/stretchr/testify/mock"
	"github.com/Cray-HPE/loftsman/internal/interfaces"
	custommocks "github.com/Cray-HPE/loftsman/mocks/custom-mocks"
	mocks "github.com/Cray-HPE/loftsman/mocks/interfaces"
)

var releaseErrors []*interfaces.ManifestReleaseError

func setReleaseErrors(msg string) {
	releaseErrors = []*interfaces.ManifestReleaseError{
		&interfaces.ManifestReleaseError{
			Chart:     "tests",
			Version:   "0.0.0",
			Namespace: "default",
			Error:     errors.New(msg),
		},
	}
}
func resetReleaseErrors() {
	releaseErrors = []*interfaces.ManifestReleaseError{}
}

func getTestLoftsman(initializeForCommand string) *Loftsman {
	var availableChartVersions []*interfaces.HelmAvailableChartVersion
	loftsman := NewLoftsman()
	loftsman.Settings.JSONLog.Path = filepath.Join(os.TempDir(), "loftsman-tests-internal.log")
	loftsman.Settings.JSONLog.File, _ = os.OpenFile(loftsman.Settings.JSONLog.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	loftsman.Settings.Manifest.Path = "./.test-fixtures/manifest-v1beta1.yaml"
	loftsman.kubernetes = custommocks.GetKubernetesMock(false)
	loftsman.helm = custommocks.GetHelmMock(availableChartVersions)
	loftsman.manifest = getManifestMock()
	if initializeForCommand != "" {
		loftsman.Initialize(initializeForCommand)
	}
	return loftsman
}

func getManifestMock() *mocks.Manifest {
	m := &mocks.Manifest{}
	m.On("GetName").Return("test-manifest")
	m.On("SetLogger", mock.AnythingOfType("*logger.Logger"))
	m.On("SetTempDirectory", mock.AnythingOfType("string"))
	m.On("Release", mock.AnythingOfType("*mocks.Kubernetes"), mock.AnythingOfType("*mocks.Helm")).Return(releaseErrors)
	return m
}
