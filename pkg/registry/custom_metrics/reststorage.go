/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package apiserver

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/metrics/pkg/apis/custom_metrics"
	"k8s.io/custom-metrics-boilerplate/pkg/provider"
)

type REST struct {
	cmProvider provider.CustomMetricsProvider
}

var _ rest.KindProvider = &REST{}
var _ rest.Storage = &REST{}
var _ rest.GetterWithOptions = &REST{}

func NewREST(cmProvider provider.CustomMetricsProvider) *REST {
	return &REST{
		cmProvider: cmProvider,
	}
}

// Implement Storage

func (r *REST) New() runtime.Object {
	return &custom_metrics.MetricValueList{}
}

// Implement KindProvider

func (r *REST) Kind() string {
	return "MetricValueList"
}

func (r *REST) Get(ctx genericapirequest.Context, name string, options runtime.Object) (runtime.Object, error) {
	var err error
	selector := labels.Everything()
	if options != nil {
		listOpts := options.(*metav1.ListOptions)
		if listOpts.LabelSelector != "" {
			selector, err = labels.Parse(listOpts.LabelSelector)
			if err != nil {
				return nil, err
			}
		}
	}

	namespace := genericapirequest.NamespaceValue(ctx)

	resourceRaw, metricName, ok := genericapirequest.ResourceInformationFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("unable to get resource and metric name from request")
	}

	groupResource := schema.ParseGroupResource(resourceRaw)

	if namespace == "" {
		if name == "*" {
			return r.cmProvider.GetRootScopedMetricBySelector(groupResource, selector, metricName)
		}

		return r.cmProvider.GetRootScopedMetricByName(groupResource, name, metricName)
	}

	if name == "*" {
		return r.cmProvider.GetNamespacedMetricBySelector(groupResource, namespace, selector, metricName)
	}

	return r.cmProvider.GetNamespacedMetricByName(groupResource, namespace, name, metricName)
}

func (r *REST) NewGetOptions() (runtime.Object, bool, string) {
	return &metav1.ListOptions{}, false, ""
}
