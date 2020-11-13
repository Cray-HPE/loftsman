package manifest

import (
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestValidateInvalidAPIVersion(t *testing.T) {
	manifest := `---
apiVersion: v0
`
	_, err := Validate(manifest)
	if err == nil || err.Error() != "the manifest apiVersion is not supported: v0" {
		t.Errorf("Didn't get expected error from manifest.TestValidateInvalidAPIVersion(), instead got: %s", err)
	}
}

func TestCreate(t *testing.T) {
	manifestContent, err := Create([]string{"chart1", "chart2"})
	if err != nil {
		t.Errorf("Got unexpected error from TestManifestCreate: %s", err)
	}
	var manifest *OnlyAPIVersion
	err = yaml.Unmarshal([]byte(manifestContent), &manifest)
	if err != nil {
		t.Errorf("could not parse the manifest from TestManifestCreate() as yaml: %s", err)
	}
}
