# Custom Metrics Adapter Server Boilerplate

[![Go Reference](https://pkg.go.dev/badge/sigs.k8s.io/custom-metrics-apiserver.svg)](https://pkg.go.dev/sigs.k8s.io/custom-metrics-apiserver)

## Purpose

This repository contains boilerplate code for setting up an implementation
of the [Metrics APIs](https://github.com/kubernetes/metrics):

- Custom metrics (`k8s.io/metrics/pkg/apis/custom_metrics`)
- External metrics (`k8s.io/metrics/pkg/apis/external_metrics`)

It includes the necessary boilerplate for setting up an implementation
(generic API server setup, registration of resources, etc), plus an
implementation for testing that allows setting custom metric values over HTTP.

## How to use this repository

This repository is designed to be used as a library. First, implement one
or more of the metrics provider interfaces in `pkg/provider` (for example,
`CustomMetricsProvider`), depending on which APIs you want to support.

Then, use the `AdapterBase` in `pkg/cmd` to initialize the necessary flags
and set up the API server, passing in your providers.

More information can be found in the [getting started
guide](/docs/getting-started.md), and the testing implementation can be
found in the [test-adapter directory](/test-adapter).

### Prerequisites

[Go](https://go.dev/doc/install): this library requires the same version of
[Go as Kubernetes](https://git.k8s.io/community/contributors/devel/development.md#go).

## Test Adapter

There is a test adapter in this repository that can be used for testing
changes to the repository, as a mock implementation of the APIs for
automated unit tests, and also as an example implementation.

Note that this adapter **should not** be used for production. It's for
writing automated e2e tests and serving as a sample only.

To build and deploy it:

```bash
# build the test-adapter container as $REGISTRY/k8s-test-metrics-adapter-amd64
export REGISTRY=<some-prefix>
make test-adapter-container

# push the container up to a registry (optional if your cluster is local)
docker push $REGISTRY/k8s-test-metrics-adapter-amd64

# launch the adapter using the test adapter deployment manifest
kubectl apply -f test-adapter-deploy/testing-adapter.yaml
```

When the deployment is ready, you can define new metrics on the test adapter
by querying the write endpoint:

```bash
# set up a proxy to the API server so we can access write endpoints
# of the testing adapter directly
kubectl proxy &
# write a sample metric -- the write paths match the same URL structure
# as the read paths, but at the /write-metrics base path.
# data needs to be in json, so we also need to set the content-type header
curl -X POST \
  -H 'Content-Type: application/json' \
  http://localhost:8001/api/v1/namespaces/custom-metrics/services/custom-metrics-apiserver:http/proxy/write-metrics/namespaces/default/services/kubernetes/test-metric \
  --data-raw '"300m"'
```

```bash
# you can pipe to `jq .` to pretty-print the output, if it's installed
# (otherwise, it's not necessary)
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta2" | jq .
# fetching certain custom metrics of namespaced resources
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta2/namespaces/default/services/kubernetes/test-metric" | jq .
```

If you wanted to target a simple nginx-deployment and then use this as an HPA scaler metric, something like this would work following the previous curl command:

```yaml
apiVersion: autoscaling/v2
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
        value: 300m
```

You can also query the external metrics:

```bash
kubectl get --raw "/apis/external.metrics.k8s.io/v1beta1" | jq .
# fetching certain custom metrics of namespaced resources
kubectl get --raw "/apis/external.metrics.k8s.io/v1beta1/namespaces/default/my-external-metric" | jq .
```

## Compatibility

The APIs in this repository follow the standard guarantees for Kubernetes
APIs, and will follow Kubernetes releases.

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the
[community page](https://kubernetes.io/community/).

You can reach the maintainers of this repository at:

- [Slack](https://slack.k8s.io/): channel `#sig-instrumentation`
- [Mailing List](https://groups.google.com/g/kubernetes-sig-instrumentation)

### Code of Conduct

Participation in the Kubernetes community is governed by the [Kubernetes
Code of Conduct](code-of-conduct.md).

### Contribution Guidelines

See [CONTRIBUTING.md](CONTRIBUTING.md) for more information.
