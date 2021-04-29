### 1.1.0 Changelog

* Support for pointing at multiple chart sources, both local and remote, added to manifest schema
* Implement a global and per-chart Helm install/upgrade timeout override in the manifest
* Ability to set a custom release name for a chart in the manifest, previously would have to be the chart name
* Improve log file configuration, previously was just `loftsman.log` output always on all commands, we've moved towards a default of not outputting
* Add pipeline/action to publish an official container image for the Loftman CLI
* Initial functional testing script added to repo
