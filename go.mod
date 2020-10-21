module github.com/kubernetes-sigs/custom-metrics-apiserver

go 1.13

require (
	github.com/emicklei/go-restful v2.9.5+incompatible
	github.com/emicklei/go-restful-swagger12 v0.0.0-20170208215640-dcef7f557305
	github.com/go-openapi/spec v0.19.3
	github.com/googleapis/gnostic v0.1.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/apiserver v0.18.2
	k8s.io/client-go v0.18.2
	k8s.io/component-base v0.18.2
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c
	k8s.io/metrics v0.18.2
	k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89
)
