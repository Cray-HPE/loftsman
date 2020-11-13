package helm

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	shellmocks "github.com/Cray-HPE/go-lib/mocks/shell"
	"github.com/Cray-HPE/go-lib/shell"
	"github.com/Cray-HPE/loftsman/internal/interfaces"
	"github.com/stretchr/testify/mock"

	"github.com/jarcoal/httpmock"
)

var testRepoIndexYAMLTemplate = `---
apiVersion: v1
entries:
  chart1:
  - apiVersion: v2
    created: "2020-01-28T20:31:31.354036767Z"
    description: test
    digest: 4f10684ec0d05612c6667d01a66ece6b92d333bb21b19e7163dab12f7563051
    name: chart1
    urls:
    - %scharts/chart1-0.1.0.tgz
    version: 0.1.0
  - apiVersion: v2
    created: "2020-01-28T20:31:31.354036767Z"
    description: test
    digest: 4f10684ec0d05612c6667d01a66ece6b92d333bb21b19e7163dab12f7563051
    name: chart1
    urls:
    - %scharts/chart1-0.1.1.tgz
    version: 0.1.1
  chart2:
  - apiVersion: v2
    created: "2020-01-28T20:31:31.354036767Z"
    description: test
    digest: 4f10684ec0d05612c6667d01a66ece6b92d333bb21b19e7163dab12f7563051
    name: chart2
    urls:
    - %scharts/chart2-0.2.0.tgz
    version: 0.2.0
`

var execError error

func setExecError(msg string) {
	execError = errors.New(msg)
}
func resetExecError() {
	execError = nil
}

func getMockExecConfig(useHelmV2 bool) *interfaces.HelmExecConfig {
	shellMock := &shellmocks.Interface{}
	shellMock.On("Exec", mock.AnythingOfType("string"), mock.AnythingOfType("shell.ExecOptions")).Return(func(command string, options shell.ExecOptions) string {
		if command == "helm version --client" && !useHelmV2 {
			return `version.BuildInfo{Version:"v3.2.0", GitCommit:"e11b7ce3b12db2941e90399e874513fbd24bcb71", GitTreeState:"clean", GoVersion:"go1.13.10"}`
		}
		if command == "helm version --client" && useHelmV2 {
			return `Client: &version.Version{SemVer:"v2.16.6", GitCommit:"dd2e5695da88625b190e6b22e9542550ab503a47", GitTreeState:"clean"}`
		}
		if strings.Contains(command, "status test-release-status-invalid-status") {
			return `---
info:
status: deployed
	`
		}
		if strings.Contains(command, "status test-release-status") {
			return `---
info:
  status: deployed`
		}
		return command
	}, execError)
	return &interfaces.HelmExecConfig{
		Shell:          shellMock,
		Binary:         "helm",
		KubeconfigPath: "",
		KubeContext:    "",
	}
}

func TestInitialize(t *testing.T) {
	h := &Helm{}
	err := h.Initialize(getMockExecConfig(false), &interfaces.HelmChartsSource{})
	if err != nil {
		t.Errorf("Got unexpected error from helm.TestInitialize(): %s", err)
	}
}

func TestInitializeOldHelmVersion(t *testing.T) {
	h := &Helm{}
	err := h.Initialize(getMockExecConfig(true), &interfaces.HelmChartsSource{})
	if err == nil || !strings.Contains(err.Error(), "v3 client binary is required") {
		t.Errorf("Didn't get expected error from helm.TestInitializeOldHelmVersion(), instead got: %s", err)
	}
}

func TestNewExecError(t *testing.T) {
	setExecError("new-exec-error")
	defer resetExecError()
	h := &Helm{}
	err := h.Initialize(getMockExecConfig(false), &interfaces.HelmChartsSource{})
	if err == nil || err.Error() != "new-exec-error" {
		t.Errorf("Didn't get expected error from helm.TestNewExecError(), instead got: %s", err)
	}
}

func TestExec(t *testing.T) {
	h := &Helm{}
	err := h.Initialize(getMockExecConfig(false), &interfaces.HelmChartsSource{})
	if err != nil {
		t.Errorf("Got unexpected error from helm.Initialize() in helm.TestExec(): %s", err)
		return
	}
	out, err := h.Exec("ls")
	if err != nil {
		t.Errorf("Got unexpected error from helm.TestExec(): %s", err)
		return
	}
	if out != "helm ls" {
		t.Errorf("Didn't get expected result from helm.TestExec(), instead got: %s", out)
	}
}

