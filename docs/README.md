# Loftsman Documentation

1. [About Loftsman](#about-loftsman)
2. [Getting Started](#getting-started)
    * [Requirements](#requirements)
    * [Installation](#installation)
    * [CLI Usage](#cli-usage)
    * [Building your First Manifest](#building-your-first-manifest)
    * [Shipping your Manifest](#shipping-your-manifest)
    * [Options for Shipping](#options-for-shipping)
3. [Next Steps in Working with Loftsman](#next-steps-in-working-with-loftsman)
    * [Understanding Loftsman Logs and Records](#understanding-loftsman-logs-and-records)
    * [Identifying and Fixing Errors](#identifying-and-fixing-errors)
    * [Manifest Schema Versions](#manifest-schema-versions)
        * [`manifests/v1beta1`](#manifests-v1beta1)

## About Loftsman

Loftsman exists to serve the need of easily defining and shipping everything you need to your Kubernetes cluster. Even with tools like [Helm](https://helm.sh) or [Kustomize](https://kustomize.io/), it can be difficult to coordinate installing and upgrading complex cluster workloads and following along with what's running, what's broken, and generally the state of your specific Kubernetes environment. Loftsman promises to solve these needs.

By allowing the user to define a manifest that's easy to write, easy to read, and easy to maintain, Loftsman provides us with a tool for various release workflows, manual and automated alike. We're able to more-easily manage and record historical visibility into our cluster state as a whole.

## Getting Started

### Requirements

Loftsman has two basic requirements on the machine where the CLI is running:

* [Helm v3 Installed](https://helm.sh/docs/intro/install/)
* [A `kubeconfig` generated and available with appropriate access to the Kubernetes cluster](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/). Loftsman will look for this `kubeconfig` in the default location, but you can tell Loftsman if it's in a different location (see `loftsman --help`)

Once you have these requirements met, you're ready to install Loftsman and start shipping manifests to your cluster.

### Installation

You can install the `loftsman` CLI in a few different ways:

* From our [official, published releases](https://github.com/Cray-HPE/loftsman/releases)
* Using `go get`
    * compile the binary, version = `dev`: `go get -u github.com/Cray-HPE/loftsman`
    * If you want to compile with your own CLI version: `go get -u -ldflags "-X 'github.com/Cray-HPE/loftsman/cmd.Version=1.0.4-custom'" github.com/Cray-HPE/loftsman`

### CLI Usage

When in doubt about a particular part of the CLI, `--help` is meant to answer your questions:

```
loftsman --help
```

The root CLI `--help` argument will provide you with a starting point for getting the CLI usage help you need. You can also dive in further for usage help on subcommands, e.g.:

```
loftsman ship --help
```

### Building your First Manifest

_This first example will use an approach that uses charts packaged/downloaded to a local directory. See the [other options for shipping](#options-for-shipping) to understand alternatives such as installing charts from chart repositories._

_This guide does some shell work that assumes a \*nix shell environment, but should easily translate to other platforms and environments_

#### Pulling Helm Charts

In the example manifest, we'll define it to use and ship a few community-provided Helm charts. So, let's set up a workspace and download some charts:

```
$ mkdir -p loftsman-workspace/charts
$ cd loftsman-workspace
$ helm repo add victoria-metrics https://victoriametrics.github.io/helm-charts
$ helm repo update
$ helm pull --version 0.8.24 -d ./charts/ victoria-metrics/victoria-metrics-cluster
$ ls charts/
victoria-metrics-cluster-0.8.24.tgz
```
We've downloaded one of our charts locally, and we'll install another from a remote Helm repo location. Time to move over to defining a new manifest to use these charts.

Manifests are meant to be easy to read and maintain, so it's pretty easy to get started and understand the options available to you. It's recommended that you use the Loftsman CLI itself for getting started with any new manifest:

```
$ loftsman manifest create
apiVersion: manifests/v1beta1
metadata: {}
spec:
  sources:
    charts: []
  charts:
  - {}
```

In the above output, we see the most-basic initialized version of a new manifest.

_NOTE: Loftsman will always use the current default schema to generate this manifest. Future versions of Loftsman will likely allow a flag for generating from different schema versions_

Let's save the output of the `loftsman manifest create` command to a new manifest yaml file, and start to fill in some of the details:

```yaml
apiVersion: manifests/v1beta1
metadata:
  name: my-first-manifest
spec:
  sources:
    charts:
    - type: directory
      name: local
      location: ./charts
    - type: repo
      name: hashicorp
      location: https://helm.releases.hashicorp.com
  charts:
  - name: consul
    source: hashicorp
    version: 0.33.0
    namespace: default
  - name: victoria-metrics-cluster
    source: local
    version: 0.8.24
    namespace: default
```

We've updated the name to be `my-first-manifest`. Loftsman allows us to apply any number of manifests to a given cluster. It will track history and operations keying off this name. For example, two manifests with the same name cannot ship at the same time.

Next, we see that we've defined our manifest to ship our two charts, both `consul` and `victoria-metrics-cluster`, specifying the versions of the charts that we want to install or upgrade if the `ship` operation is to upgrade already-running workloads. Both of these charts will be deployed into the Kubernetes cluster `default` namespace.

Save this file to `manifest.yaml` in your `loftsman-workspace` directory, and let's ship it!

### Shipping your Manifest

We mentioned in the [requirements](#requirements) section that you needed a valid `kubeconfig` connected to your cluster so that Loftsman can actually perform operations against Kubernetes. Now's a good time to check that and that it's pointed to the right cluster. If you have `kubectl` installed and configured, it's pretty easy to do so:

```
$ kubectl config current-context
[should show the context for the cluster configured in kubeconfig]
$ kubectl get all -A
[to verify the things currently in your cluster, just some initial view into the cluster where we're about to ship]
```

Having verified we have a cluster where we want to ship our Loftsman manifest and related charts, we're ready to run our `loftsman ship` of the manifest we created.

_NOTE: that `loftsman` also accepts a `--kubeconfig` argument, which is the path to some `kubeconfig` file, and `--kube-context` which is the context to use within that file if a `kubeconfig` and context within other than the current system/user default is needed. See [the official doc on managing `kubeconfig` files for more info](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/)_

Here's the current state of our local workspace:

```
.
├── charts
│   └── victoria-metrics-cluster-0.8.24.tgz
└── manifest.yaml

1 directory, 3 files
```

So, let's run `loftsman ship` with our defined manifest:

```
$ loftsman ship --manifest-path ./manifest.yaml
2021-12-09T14:08:30-06:00 INF Initializing the connection to the Kubernetes cluster using KUBECONFIG (system default), and context (current-context) command=ship
2021-12-09T14:08:30-06:00 INF Initializing helm client object command=ship
         |\
         | \
         |  \
         |___\      Shipping your Helm workloads with Loftsman
       \--||___/
  ~~~~~~\_____/~~~~~~~
  
2021-12-09T14:08:31-06:00 INF Ensuring that the loftsman namespace exists command=ship
2021-12-09T14:08:31-06:00 INF Running a release for the provided manifest at manifest.yaml command=ship

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Releasing consul v0.33.0
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

2021-12-09T14:08:31-06:00 INF Running helm install/upgrade with arguments: upgrade --install consul https://helm.releases.hashicorp.com/consul-0.33.0.tgz --namespace default --create-namespace --set global.chart.name=consul --set global.chart.version=0.33.0 chart=consul command=ship namespace=default version=0.33.0
2021-12-09T14:08:37-06:00 INF Release "consul" has been upgraded. Happy Helming!
NAME: consul
LAST DEPLOYED: Thu Dec  9 14:08:37 2021
NAMESPACE: default
STATUS: deployed
REVISION: 3
NOTES:
Thank you for installing HashiCorp Consul!

Now that you have deployed Consul, you should look over the docs on using 
Consul with Kubernetes available here: 

https://www.consul.io/docs/platform/k8s/index.html


Your release is named consul.

To learn more about the release, run:

  $ helm status consul
  $ helm get all consul
 chart=consul command=ship namespace=default version=0.33.0

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Releasing victoria-metrics-cluster v0.8.24
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

2021-12-09T14:08:37-06:00 INF Running helm install/upgrade with arguments: upgrade --install victoria-metrics-cluster charts/victoria-metrics-cluster-0.8.24.tgz --namespace default --create-namespace --set global.chart.name=victoria-metrics-cluster --set global.chart.version=0.8.24 chart=victoria-metrics-cluster command=ship namespace=default version=0.8.24
2021-12-09T14:08:39-06:00 INF Release "victoria-metrics-cluster" has been upgraded. Happy Helming!
NAME: victoria-metrics-cluster
LAST DEPLOYED: Thu Dec  9 14:08:38 2021
NAMESPACE: default
STATUS: deployed
REVISION: 3
TEST SUITE: None
NOTES:
Write API:

The Victoria Metrics write api can be accessed via port 8480 on the following DNS name from within your cluster:
victoria-metrics-cluster-vminsert.default.svc.cluster.local

[redacted]

 chart=victoria-metrics-cluster command=ship namespace=default version=0.8.24
2021-12-09T14:08:39-06:00 INF Ship status: success. Recording status, manifest to configmap loftsman-my-first-manifest in namespace loftsman command=ship
2021-12-09T14:08:39-06:00 INF Recording log data to configmap loftsman-my-first-manifest-ship-log in namespace loftsman command=ship
```

Great! We've shipped our system which includes two charts with various workloads in this example. We have a result and record of the ship operation per our logs:

```
2021-12-09T14:08:39-06:00 INF Ship status: success. Recording status, manifest to configmap loftsman-my-first-manifest in namespace loftsman command=ship
2021-12-09T14:08:39-06:00 INF Recording log data to configmap loftsman-my-first-manifest-ship-log in namespace loftsman command=ship
```
 _NOTE: if any of our Helm charts had failed to install, we'd get a clear indication of that at the end of this log. Loftsman won't consider an install of one chart an overall failure, rather will aggregate these failures and report them at the end of the log._

 We also have a record of the manifest and it's logs now store in our cluster so we can reference it.

 ### Secure, remote Helm chart repos

 We just saw an example of shipping a manifest with some local charts. We could've also simply pointed Loftsman at a remote Helm charts repo as well, even if it's secured and needs credentialed access, and can use a manifest like:

 ```
 apiVersion: manifests/v1beta1
 metadata:
   name: my-alt-manifest
 spec:
   sources:
     charts:
     - type: repo
       name: myorg
       location: https://charts.myorg.com
   charts:
   - name: my-chart
     source: myorg
     version: 0.1.0
     namespace: default
 ```

## Next Steps in Working with Loftsman

_NOTE: v2.x of Loftsman, which will also include support for Loftsman running as an operator in the cluster and receiving applied manifests, will be able to deal with multiple chart repos at a time. In short, we're moving almost everything out of CLI args and going to let it be driven by manifest configuration._

Now that we've shipped our first manifest, we're in good shape. But there are other things that will help us use Loftsman well.

### Understanding Loftsman Logs and Records

We were able to see in the logs from our first `loftsman ship` operation:

```
2021-12-09T14:08:39-06:00 INF Ship status: success. Recording status, manifest to configmap loftsman-my-first-manifest in namespace loftsman command=ship
2021-12-09T14:08:39-06:00 INF Recording log data to configmap loftsman-my-first-manifest-ship-log in namespace loftsman command=ship
```

#### Ship result configmap
So, let's take a closer look at that Kubernetes-stored `ConfigMap` that contains the ship result:

```yaml
apiVersion: v1
data:
  manifest.yaml: |-
    apiVersion: manifests/v1beta1
    metadata:
      name: my-first-manifest
    spec:
      sources:
        charts:
        - type: directory
          name: local
          location: ./charts
        - type: repo
          name: hashicorp
          location: https://helm.releases.hashicorp.com
      charts:
      - name: consul
        source: hashicorp
        version: 0.33.0
        namespace: default
      - name: victoria-metrics-cluster
        source: local
        version: 0.8.24
        namespace: default
  status: success
kind: ConfigMap
metadata:
  annotations:
    loftsman.io/ship-log-configmap: loftsman-my-first-manifest-ship-log
  creationTimestamp: "2021-12-09T20:07:14Z"
  labels:
    app.kubernetes.io/managed-by: loftsman
  name: loftsman-my-first-manifest
  namespace: loftsman
  resourceVersion: "16711"
  uid: d3274d10-f5b2-4538-8f05-0c483fdb7157
```

Let's look at all the individual pieces:

* `metadata.name`: `loftsman-my-first-manifest`, a loftsman generated resource name based on the name of your manifest. Again, Loftsman considers manifests unique and connected across ship operations based on the name of the manifest itself
* `namespace`: `loftsman`, by default, loftsman will store everything it needs to in the `loftsman` namespace. You can control what namespace to use via the CLI `--loftsman-namespace` argument.
* `data."manifest.yaml"`: a record of the actual manifest shipped for this run
* `data.success`: whether or not the ship was successful or encountered failures

This `ConfigMap` will currently store the last ship data, think of it as state of a shipped manifest.

#### Ship log configmap

So, let's take a closer look at that Kubernetes-stored `ConfigMap` that contains the logs from the ship operation:
```yaml
apiVersion: v1
data:
  loftsman.log: |
    {"level":"info","command":"ship","time":"2021-12-09T14:08:30-06:00","message":"Initializing the connection to the Kubernetes cluster using KUBECONFIG (system default), and context (current-context)"}
    {"level":"info","command":"ship","time":"2021-12-09T14:08:30-06:00","message":"Initializing helm client object"}
    {"command":"ship","header":"Shipping your Helm workloads with Loftsman","time":"2021-12-09T14:08:31-06:00"}
    {"level":"info","command":"ship","time":"2021-12-09T14:08:31-06:00","message":"Ensuring that the loftsman namespace exists"}
    {"level":"info","command":"ship","time":"2021-12-09T14:08:31-06:00","message":"Running a release for the provided manifest at manifest.yaml"}
    {"command":"ship","sub-header":"Releasing consul v0.33.0","time":"2021-12-09T14:08:31-06:00"}
    {"level":"info","command":"ship","chart":"consul","version":"0.33.0","namespace":"default","time":"2021-12-09T14:08:31-06:00","message":"Running helm install/upgrade with arguments: upgrade --install consul https://helm.releases.hashicorp.com/consul-0.33.0.tgz --namespace default --create-namespace --set global.chart.name=consul --set global.chart.version=0.33.0"}
    {"level":"info","command":"ship","chart":"consul","version":"0.33.0","namespace":"default","time":"2021-12-09T14:08:37-06:00","message":"Release \"consul\" has been upgraded. Happy Helming!\nNAME: consul\nLAST DEPLOYED: Thu Dec  9 14:08:37 2021\nNAMESPACE: default\nSTATUS: deployed\nREVISION: 3\nNOTES:\nThank you for installing HashiCorp Consul!\n\nNow that you have deployed Consul, you should look over the docs on using \nConsul with Kubernetes available here: \n\nhttps://www.consul.io/docs/platform/k8s/index.html\n\n\nYour release is named consul.\n\nTo learn more about the release, run:\n\n  $ helm status consul\n  $ helm get all consul\n"}
    {"command":"ship","sub-header":"Releasing victoria-metrics-cluster v0.8.24","time":"2021-12-09T14:08:37-06:00"}
    {"level":"info","command":"ship","chart":"victoria-metrics-cluster","version":"0.8.24","namespace":"default","time":"2021-12-09T14:08:37-06:00","message":"Running helm install/upgrade with arguments: upgrade --install victoria-metrics-cluster charts/victoria-metrics-cluster-0.8.24.tgz --namespace default --create-namespace --set global.chart.name=victoria-metrics-cluster --set global.chart.version=0.8.24"}
    {"level":"info","command":"ship","chart":"victoria-metrics-cluster","version":"0.8.24","namespace":"default","time":"2021-12-09T14:08:39-06:00","message":"Release \"victoria-metrics-cluster\" has been upgraded. Happy Helming!\nNAME: victoria-metrics-cluster\nLAST DEPLOYED: Thu Dec  9 14:08:38 2021\nNAMESPACE: default\nSTATUS: deployed\nREVISION: 3\nTEST SUITE: None\nNOTES:\nWrite API:\n\nThe Victoria Metrics write api can be accessed via port 8480 on the following DNS name from within your cluster:\nvictoria-metrics-cluster-vminsert.default.svc.cluster.local\n\nGet the Victoria Metrics insert service URL by running these commands in the same shell:\n  export POD_NAME=$(kubectl get pods --namespace default -l \"app=vminsert\" -o jsonpath=\"{.items[0].metadata.name}\")\n  kubectl --namespace default port-forward $POD_NAME 8480\n\nYou need to update your prometheus configuration file and add next lines into it:\n\nprometheus.yml\n```yaml\nremote_write:\n  - url: \"http://<insert-service>/insert/0/prometheus/\"\n\n```\n\nfor e.g. inside the kubernetes cluster:\n```yaml\nremote_write:\n  - url: \"http://victoria-metrics-cluster-vminsert.default.svc.cluster.local:8480/insert/0/prometheus/\"\n\n```\nRead API:\n\nThe Victoria Metrics read api can be accessed via port 8481 on the following DNS name from within your cluster:\nvictoria-metrics-cluster-vmselect.default.svc.cluster.local\n\nGet the Victoria Metrics select service URL by running these commands in the same shell:\n  export POD_NAME=$(kubectl get pods --namespace default -l \"app=vmselect\" -o jsonpath=\"{.items[0].metadata.name}\")\n  kubectl --namespace default port-forward $POD_NAME 8481\n\nYou need to update specify select service URL in your Grafana:\n NOTE: you need to use Prometheus Data Source\n\nInput for URL field in Grafana\n\n```\nhttp://<select-service>/select/0/prometheus/\n```\n\nfor e.g. inside the kubernetes cluster:\n```\nhttp://victoria-metrics-cluster-vmselect.default.svc.cluster.local:8481/select/0/prometheus/\"\n```\n"}
    {"level":"info","command":"ship","time":"2021-12-09T14:08:39-06:00","message":"Ship status: success. Recording status, manifest to configmap loftsman-my-first-manifest in namespace loftsman"}
    {"level":"info","command":"ship","time":"2021-12-09T14:08:39-06:00","message":"Recording log data to configmap loftsman-my-first-manifest-ship-log in namespace loftsman"}
kind: ConfigMap
metadata:
  creationTimestamp: "2021-12-09T20:07:38Z"
  labels:
    app.kubernetes.io/managed-by: loftsman
  name: loftsman-my-first-manifest-ship-log
  namespace: loftsman
  resourceVersion: "16887"
  uid: f64b003c-db68-43f5-9d5f-22dc854f37a8
```

Let's look at all the individual pieces:
* `data."loftsman.log"`: is a record of the full log of the `loftsman ship` run, in JSON/machine-readable log format

### Identifying and Fixing Errors

This section is a work-in-progress, so bear with us as we build it out. In the meantime, these are the most helpful tips:

* Deal with `loftsman ship` CLI exit codes appropriately. It will fail with an exit code of 1 if any chart install or upgrade fails
* In many cases, dependencies across resources that are being installed can mean a chart and its resources fail to install b/c they're waiting on conditions like maybe readiness or liveness checks, Helm chart hooks, etc. We're currently exploring further ways of [integrating dependency awareness into Loftsman itself](https://github.com/Cray-HPE/loftsman/issues/5).
* The Loftsman log delivers helpful info, including Helm logs. Review Loftsman logs as thoroughly as possible, you'll usually find your answer, or at least a pointer that will lead you to your answer in the logs.

### Manifest Schema Versions

Loftsman will support multiple schema versions, and deprecate these versions in appropriate ways as we move forward. You can find source for all schema versions [here](../schemas/). We'll provide a bit more context about each version here though:

#### <a name="manifests-v1beta1"/> [`manifests/v1beta1`](../schemas/manifests/v1beta1)

* Our first official Loftsman manifest schema, in beta as we carefully determine what our next schema version should include
* See an [example manifest, with all available options filled in and commented](../schemas/manifests/v1beta1/examples/comprehensive.yaml)
