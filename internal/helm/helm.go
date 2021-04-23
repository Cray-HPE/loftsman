// Package helm is for our default helm command object and operations
package helm

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Cray-HPE/go-lib/shell"
	"github.com/Cray-HPE/loftsman/internal/interfaces"
	yaml "gopkg.in/yaml.v2"
)

// Helm is our object for running helm command and operations, implements internal/interfaces/helm.go
type Helm struct {
	ExecConfig   *interfaces.HelmExecConfig
	ChartsSource *interfaces.HelmChartsSource
}

// ChartRepoIndexYAML is the root index of a chart repo
type ChartRepoIndexYAML struct {
	Entries map[string][]*ChartRepoEntryVersion `yaml:"entries"`
}

// ChartRepoEntryVersion is a single version for a chart in a chart repo index.yaml
type ChartRepoEntryVersion struct {
	URLs    []string `yaml:"urls"`
	Version string   `yaml:"version"`
}

// Initialize will set our instance up with necessary config/settings and run some initial validation as well
func (h *Helm) Initialize(execConfig *interfaces.HelmExecConfig, chartsSource *interfaces.HelmChartsSource) error {
	h.ExecConfig = execConfig
	h.ChartsSource = chartsSource
	versionOutput, err := h.Exec("version --client")
	if err != nil {
		return err
	}
	if h.ChartsSource.Repo != "" {
		if _, err = url.Parse(h.ChartsSource.Repo); err != nil {
			return fmt.Errorf("Charts repo url is invalid: %s", err)
		}
	}
	if !strings.Contains(versionOutput, `v3`) {
		return fmt.Errorf("Helm v3 client binary is required to run the Loftsman tool, found: %s", versionOutput)
	}
	return nil
}

// Exec will run a helm cli command/sub-command
func (h *Helm) Exec(subCommand string) (string, error) {
	command := h.ExecConfig.Binary
	if h.ExecConfig.KubeconfigPath != "" {
		command = fmt.Sprintf("%s --kubeconfig %s", command, h.ExecConfig.KubeconfigPath)
	}
	if h.ExecConfig.KubeContext != "" {
		command = fmt.Sprintf("%s --kube-context %s", command, h.ExecConfig.KubeContext)
	}
	return h.ExecConfig.Shell.Exec(fmt.Sprintf("%s %s", strings.TrimSpace(command), strings.TrimSpace(subCommand)),
		shell.ExecOptions{Silent: true, TrimOutput: true})
}

// GetAvailableChartVersions will return a list of available versions for a given chart according to our charts source
func (h *Helm) GetAvailableChartVersions(chartName string) ([]*interfaces.HelmAvailableChartVersion, error) {
	var available []*interfaces.HelmAvailableChartVersion
	if h.ChartsSource.Path != "" {
		localChartFiles, err := ioutil.ReadDir(h.ChartsSource.Path)
		if err != nil {
			return available, err
		}
		for _, localChartFile := range localChartFiles {
			matched, _ := regexp.MatchString(fmt.Sprintf("^%s-", chartName), localChartFile.Name())
			if matched {
				version := strings.ReplaceAll(localChartFile.Name(), fmt.Sprintf("%s-", chartName), "")
				version = strings.ReplaceAll(version, ".tgz", "")
				available = append(available, &interfaces.HelmAvailableChartVersion{
					Path:    filepath.Join(h.ChartsSource.Path, localChartFile.Name()),
					Version: version,
				})
			}
		}
	} else if h.ChartsSource.Repo != "" {
		indexURL, _ := url.Parse(h.ChartsSource.Repo)
		indexURL.Path = path.Join(indexURL.Path, "index.yaml")
		httpClient := &http.Client{}
		req, err := http.NewRequest("GET", indexURL.String(), nil)
		if err != nil {
			return available, err
		}
		if h.ChartsSource.RepoUsername != "" && h.ChartsSource.RepoPassword != "" {
			req.SetBasicAuth(h.ChartsSource.RepoUsername, h.ChartsSource.RepoPassword)
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return available, err
		}
		indexYAMLBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return available, err
		}
		var indexYAML *ChartRepoIndexYAML
		if err = yaml.Unmarshal(indexYAMLBytes, &indexYAML); err != nil {
			return available, err
		}
		for entryChartName, entryChartVersions := range indexYAML.Entries {
			if entryChartName == chartName {
				for _, version := range entryChartVersions {
					urlPath := version.URLs[0]
					if !strings.Contains(urlPath, h.ChartsSource.Repo) {
						fullURL, _ := url.Parse(h.ChartsSource.Repo)
						fullURL.Path = path.Join(fullURL.Path, version.URLs[0])
						urlPath = fullURL.String()
					}
					available = append(available, &interfaces.HelmAvailableChartVersion{
						Path:    urlPath,
						Version: version.Version,
					})
				}
				break
			}
		}
	}
	return available, nil
}

// GetReleaseStatus attempts to retrieve the status of a chart release
func (h *Helm) GetReleaseStatus(chartName string, chartNamespace string) (*interfaces.HelmReleaseStatus, error) {
	output, err := h.Exec(fmt.Sprintf("status %s --namespace %s --output yaml", chartName, chartNamespace))
	rs := &interfaces.HelmReleaseStatus{}
	if err != nil {
		return rs, err
	}
	if err = yaml.Unmarshal([]byte(output), &rs); err != nil {
		return rs, fmt.Errorf("error parsing release status info for %s: %s", chartName, err)
	}
	return rs, nil
}

// GetExecConfig returns the existing ExecConfig
func (h *Helm) GetExecConfig() *interfaces.HelmExecConfig {
	return h.ExecConfig
}
