```
 _        __ _                                |\
| | ___  / _| |_ ___ _ __ ___   __ _ _ __     | \
| |/ _ \| |_| __/ __|  _   _ \ / _  |  _ \    |  \
| | |_| |  _| |_\__ \ | | | | | |_| | | | |   |___\
|_|\___/|_|  \__|___/_| |_| |_|\__,_|_| |_|  \--||___/
                                ~~~~~~~~~~~~~~\_____/~~~~~~~~~~
```

> Loftsmen at the mould lofts of shipyards were responsible for taking the dimensions and details from drawings and plans, and translating this information into templates, battens, ordinates, cutting sketches, profiles, margins and other data.
>
> <cite>https://en.wikipedia.org/wiki/Lofting</cite>

## Define, organize, and ship your Kubernetes workloads with Helm charts easily

Loftsman is a manifest-based utility for managing the installation and upgrading of Kubernetes workloads via Helm charts. Some of the features include:

* Support for Helm v3 (and only Helm v3 in fact)
* Support for installing Helm charts from both a local directory or an authenticated charts repository
* Manifest-driven Helm chart installs, so defining the charts that are running in your cluster, the order in which they're installed/upgraded, and dynamic value overrides can all be managed by a manifest that Loftsman can help you create and understand
* A handful of release flow wrapper enhancements around Helm to address long-running issues with using Helm directly, such as https://github.com/helm/helm/issues/3353

## Developing and Testing

To develop loftsman, you'll need:

* a local Go 1.13 environment
* `golint` installed for running tests/linting locally
* `mockery` installed for regenerating mocks

### Incrementing the Version

Prior to a release, we need to change the version value in `/cmd/version/version.go`. All other versioning tasks will be handled for you in the pipeline, but this must be done on merge to master.
