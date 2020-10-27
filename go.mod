module github.com/kubernetes-sigs/custom-metrics-apiserver

go 1.15

require (
	github.com/emicklei/go-restful v2.14.3+incompatible
	github.com/emicklei/go-restful-swagger12 v0.0.0-20201014110547-68ccff494617
	github.com/go-openapi/spec v0.19.3
	github.com/googleapis/gnostic v0.4.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.19.3
	k8s.io/apimachinery v0.19.3
	k8s.io/apiserver v0.19.3
	k8s.io/client-go v0.19.3
	k8s.io/component-base v0.19.3
	k8s.io/klog/v2 v2.3.0
	k8s.io/kube-openapi v0.0.0-20200805222855-6aeccd4b50c6
	k8s.io/metrics v0.19.3
	k8s.io/utils v0.0.0-20200729134348-d5654de09c73
)
