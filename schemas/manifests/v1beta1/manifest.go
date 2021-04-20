package v1beta1

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/Cray-HPE/loftsman/internal/interfaces"
	"github.com/Cray-HPE/loftsman/internal/logger"
	"github.com/rs/zerolog"
	yaml "gopkg.in/yaml.v2"
)

type releaseError struct {
	Chart     string
	Version   string
	Namespace string
	Error     error
}

// SetLogger sets the logger to use
func (m *Manifest) SetLogger(log *logger.Logger) {
	m.logger = log
}

// SetTempDirectory will set the path to use for temp files/directories
func (m *Manifest) SetTempDirectory(tempDirectory string) {
	m.tempDirectory = tempDirectory
}

// GetName will return the unique name for this manifest
func (m *Manifest) GetName() string {
	return m.Metadata.Name
}

// Load will load manifest string/file/byte content into a v1.Manifest object
func (m *Manifest) Load(manifestContent string) error {
	return yaml.Unmarshal([]byte(manifestContent), &m)
}

// Create will make a baseline manifest for this version
func (m *Manifest) Create(initializeCharts []string) (string, error) {
	var charts []*Chart
	for _, initializeChart := range initializeCharts {
		charts = append(charts, &Chart{
			Name:      initializeChart,
			Namespace: "",
			Version:   "",
		})
	}
	manifest := &Manifest{
		APIVersion: APIVersion,
		Metadata: &Metadata{
			Name: "",
		},
		Spec: &Spec{
			Charts: charts,
		},
	}
	manifestContent, err := yaml.Marshal(manifest)
	if err != nil {
		return "", err
	}
	return string(manifestContent), nil
}

// Release will run a full release/install/upgrade of all charts in the manifest
func (m *Manifest) Release(kubernetes interfaces.Kubernetes, helm interfaces.Helm) []*interfaces.ManifestReleaseError {
	var releaseErrors []*interfaces.ManifestReleaseError

	for _, chart := range m.Spec.Charts {
		logForChart := func(level zerolog.Level, msg string) {
			if strings.TrimSpace(msg) == "" {
				return
			}
			m.logger.WithLevel(level).
				Str("chart", chart.Name).
				Str("version", chart.Version).
				Str("namespace", chart.Namespace).
				Msg(msg)
		}
		recordReleaseError := func(releaseErr error) {
			releaseErrors = append(releaseErrors, &interfaces.ManifestReleaseError{
				Chart:     chart.Name,
				Version:   chart.Version,
				Namespace: chart.Namespace,
				Error:     releaseErr,
			})
			logForChart(zerolog.ErrorLevel, strings.TrimSpace(releaseErr.Error()))
		}
		releaseStatus, _ := helm.GetReleaseStatus(chart.Name, chart.Namespace)
		removedFailedRelease := false
		if releaseStatus.Info != nil && releaseStatus.Info.Status == "failed" && releaseStatus.Revision == 1 {
			// in the case of a failed release from the first install, we want to remove it before attempting an "upgrade": https://github.com/helm/helm/issues/3353
			logForChart(zerolog.InfoLevel, fmt.Sprintf("Attempting to remove previously-failed first release for %s", chart.Name))
			_, err := helm.Exec(fmt.Sprintf("uninstall %s --namespace %s --no-hooks", chart.Name, chart.Namespace))
			if err != nil {
				recordReleaseError(fmt.Errorf("Error attempting to remove previously-failed first release for %s: %s", chart.Name, err))
				continue
			}
			removedFailedRelease = true
		}

		m.logger.SubHeader(fmt.Sprintf("Releasing %s v%s", chart.Name, chart.Version))
		if removedFailedRelease {
			logForChart(zerolog.InfoLevel, "Removed previously-failed first release successfully")
		}

		availableVersions, err := helm.GetAvailableChartVersions(chart.Name)
		if err != nil {
			recordReleaseError(fmt.Errorf("Error determining available versions for the chart %s: %s", chart.Name, err))
			continue
		}
		chartPath := ""
		for _, availableVersion := range availableVersions {
			if chart.Version == availableVersion.Version {
				chartPath = availableVersion.Path
			}
		}
		if chartPath == "" {
			recordReleaseError(fmt.Errorf("Unable to find chart %s v%s in the configured charts location", chart.Name, chart.Version))
			continue
		}

		installUpgradeCmd := strings.TrimSpace(fmt.Sprintf(
			"upgrade --install %s %s --namespace %s --create-namespace --set global.chart.name=%s --set global.chart.version=%s %s",
			chart.Name,
			chartPath,
			chart.Namespace,
			chart.Name,
			chart.Version,
			m.getHelmTimeoutArg(chart),
		))
		if chart.Values != nil {
			valuesBytes, err := yaml.Marshal(chart.Values)
			if err != nil {
				recordReleaseError(fmt.Errorf("Error parsing override values for chart %s: %s", chart.Name, err))
				continue
			}
			valuesFilePath := filepath.Join(m.tempDirectory, fmt.Sprintf("%s-values.yaml", chart.Name))
			if err = ioutil.WriteFile(valuesFilePath, valuesBytes, 0644); err != nil {
				recordReleaseError(fmt.Errorf("Error writing Helm values for for chart %s: %s", chart.Name, err))
				continue
			}
			logForChart(zerolog.InfoLevel, fmt.Sprintf("Found value overrides for chart, applying: \n%s", valuesBytes))
			installUpgradeCmd = fmt.Sprintf("%s -f %s", installUpgradeCmd, valuesFilePath)
		}
		logForChart(zerolog.InfoLevel, fmt.Sprintf("Running helm install/upgrade with arguments: %s", installUpgradeCmd))
		output, err := helm.Exec(installUpgradeCmd)
		if err != nil {
			recordReleaseError(fmt.Errorf("Error releasing chart %s v%s: %s", chart.Name, chart.Version, err))
			continue
		}
		logForChart(zerolog.InfoLevel, fmt.Sprintf("%s\n", output))
	}
	return releaseErrors
}

func (m *Manifest) getHelmTimeoutArg(chart *Chart) string {
	if chart.Timeout != "" {
		return fmt.Sprintf("--timeout %s", chart.Timeout)
	} else if m.Spec.ChartTimeout != "" {
		return fmt.Sprintf("--timeout %s", m.Spec.ChartTimeout)
	}
	return ""
}
