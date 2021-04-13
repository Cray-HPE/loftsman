### 1.0.4 Changelog

* Move to Go 1.16
* Add initial Github actions for publishing artifacts to GitHub releases
* Build out Github templates for issues and PRs
* Adjust other documentation, build utilities, etc. for OSS
* First, initial repo-hosted documentation, including addition of `manifests/v1beta1` comprehensive example
* Removed lingering `Replaces` struct property, remained from some early v1 schema functionality proposal that didn't get implemented
* Add various kubernetes go-client auth libraries as they're not included automatically (ref https://github.com/kubernetes/client-go/issues/242)
* Removed previous pipeline internal-use only build and pipeline scripts, etc.
