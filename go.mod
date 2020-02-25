module github.com/kubernetes-incubator/custom-metrics-apiserver

go 1.13.6

require (
	github.com/emicklei/go-restful v2.9.5+incompatible
	github.com/emicklei/go-restful-swagger12 v0.0.0-20170208215640-dcef7f557305
	github.com/googleapis/gnostic v0.0.0-20170729233727-0c5108395e2d
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/apiserver v0.17.3
	k8s.io/client-go v0.17.3
	k8s.io/component-base v0.17.3
	k8s.io/klog v1.0.0
	k8s.io/metrics v0.17.3
	k8s.io/utils v0.0.0-20191114184206-e782cd3c129f
)
