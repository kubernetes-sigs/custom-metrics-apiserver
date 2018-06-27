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

### Example for developing your own Custom Metric Server API
This is an example of how to vendor this project and setup your own provider.  This example uses [glide](https://github.com/Masterminds/glide) but you can use any other [dependency management tool](https://github.com/golang/go/wiki/PackageManagementTools) of your choice.

#### Create your new project:

- `mkdir $GOPATH/src/github.com/your-company/my-custom-metric-server`
- `cd $GOPATH/src/github.com/your-company/my-custom-metric-server`

#### Vendor this repository:

- `glide create`
- `glide get github.com/kubernetes-incubator/custom-metrics-apiserver`

#### Create the entry point from the sample:

- `touch main.go` 

> See the [sample entry point](https://github.com/kubernetes-incubator/custom-metrics-apiserver/blob/master/sample-main.go) for a complete example

#### Add the server:

- `mkdir -p pkg/cmd/server`
- `touch pkg/cmd/server/start.go` 

> See the [sample server](https://github.com/kubernetes-incubator/custom-metrics-apiserver/blob/master/pkg/sample-cmd/server/start.go) for complete example

#### Add and implement [the custom metrics interface](https://github.com/kubernetes-incubator/custom-metrics-apiserver/blob/d8f23423aa1d0ff2bc9656da863d721725b3c68a/pkg/provider/interfaces.go#L84) for your provider:

- `mkdir -p pkg/cmd/provider`
- `touch pkg/cmd/server/provider.go` 

> See the [sample provider](https://github.com/kubernetes-incubator/custom-metrics-apiserver/blob/master/pkg/sample-cmd/provider/provider.go) for complete example

#### Build the project
After adding and implementing the files above build you project and run it.

## Development for boilerplate project

### Pre-reqs

- [glide](https://github.com/Masterminds/glide#install) to install dependencies before you can use this project.
- [Go](https://golang.org/doc/install) 1.8+ 
- [Mercurial](https://www.mercurial-scm.org/downloads) - one of dependencies requires hg
- [git](https://git-scm.com/downloads)

### Clone and Build boilerplate project

Clone this repository:
- `cd $GOPATH` 
- `go get github.com/kubernetes-incubator/custom-metrics-apiserver`
- `cd $GOPATH/src/github.com/kubernetes-incubator/custom-metrics-apiserver`

Fork project in GitHub and add your remote (**optional** if you want to contribute back on this project):
- [Fork in GitHub](https://help.github.com/articles/fork-a-repo/)
- `git remote add fork <fork-url>`

Build the project and run it:
- `make`
- `make run`

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
