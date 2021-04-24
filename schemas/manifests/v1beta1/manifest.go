package v1beta1

import (
	"crypto/md5"
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

// Load will load manifest string/file/byte content into a v1beta1.Manifest object
func (m *Manifest) Load(manifestContent string) error {
	if err := yaml.Unmarshal([]byte(manifestContent), &m); err != nil {
		return err
	}
	// We haven't yet validated our resulting, loaded manifest, so we can't assume anything is initialized
	// TODO: is this going to be a good idea thinking about validating a manifest that's been modified instead
	//       of always just the original? Not something we have to definitely answer on this round while supporting
	//       just all.timeout, but when we get to some more generic possibilities (see comment below) we'll
	//       probably want to settle on some answers
	if m.Spec != nil && len(m.Spec.Charts) > 0 {
		for _, chart := range m.Spec.Charts {
			if m.Spec.All != nil {
				// TODO: we'll eventually use some generic merge capability here, since we're only supporting all.timeout
				//       for now, we can simply assume that's the only thing we care about
				if m.Spec.All.Timeout != "" && chart.Timeout == "" {
					chart.Timeout = m.Spec.All.Timeout
				}
			}
		}
	}
	return nil
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
	var err error
	var releaseErrors []*interfaces.ManifestReleaseError
	addedRepos := []string{}

CHARTS:
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
				continue CHARTS
			}
			removedFailedRelease = true
		}

		m.logger.SubHeader(fmt.Sprintf("Releasing %s v%s", chart.Name, chart.Version))
		if removedFailedRelease {
			logForChart(zerolog.InfoLevel, "Removed previously-failed first release successfully")
		}

		// TODO: when we're able to deprecate --charts-* CLI args, we can move to some slightly cleaner patterns here. In order to continue to
		//       support both manifest-defined chart sources and the CLI ones, and doing as little as possible around it for now, this is deemed the
		//       best path
		extraCmdArgs := ""
		helmChartsSource := &interfaces.HelmChartsSource{}
		if m.Spec.Sources != nil && len(m.Spec.Sources.Charts) > 0 {
			foundSource := false
			for _, chartSource := range m.Spec.Sources.Charts {
				if chartSource.Name == chart.Source {
					foundSource = true
					if chartSource.Type == ChartSourceTypeRepo {
						helmChartsSource.RepoName = chartSource.Name
						helmChartsSource.Repo = chartSource.Location
						if chartSource.CredentialsSecret != nil {
							helmChartsSource.RepoUsername, err = kubernetes.GetSecretKeyValue(chartSource.CredentialsSecret.Name, chartSource.CredentialsSecret.Namespace,
								chartSource.CredentialsSecret.UsernameKey)
							if err != nil {
								recordReleaseError(fmt.Errorf("Error getting chart source username from secret %s for spec.sources.charts[] name = %s: %s",
									chartSource.CredentialsSecret.Name, chartSource.Name, err))
								continue CHARTS
							}
							helmChartsSource.RepoPassword, err = kubernetes.GetSecretKeyValue(chartSource.CredentialsSecret.Name, chartSource.CredentialsSecret.Namespace,
								chartSource.CredentialsSecret.PasswordKey)
							if err != nil {
								recordReleaseError(fmt.Errorf("Error getting chart source password from secret %s for spec.sources.charts[] name = %s: %s",
									chartSource.CredentialsSecret.Name, chartSource.Name, err))
								continue CHARTS
							}
						}
					} else if chartSource.Type == ChartSourceTypeDirectory {
						helmChartsSource.Path = chartSource.Location
					}
					if err = helm.Initialize(helm.GetExecConfig(), helmChartsSource); err != nil {
						recordReleaseError(fmt.Errorf("Error re-initializing Helm for specific source %s for chart %s: %s", chart.Source, chart.Name, err))
						continue CHARTS
					}
					break
				}
			}
			if !foundSource {
				recordReleaseError(fmt.Errorf("Source name not found in spec.sources.charts[]: %s", chart.Source))
				continue CHARTS
			}
		}

		availableVersions, err := helm.GetAvailableChartVersions(chart.Name)
		if err != nil {
			recordReleaseError(fmt.Errorf("Error determining available versions for the chart %s: %s", chart.Name, err))
			continue CHARTS
		}
		chartPath := ""
		for _, availableVersion := range availableVersions {
			if chart.Version == availableVersion.Version {
				chartPath = availableVersion.Path
			}
		}
		if chartPath == "" {
			recordReleaseError(fmt.Errorf("Unable to find chart %s v%s in the configured charts location", chart.Name, chart.Version))
			continue CHARTS
		}

		releaseName := chart.Name
		if chart.ReleaseName != "" {
			releaseName = chart.ReleaseName
		}
		if chart.Timeout != "" {
			extraCmdArgs = fmt.Sprintf("%s --timeout %s", extraCmdArgs, chart.Timeout)
		}
		if helmChartsSource.RepoUsername != "" {
			if helmChartsSource.RepoName == "" {
				helmChartsSource.RepoName = fmt.Sprintf("%x", md5.Sum([]byte(helmChartsSource.Repo)))
			}
			isAlreadyAdded := false
			for _, addedRepo := range addedRepos {
				if helmChartsSource.RepoName == addedRepo {
					isAlreadyAdded = true
					break
				}
			}
			if !isAlreadyAdded {
				_, err := helm.Exec(fmt.Sprintf("repo add %s %s --username %s --password %s", helmChartsSource.RepoName,
					helmChartsSource.Repo, helmChartsSource.RepoUsername, helmChartsSource.RepoPassword))
				if err != nil {
					recordReleaseError(fmt.Errorf("Error adding secure chart repo %s: %s", helmChartsSource.Repo, err))
					continue CHARTS
				}
				addedRepos = append(addedRepos, helmChartsSource.RepoName)
			}
			chartPath = fmt.Sprintf("%s/%s", helmChartsSource.RepoName, chart.Name)
			extraCmdArgs = fmt.Sprintf("%s --version %s", extraCmdArgs, chart.Version)
		}
		installUpgradeCmd := strings.TrimSpace(fmt.Sprintf(
			"upgrade --install %s %s --namespace %s --create-namespace --set global.chart.name=%s --set global.chart.version=%s %s",
			releaseName,
			chartPath,
			chart.Namespace,
			chart.Name,
			chart.Version,
			strings.TrimSpace(extraCmdArgs),
		))
		if chart.Values != nil {
			valuesBytes, err := yaml.Marshal(chart.Values)
			if err != nil {
				recordReleaseError(fmt.Errorf("Error parsing override values for chart %s: %s", chart.Name, err))
				continue CHARTS
			}
			valuesFilePath := filepath.Join(m.tempDirectory, fmt.Sprintf("%s-values.yaml", chart.Name))
			if err = ioutil.WriteFile(valuesFilePath, valuesBytes, 0644); err != nil {
				recordReleaseError(fmt.Errorf("Error writing Helm values for for chart %s: %s", chart.Name, err))
				continue CHARTS
			}
			logForChart(zerolog.InfoLevel, fmt.Sprintf("Found value overrides for chart, applying: \n%s", valuesBytes))
			installUpgradeCmd = fmt.Sprintf("%s -f %s", installUpgradeCmd, valuesFilePath)
		}
		logForChart(zerolog.InfoLevel, fmt.Sprintf("Running helm install/upgrade with arguments: %s", installUpgradeCmd))
		output, err := helm.Exec(installUpgradeCmd)
		if err != nil {
			recordReleaseError(fmt.Errorf("Error releasing chart %s v%s: %s", chart.Name, chart.Version, err))
			continue CHARTS
		}
		logForChart(zerolog.InfoLevel, fmt.Sprintf("%s\n", output))
	}
	for _, addedRepo := range addedRepos {
		_, _ = helm.Exec(fmt.Sprintf("repo rm %s", addedRepo))
	}
	return releaseErrors
}
