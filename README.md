# Custom Metrics Adapter Server Boilerplate

## Purpose

This repository contains boilerplate code for setting up an implementation
of the custom metrics API (https://github.com/kubernetes/metrics).

It includes the necessary boilerplate for setting up an implementation
(generic API server setup, registration of resources, etc), plus a sample
implementation backed by fake data.

## How to use this repository

In order to use this repository, you should vendor this repository at
`github.com/kubernetes-incubator/custom-metrics-apiserver`, and implement the
`"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider".CustomMetricsProvider`
interface.  You can then pass this to the main setup functions.

The `pkg/cmd` package contains the building blocks of the actual API
server setup.  You'll most likely want to wrap the existing options and
flags setup to add your own flags for configuring your provider.

A sample implementation of this can be found in the file `sample-main.go`
and `pkg/sample-cmd` directory.  You'll want to have the equivalent files
in your project.

### Building your own Custom Metric Server 

See [getting-started.md](docs/getting-started.md) for a walk through on creating your own custom metric server api.

## Development for boilerplate project

### Pre-reqs

- [glide](https://github.com/Masterminds/glide#install) to install dependencies before you can use this project.
- [Go](https://golang.org/doc/install) same version of [Go as Kubernetes](https://github.com/kubernetes/community/blob/master/contributors/devel/development.md#go)
- [Mercurial](https://www.mercurial-scm.org/downloads) - one of dependencies requires hg
- [git](https://git-scm.com/downloads)

### Clone and Build boilerplate project

There is a sample adapter in this repository that can be used for testing changes to the repository, and also acts as an example implementations.

To build it:

```bash
export REGISTRY=<your registory name>
make sample-container
kubectl create -f sample-deploy/manifests
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
