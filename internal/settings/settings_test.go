package settings

import (
	"reflect"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	s := New()
	settingsType := reflect.TypeOf(s)
	if settingsType.String() != "*settings.Settings" {
		t.Errorf("Expected settings.TestNew() to return type of 'settings.Settings', but returned %s", settingsType)
	}
}

func TestValidateChartsSourceWithChartsPathAndRepo(t *testing.T) {
	s := New()
	s.ChartsSource.Path = "../helm/.test-fixtures/charts"
	s.ChartsSource.Repo = "http://chartrepo"
	err := s.ValidateChartsSource()
	if err == nil || !strings.Contains(err.Error(), "both charts-repo and charts-path are set") {
		t.Errorf("Didn't get expected error from settings.ValidateChartsSource() when settings.ChartsSource.Path and settings.ChartsSource.Repo are both set, got: %s", err)
	}
}

func TestValidateChartsSourceChartsPathInvalid(t *testing.T) {
	s := New()
	s.ChartsSource.Path = "/path/that/does/not/exist/charts.tar.gz"
	err := s.ValidateChartsSource()
	if err == nil || !strings.Contains(err.Error(), "charts-path") || !strings.Contains(err.Error(), "not found") {
		t.Errorf("Didn't get expected error from settings.ValidateChartsSource() when settings.ChartsSource.Path doesn't exist, got: %s", err)
	}
}

func TestValidateChartsSourceChartsPathValid(t *testing.T) {
	s := New()
	s.ChartsSource.Path = "../helm/.test-fixtures/charts"
	err := s.ValidateChartsSource()
	if err != nil {
		t.Errorf("Got unexpected error from settings.ValidateChartsSource() when settings.ChartsSource.Path is valid, exists, got: %s", err)
	}
}

func TestValidateChartsSourceChartsRepoValid(t *testing.T) {
	s := New()
	s.ChartsSource.Repo = "http://chartsrepo"
	err := s.ValidateChartsSource()
	if err != nil {
		t.Errorf("Got unexpected error from settings.ValidateChartsSource() when settings.ChartsSource.Repo is valid, got: %s", err)
	}
}

func TestValidateManifestPathUnset(t *testing.T) {
	s := New()
	s.Manifest.Path = ""
	err := s.ValidateManifestPath()
	if err == nil || !strings.Contains(err.Error(), "manifest path") || !strings.Contains(err.Error(), "not found") {
		t.Errorf("Didn't get expected error from settings.ValidateManifestPath() when settings.Manifest.Path is unset, got: %s", err)
	}
}

func TestValidateManifestPathInvalid(t *testing.T) {
	s := New()
	s.Manifest.Path = "/path/that/does/not/exist/manifest.yaml"
	err := s.ValidateManifestPath()
	if err == nil || !strings.Contains(err.Error(), "manifest path") || !strings.Contains(err.Error(), "not found") {
		t.Errorf("Didn't get expected error from settings.ValidateManifestPath() when settings.Manifest.Path doesn't exist, got: %s", err)
	}
}

func TestValidateManifestPathSuccess(t *testing.T) {
	s := New()
	s.Manifest.Path = "../.test-fixtures/manifest-v1beta1.yaml"
	err := s.ValidateManifestPath()
	if err != nil {
		t.Errorf("Got unexpected error from settings.ValidateManifestPath() when settings.Manifest.Path exists, is valid: %s", err)
	}
}
