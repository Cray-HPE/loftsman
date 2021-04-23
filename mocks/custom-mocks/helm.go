package mocks

import (
	"errors"
	"strings"

	helminterface "github.com/Cray-HPE/loftsman/internal/interfaces"
	helmmocks "github.com/Cray-HPE/loftsman/mocks/interfaces"
	"github.com/stretchr/testify/mock"
)

// GetHelmMock will return a common mock for the Helm interface/object
func GetHelmMock(availableChartVersions []*helminterface.HelmAvailableChartVersion) *helmmocks.Helm {
	h := &helmmocks.Helm{}
	h.On("Initialize", mock.AnythingOfType("*interfaces.HelmExecConfig"), mock.AnythingOfType("*interfaces.HelmChartsSource")).Return(nil)
	h.On("Exec", mock.MatchedBy(func(command string) bool {
		if strings.Contains(command, "upgrade --install failed") {
			return true
		}
		return false
	})).Return("", errors.New("failed install/upgrade"))
	h.On("Exec", "uninstall failed-remove --namespace default").Return("", errors.New("failed removing failed release"))
	h.On("Exec", mock.AnythingOfType("string")).Return("", nil)
	h.On("GetAvailableChartVersions", mock.AnythingOfType("string")).Return(availableChartVersions, nil)
	h.On("GetReleaseStatus", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(func(chartName string, chartNamespace string) *helminterface.HelmReleaseStatus {
		rs := &helminterface.HelmReleaseStatus{
			Revision: 0,
			Info: &helminterface.HelmReleaseStatusInfo{
				Status: "",
			},
		}
		rs.Revision = 1
		if strings.Contains(chartName, "failed") {
			rs.Info.Status = "failed"
		}
		return rs
	}, nil)
	h.On("GetExecConfig").Return(&helminterface.HelmExecConfig{})
	return h
}