func TestExecKubeconfigAndContext(t *testing.T) {
	h := &Helm{}
	err := h.Initialize(getMockExecConfig(false), &interfaces.HelmChartsSource{})
	if err != nil {
		t.Errorf("Got unexpected error from helm.Initialize() in helm.TestExecKubeconfigAndContext(): %s", err)
		return
	}
	h.ExecConfig.KubeconfigPath = "kubeconfig"
	h.ExecConfig.KubeContext = "default"
	out, err := h.Exec("ls")
	if err != nil {
		t.Errorf("Got unexpected error from helm.TestExecKubeconfigAndContext(): %s", err)
		return
	}
	if out != "helm --kubeconfig kubeconfig --kube-context default ls" {
		t.Errorf("Didn't get expected result from helm.TestExecKubeconfigAndContext(), instead got: %s", out)
	}
}

func TestGetAvailableChartVersionsWithLocalPath(t *testing.T) {
	h := &Helm{}
	chartsPath := ".test-fixtures/charts"
	err := h.Initialize(getMockExecConfig(false), &interfaces.HelmChartsSource{
		Path: chartsPath,
	})
	if err != nil {
		t.Errorf("Got unexpected error from helm.Initialize() in helm.TestGetAvailableChartVersionsWithLocalPath(): %s", err)
		return
	}
	available, err := h.GetAvailableChartVersions("chart1")
	if err != nil {
		t.Errorf("Got unexpected error from helm.TestGetAvailableChartVersionsWithLocalPath(): %s", err)
		return
	}
	if len(available) != 2 {
		t.Errorf("Didn't get expected size for available list from helm.TestGetAvailableChartVersionsWithLocalPath(), expected 2, but got %d", len(available))
	}
	for _, version := range available {
		if version.Version != "0.1.0" && version.Version != "0.1.1" {
			t.Errorf("Found unexpected version in available list from helm.TestGetAvailableChartVersionsWithLocalPath(): %s", version.Version)
		}
		if version.Path != filepath.Join(chartsPath, "chart1-0.1.0.tgz") && version.Path != filepath.Join(chartsPath, "chart1-0.1.1.tgz") {
			t.Errorf("Found unexpected path in available list from helm.TestGetAvailableChartVersionsWithLocalPath(): %s", version.Path)
		}
	}
}

func TestGetAvailableChartVersionsWithRepo(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	indexYAML := fmt.Sprintf(testRepoIndexYAMLTemplate, "", "", "")
	httpmock.RegisterResponder("GET", `=~^http://charts\.io`, httpmock.NewStringResponder(200, indexYAML))
	h := &Helm{}
	err := h.Initialize(getMockExecConfig(false), &interfaces.HelmChartsSource{
		Repo: "http://charts.io",
	})
	if err != nil {
		t.Errorf("Got unexpected error from helm.Initialize() in helm.TestGetAvailableChartVersionsWithRepo(): %s", err)
		return
	}
	available, err := h.GetAvailableChartVersions("chart1")
	if err != nil {
		t.Errorf("Got unexpected error from helm.TestGetAvailableChartVersionsWithRepo(): %s", err)
		return
	}
	if len(available) != 2 {
		t.Errorf("Didn't get expected size for available list from helm.TestGetAvailableChartVersionsWithRepo(), expected 2, but got %d", len(available))
	}
	for _, version := range available {
		if version.Version != "0.1.0" && version.Version != "0.1.1" {
			t.Errorf("Found unexpected version in available list from helm.TestGetAvailableChartVersionsWithRepo(): %s", version.Version)
		}
		if version.Path != "http://charts.io/charts/chart1-0.1.0.tgz" && version.Path != "http://charts.io/charts/chart1-0.1.1.tgz" {
			t.Errorf("Found unexpected path in available list from helm.TestGetAvailableChartVersionsWithRepo(): %s", version.Path)
		}
	}
}

