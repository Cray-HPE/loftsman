// Package manifest is for manifest-related resources
package manifest

import (
	"errors"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
	yaml "gopkg.in/yaml.v2"
	"github.com/Cray-HPE/loftsman/internal/interfaces"
	"github.com/Cray-HPE/loftsman/schemas/manifests/v1beta1"
)

// OnlyAPIVersion is just a struct with apiVersion field, which should be common across all versions
// useful so we can parse a manifest early to get this value for any manifest version and nothing else
type OnlyAPIVersion struct {
	APIVersion string `yaml:"apiVersion" json:"apiVersion"`
}

func getVersion(manifestContent string) (string, error) {
	var err error
	var manifest *OnlyAPIVersion
	err = yaml.Unmarshal([]byte(manifestContent), &manifest)
	if err != nil {
		return "", errors.New("could not parse the manifest as yaml to retrieve the apiVersion")
	}
	return manifest.APIVersion, nil
}

// Validate will accept a string that is the manifest file content and validate it
func Validate(manifestContent string) (interfaces.Manifest, error) {
	var err error
	var manifest interfaces.Manifest
	var schema gojsonschema.JSONLoader
	var document gojsonschema.JSONLoader
	apiVersion, err := getVersion(manifestContent)
	if err != nil {
		return nil, err
	}
	switch apiVersion {
	case v1beta1.APIVersion:
		manifest = &v1beta1.Manifest{}
		schema = gojsonschema.NewStringLoader(v1beta1.Schema)
	default:
		return nil, fmt.Errorf("the manifest apiVersion is not supported: %s", apiVersion)
	}
	if err = manifest.Load(manifestContent); err != nil {
		return nil, fmt.Errorf("could not parse the manifest as %s yaml: %s", v1beta1.APIVersion, err)
	}
	document = gojsonschema.NewGoLoader(manifest)
	result, err := gojsonschema.Validate(schema, document)
	if err != nil {
		return nil, fmt.Errorf("error running manifest schema validation: %s", err)
	}
	if !result.Valid() {
		errorList := "manifest validation errors:"
		for i, validationErr := range result.Errors() {
			errorList = fmt.Sprintf("%s (%d) %v", errorList, (i + 1), validationErr)
		}
		return nil, errors.New(errorList)
	}
	return manifest, nil
}

// Create is the entrypoint for creating a baseline manifest for the most-recent schema version
func Create(initializeCharts []string) (string, error) {
	manifestV1Beta1 := v1beta1.Manifest{}
	return manifestV1Beta1.Create(initializeCharts)
}
