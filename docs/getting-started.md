# Getting started with developing your own Custom Metric Server API
This is an example of how to vendor this project and setup your own provider.  This example uses [glide](https://github.com/Masterminds/glide) but you can use any other [dependency management tool](https://github.com/golang/go/wiki/PackageManagementTools) of your choice.

## Create your new project:

- `mkdir $GOPATH/src/github.com/your-company/my-custom-metric-server`
- `cd $GOPATH/src/github.com/your-company/my-custom-metric-server`

## Vendor this repository:

- `glide create`
- `glide get github.com/kubernetes-incubator/custom-metrics-apiserver`

## Create the entry point and add the server:

- `mkdir -p cmd/server`
- `touch cmd/main.go`
- `touch cmd/server/start.go`

> See the [sample server](https://github.com/kubernetes-incubator/custom-metrics-apiserver/blob/master/pkg/sample-cmd/server/start.go) and [sample entry point](https://github.com/kubernetes-incubator/custom-metrics-apiserver/blob/master/sample-main.go) for an example implementation

## Add and implement [the custom metrics interface](https://github.com/kubernetes-incubator/custom-metrics-apiserver/blob/d8f23423aa1d0ff2bc9656da863d721725b3c68a/pkg/provider/interfaces.go#L84) for your provider:

- `mkdir -p pkg/provider`
- `touch pkg/provider/provider.go`

> See the [sample provider](https://github.com/kubernetes-incubator/custom-metrics-apiserver/blob/master/pkg/sample-cmd/provider/provider.go) for complete example

## Build the project

After adding and implementing the files above build you project and run it.
