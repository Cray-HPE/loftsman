# Contributing

Welcome to Loftsman! We're glad you're interested to contributing. We don't have too many overall contribution guidelines currently:

* Before starting work on any contribution, make sure you identify related [issues](https://github.com/Cray-HPE/loftsman/issues) and more importantly [any pull requests that might already exist](https://github.com/Cray-HPE/loftsman/pulls) to the same end.
* If this is your first contribution, consider [issues labeled "good first issue"](https://github.com/Cray-HPE/loftsman/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).

## Developing Loftsman

In order to develop Loftsman, you need to have the following:

* A Golang environment in which to work, version `1.16` currently
* Ideally some running Kubernetes cluster running on which to test your work

And that's really it, you can begin to work on the source locally, actively testing against a Kubernetes cluster as you go along. A few other things to help as you work:

```
./scripts/tests-unit.sh
```

The above unit test bash script is the same one that runs in official pipelines and can be used locally to both lint and test your changes or additions as you go along. **Always write new tests for any new code that you're hoping to contribute**.

If your work includes `Interface` additions, make sure you re-run the following before writing related tests and of course before submitting your pull request:

```
./scripts/generate-mocks.sh
```

The above will auto-generate mocks for `Interface` definitions that you can use while writing your new tests for the interface and related code that might use it.

For any test fixtures needed, use a `.test-fixtures` directory to store related files, data, etc. for this need.

## Releasing Loftsman

If you're a contributing owner of the Loftsman project, pre-releasing and releasing is pretty easy:

First, make sure you have a `CHANGELOG/CHANGELOG-[your release tag].md` file created with some description of your release or pre-release prior to tagging and pushing the tag.

* To pre-release, simply create a semantic version git tag like `^[0-9]+.[0-9]+.[0-9]+-.*$` and push it. GitHub Actions will take care of the rest to create a related release of artifacts marked as a pre-release
* To perform an official release, create a semantic version git tag like `^[0-9]+.[0-9]+.[0-9]+$` and push it. GitHub Actions will create a related official release of artifacts.
