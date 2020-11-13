package v1beta1

// AUTO-GENERATED FILE: DO NOT MODIFY

// Schema is the Go string variable container the JSON schema
const Schema = `
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$comment": "When editing this file in Loftsman source, make sure to re-generate schema .go files with ./scripts/generate-schema-go.sh",
  "title": "Loftsman manifests/v1beta1 Schema",
  "definitions": {
    "chart": {
      "type": "object",
      "required": [
        "name",
        "namespace",
        "version"
      ],
      "properties": {
        "name": { "type": "string" },
        "namespace": { "type": "string" },
        "version": { "type": "string" },
        "values": { "type": [ "object", "null" ] }
      }
    }
  },
  "type": "object",
  "required": [
    "apiVersion",
    "metadata",
    "spec"
  ],
  "properties": {
    "apiVersion": { "type": "string" },
    "metadata": {
      "type": "object",
      "properties": {
        "name": { "type": "string" },
        "labels": { "type": "object" }
      }
    },
    "spec": {
      "type": "object",
      "properties": {
        "charts": {
          "type": "array",
          "items": { "$ref": "#/definitions/chart" }
        }
      }
    }
  }
}
`
