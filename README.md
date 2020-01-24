# Custom Metrics Adapter Server Boilerplate

## Purpose

This repository contains boilerplate code for setting up an implementation
of the custom metrics API (https://github.com/kubernetes/metrics).

It includes the necessary boilerplate for setting up an implementation
(generic API server setup, registration of resources, etc), plus an
implementation for testing that allows setting metric values over HTTP.

## How to use this repository

This repository is designed to be used as a library. First, implement one
or more of the metrics provider interfaces in `pkg/provider` (for example,
`CustomMetricsProvider`), depending on which APIs you want to support.

Then, use the `AdapterBase` in `pkg/cmd` to initialize the necessary flags
and set up the API server, passing in your providers.

More information can be found in the [getting started
guide](/docs/getting-started.md), and the testing implementation can be
found in the [test-adapter directory](/test-adapter).

## Development for boilerplate project

### Pre-reqs

- [Go](https://golang.org/doc/install) same version of [Go as Kubernetes](https://github.com/kubernetes/community/blob/master/contributors/devel/development.md#go)
- [git](https://git-scm.com/downloads)

### Clone and Build the Testing Adapter

There is a test adapter in this repository that can be used for testing
changes to the repository, as a mock implementation of the APIs for
automated unit tests, and also as an example implementation.

Note that this adapter *should not* be used for production.  It's for
writing automated e2e tests and serving as a sample only.

To build and deploy it:

```bash
# build the test-adapter container as $REGISTRY/k8s-test-metrics-adapter
export REGISTRY=<some-prefix>
make test-adapter-container

# push the container up to a registry (optional if your cluster is local)
docker push $REGISTRY/k8s-test-metrics-adapter

# launch the adapter using the test adapter deployment files
kubectl apply -f test-adapter-deploy/testing-adapter.yaml
```

After the deployment you can set new metrics on the adapter using
query the testing adapter with:

```bash
# set up a proxy to the api server so we can access write endpoints
# of the testing adapter directly
kubectl proxy &
# write a sample metric -- the write paths match the same URL structure
# as the read paths, but at the /write-metrics base path.
# data needs to be in json, so we also need to set the content-type header
curl -XPOST -H 'Content-Type: application/json' http://localhost:8080/api/v1/namespaces/custom-metrics/services/custom-metrics-apiserver:http/proxy/write-metrics/namespaces/default/services/kubernetes/test-metric --data-raw '"300m"'
```

```
# you can pipe to `jq .` to pretty-print the output, if it's installed
# (otherwise, it's not necessary)
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1" | jq .
```

If you wanted to target a simple nginx-deployment and then use this as an HPA scaler metric, something like this would work following the previous curl command:
```
apiVersion: autoscaling/v2beta2
kind: HorizontalPodAutoscaler
metadata:
  name: nginx-deployment
  namespace: default
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: nginx-deployment
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: Object
    object:
      metric:
        name: test-metric
      describedObject:
        apiVersion: v1
        kind: Service
        name: kubernetes
      target:
        type: Value
        value: 100
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
