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
      "properties": {
        "name": { "type": "string" },
        "source": { "type": "string" },
        "releaseName": { "type": "string" },
        "namespace": { "type": "string" },
        "version": { "type": "string" },
        "values": { "type": [ "object", "null" ] },
        "timeout": { "type": "string" }
      },
      "additionalProperties": false
    },
    "all": {
      "type": "object",
      "properties": {
        "timeout": { "type": "string" }
      },
      "additionalProperties": false
    }
  },
  "type": "object",
  "required": ["apiVersion", "metadata", "spec"],
  "properties": {
    "apiVersion": { "type": "string" },
    "metadata": {
      "type": "object",
      "properties": {
        "name": { "type": "string" },
        "labels": { "type": "object" }
      },
      "additionalProperties": false
    },
    "spec": {
      "type": "object",
      "required": ["charts"],
      "properties": {
        "sources": {
          "type": "object",
          "properties": {
            "charts": {
              "type": "array",
              "items": {
                "type": "object",
                "required": [ "type", "name", "location" ],
                "properties": {
                  "type": { "type": "string", "enum": ["directory", "repo"] },
                  "name": { "type": "string" },
                  "location": { "type": "string" },
                  "credentialsSecret": {
                    "type": "object",
                    "required": ["name", "namespace", "usernameKey", "passwordKey"],
                    "properties": {
                      "name": { "type": "string" },
                      "namespace": { "type": "string" },
                      "usernameKey": { "type": "string" },
                      "passwordKey": { "type": "string" }
                    },
                    "additionalProperties": false
                  }
                },
                "additionalProperties": true
              }
            },
            "repos": {
              "type": "array",
              "items": {
                "type": "object",
                "required": [ "name", "url" ],
                "properties": {
                  "name": { "type": "string" },
                  "url": { "type": "string" }
                },
                "additionalProperties": false
              }
            }
          },
          "additionalProperties": false
        },
        "all": { "$ref": "#/definitions/all" },
        "charts": {
          "type": "array",
          "items": {
            "allOf": [
              { "$ref": "#/definitions/chart" },
              { "required": ["name", "namespace", "version"] }
            ]
          }
        }
      },
      "additionalProperties": false,
      "dependencies": {
        "sources": {
          "properties": {
            "charts": {
              "type": "array",
              "items": {
                "allOf": [
                  { "$ref": "#/definitions/chart" },
                  { "required": ["name", "source", "namespace", "version"] }
                ]
              }
            }
          }
        }
      }
    }
  },
  "additionalProperties": false
}
`
