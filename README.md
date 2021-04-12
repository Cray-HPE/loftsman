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

It's hard to orchestrate, manage, and observe Kubernetes cluster workloads as a whole. Loftsman takes a manifest you've defined for all of the things that should be present or running in your cluster and handles the rest.

### Features

* Simple, declarative manifests for defining everything to exist and run in your Kubernetes cluster in an organized and maintainable way
* Any easy command-line interface with minimal requirements for generating and shipping your manifests of cluster workloads
* Support for Kubernetes cluster install and upgrade scenarios and needs ranging from cloud-native to large-scale enterprise, such as air-gapped environments
* In support of GitOps release workflows, e.g. CI/CD listening for manifest changes in a repo and shipping new manifests when detected

### Features Coming Soon

* Ability to run Loftsman as an operator in your cluster and simply apply your manifest YAML to run installs and upgrades (https://github.com/Cray-HPE/loftsman/issues/10)
* Move to a manifest-driven approach for supporting multiple remote chart repos instead of providing this info via CLI args (issues and more info coming soon on this)

## Learn More about Loftsman and Start Using it

See the following for more information on how Loftsman can help you:

* [Installing and getting started with Loftsman](./docs#getting-started)
* Browse all [Loftsman documentation](./docs)

And, if you're interested, how you can help Loftsman:

* [Contributing to Loftsman](./CONTRIBUTING.md)