func TestGetAvailableChartVersionsWithRepoWithCreds(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	indexYAML := fmt.Sprintf(testRepoIndexYAMLTemplate, "", "", "")
	httpmock.RegisterResponder("GET", `=~^http://charts\.io`, httpmock.NewStringResponder(200, indexYAML))
	h := &Helm{}
	err := h.Initialize(getMockExecConfig(false), &interfaces.HelmChartsSource{
		Repo:         "http://charts.io",
		RepoUsername: "user",
		RepoPassword: "pass",
	})
	if err != nil {
		t.Errorf("Got unexpected error from helm.Initialize() in helm.TestGetAvailableChartVersionsWithRepoWithCreds(): %s", err)
		return
	}
	available, err := h.GetAvailableChartVersions("chart1")
	if err != nil {
		t.Errorf("Got unexpected error from helm.TestGetAvailableChartVersionsWithRepoWithCreds(): %s", err)
		return
	}
	if len(available) != 2 {
		t.Errorf("Didn't get expected size for available list from helm.TestGetAvailableChartVersionsWithRepoWithCreds(), expected 2, but got %d", len(available))
	}
	for _, version := range available {
		if version.Version != "0.1.0" && version.Version != "0.1.1" {
			t.Errorf("Found unexpected version in available list from helm.TestGetAvailableChartVersionsWithRepoWithCreds(): %s", version.Version)
		}
		if version.Path != "http://charts.io/charts/chart1-0.1.0.tgz" && version.Path != "http://charts.io/charts/chart1-0.1.1.tgz" {
			t.Errorf("Found unexpected path in available list from helm.TestGetAvailableChartVersionsWithRepoWithCreds(): %s", version.Path)
		}
	}
}

func TestGetAvailableChartVersionsWithRepoFullURLs(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	indexYAML := fmt.Sprintf(testRepoIndexYAMLTemplate, "http://charts.io/", "http://charts.io/", "http://charts.io/")
	httpmock.RegisterResponder("GET", `=~^http://charts\.io`, httpmock.NewStringResponder(200, indexYAML))
	h := &Helm{}
	err := h.Initialize(getMockExecConfig(false), &interfaces.HelmChartsSource{
		Repo: "http://charts.io",
	})
	if err != nil {
		t.Errorf("Got unexpected error from helm.Initialize() in helm.TestGetAvailableChartVersionsWithRepoFullURLs(): %s", err)
		return
	}
	available, err := h.GetAvailableChartVersions("chart1")
	if err != nil {
		t.Errorf("Got unexpected error from helm.TestGetAvailableChartVersionsWithRepoFullURLs(): %s", err)
		return
	}
	if len(available) != 2 {
		t.Errorf("Didn't get expected size for available list from helm.TestGetAvailableChartVersionsWithRepoFullURLs(), expected 2, but got %d", len(available))
	}
	for _, version := range available {
		if version.Version != "0.1.0" && version.Version != "0.1.1" {
			t.Errorf("Found unexpected version in available list from helm.TestGetAvailableChartVersionsWithRepoFullURLs(): %s", version.Version)
		}
		if version.Path != "http://charts.io/charts/chart1-0.1.0.tgz" && version.Path != "http://charts.io/charts/chart1-0.1.1.tgz" {
			t.Errorf("Found unexpected path in available list from helm.TestGetAvailableChartVersionsWithRepoFullURLs(): %s", version.Path)
		}
	}
}

func TestGetReleaseStatus(t *testing.T) {
	h := &Helm{}
	err := h.Initialize(getMockExecConfig(false), &interfaces.HelmChartsSource{})
	if err != nil {
		t.Errorf("Got unexpected error from helm.Initialize() in helm.TestGetReleaseStatus(): %s", err)
		return
	}
	rs, err := h.GetReleaseStatus("test-release-status", "default")
	if err != nil {
		t.Errorf("Got unexpected error from helm.TestGetReleaseStatus(): %s", err)
	}
	if rs.Info.Status != "deployed" {
		t.Errorf("Didn't get expected 'deployed' status from helm.GetReleaseStatus(), instead got: %v", rs)
	}
}

func TestGetReleaseStatusInvalidStatus(t *testing.T) {
	h := &Helm{}
	err := h.Initialize(getMockExecConfig(false), &interfaces.HelmChartsSource{})
	if err != nil {
		t.Errorf("Got unexpected error from helm.Initialize() in helm.TestGetReleaseStatusInvalidStatus(): %s", err)
		return
	}
	_, err = h.GetReleaseStatus("test-release-status-invalid-status", "default")
	if err == nil {
		t.Errorf("Didn't get expected error from helm.TestGetReleaseStatusInvalidStatus(): %s", err)
	}
}
