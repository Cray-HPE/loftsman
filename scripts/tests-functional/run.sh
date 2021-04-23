#!/bin/bash

set -e

divider="------------------------------------------------------------------------------------------"
kind_cluster_name="loftsman-functional-tests"
consul_chart_version="0.31.1"
chartmuseum_container_name="${kind_cluster_name}-chartmuseum"

this_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $this_dir/../../

if ! command -v helm &>/dev/null; then
  echo "The helm binary needs to be installed to run functional tests"
  exit 1
fi
if ! command -v kind &>/dev/null; then
  echo "Kubernetes in Docker (KIND) is required to run functional tests (https://kind.sigs.k8s.io/)"
  exit 1
fi

if kind get clusters | grep "^$kind_cluster_name$" &>/dev/null; then
  kind delete cluster --name $kind_cluster_name
fi
rm $this_dir/charts/*.tgz &>/dev/null || true
rm $this_dir/charts/*.yaml &>/dev/null || true
docker rm -f $chartmuseum_container_name &>/dev/null || true

revert_context=""
if kubectl config current-context &>/dev/null; then
  revert_context=$(kubectl config current-context)
fi

################################# Manifest chart sources ######################################
echo $divider
echo "Manifest chart sources"
kind create cluster --name $kind_cluster_name
kubectl config use-context kind-${kind_cluster_name}

# pull some test charts to include the local directory chart source in tests
helm fetch https://helm.releases.hashicorp.com/consul-${consul_chart_version}.tgz -d $this_dir/charts/

go run . ship --manifest-path $this_dir/test-manifests/chart-sources.yaml

kind delete cluster --name $kind_cluster_name
rm $this_dir/charts/*.tgz

################################# Pre 1.1.0, CLI args ######################################
echo $divider
echo "Pre 1.1.0, CLI args"
kind create cluster --name $kind_cluster_name
kubectl config use-context kind-${kind_cluster_name}

helm fetch https://helm.releases.hashicorp.com/consul-${consul_chart_version}.tgz -d $this_dir/charts/
helm fetch https://victoriametrics.github.io/helm-charts/packages/victoria-metrics-cluster-0.8.24.tgz -d $this_dir/charts/

go run . ship --charts-path $this_dir/charts --manifest-path $this_dir/test-manifests/pre-1.1.0.yaml

kind delete cluster --name $kind_cluster_name

################################# Chart source repo with creds ######################################
echo $divider
echo "Manifest chart source repo with creds"
kind create cluster --name $kind_cluster_name
kubectl config use-context kind-${kind_cluster_name}

# run a chart museum container with auth set up
docker run --rm -d --name $chartmuseum_container_name \
  -p 8080:8080 \
  -e STORAGE=local \
  -e STORAGE_LOCAL_ROOTDIR=/charts \
  -v $this_dir/charts:/charts \
  chartmuseum/chartmuseum:latest \
  --basic-auth-user=user \
  --basic-auth-pass=pass

kubectl create secret generic repo-creds --from-literal=username=user --from-literal=password=pass

go run . ship --manifest-path $this_dir/test-manifests/chart-source-repo-creds.yaml

docker rm -f $chartmuseum_container_name &>/dev/null || true
kind delete cluster --name $kind_cluster_name
rm $this_dir/charts/*.tgz

if [ ! -z "$revert_context" ]; then
  kubectl config use-context $revert_context
fi
