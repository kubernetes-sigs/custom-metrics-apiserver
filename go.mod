module sigs.k8s.io/custom-metrics-apiserver

go 1.16

require (
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/emicklei/go-restful-swagger12 v0.0.0-20201014110547-68ccff494617
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/googleapis/gnostic v0.5.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.23.3
	k8s.io/apimachinery v0.23.3
	k8s.io/apiserver v0.23.3
	k8s.io/client-go v0.23.3
	k8s.io/component-base v0.23.3
	k8s.io/klog/v2 v2.40.1
	k8s.io/kube-openapi v0.0.0-20220124234850-424119656bbf
	k8s.io/metrics v0.23.3
	k8s.io/utils v0.0.0-20211208161948-7d6a63dca704
)
