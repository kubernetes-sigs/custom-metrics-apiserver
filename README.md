# Custom Metrics Adapter Server Boilerplate

## Purpose

This repository contains boilerplate code for setting up an implementation
of the custom metrics API (https://github.com/kubernetes/metrics).

It includes the necessary boilerplate for setting up an implementation
(generic API server setup, registration of resources, etc), plus a sample
implementation backed by fake data.

## How to use this repository

This repository is designed to be used as a library. First, implement one
or more of the metrics provider interfaces in `pkg/provider` (for example,
`CustomMetricsProvider`), depending on which APIs you want to support.

Then, use the `AdapterBase` in `pkg/cmd` to initialize the necessary flags
and set up the API server, passing in your providers.

More information can be found in the [getting started
guide](/docs/getting-started.md), and a sample implementation can be found
in the [sample directory](/sample).

It is *strongly* suggested that you make use of the dependency versions
listed in [glide.yaml](/glide.yaml), as mismatched versions of Kubernetes
dependencies can lead to build issues.

## Development for boilerplate project

### Pre-reqs

- [glide](https://github.com/Masterminds/glide#install) to install dependencies before you can use this project.
- [Go](https://golang.org/doc/install) same version of [Go as Kubernetes](https://github.com/kubernetes/community/blob/master/contributors/devel/development.md#go)
- [Mercurial](https://www.mercurial-scm.org/downloads) - one of dependencies requires hg
- [git](https://git-scm.com/downloads)

### Clone and Build boilerplate project

There is a sample adapter in this repository that can be used for testing
changes to the repository, and also acts as an example implementations.

To build and deploy it:

```bash
# build the sample container as $REGISTRY/k8s-custom-metric-adapter-sample
export REGISTRY=<some-prefix>
make sample-container

# push the container up to a registry (optional if your cluster is local)
docker push $REGISTRY/k8s-custom-metric-adapter-sample

# launch the adapter using the sample deployment files
kubectl create namespace custom-metrics
kubectl apply -f sample-deploy/manifests
```

After the deployment you can query the sample adapter with:

```
# you can pipe to `jq .` to pretty-print the output, if it's installed
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1"
```

## Compatibility

The APIs in this repository follow the standard guarantees for Kubernetes
APIs, and will follow Kubernetes releases.

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community
page](http://kubernetes.io/community/).

You can reach the maintainers of this repository at:

- Slack: #sig-instrumentation (on https://kubernetes.slack.com -- get an
  invite at slack.kubernetes.io)
- Mailing List:
  https://groups.google.com/forum/#!forum/kubernetes-sig-instrumentation

### Code of Conduct

Participation in the Kubernetes community is governed by the [Kubernetes
Code of Conduct](code-of-conduct.md).

### Contribution Guidelines

See [CONTRIBUTING.md](CONTRIBUTING.md) for more information.
