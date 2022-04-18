/*
Copyright 2022 The Kubernetes Authors.

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
	"context"
	"fmt"
	"sort"

	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/metrics/pkg/apis/custom_metrics"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
)

type CustomMetrics struct {
	CustomMetricsProvider provider.CustomMetricsProvider
}

var _ rest.Storage = &CustomMetrics{}
var _ rest.KindProvider = &CustomMetrics{}
var _ rest.Getter = &CustomMetrics{}
var _ rest.Lister = &CustomMetrics{}
var _ rest.TableConvertor = &CustomMetrics{}
var _ rest.Scoper = &CustomMetrics{}

// Kind implements rest.KindProvider interface
func (s *CustomMetrics) Kind() string {
	return "CustomMetrics"
}

// New implements rest.Storage interface
func (c *CustomMetrics) New() runtime.Object {
	return &custom_metrics.MetricValue{}
}

// NewList implements rest.Lister interface
func (s *CustomMetrics) NewList() runtime.Object {
	return &custom_metrics.MetricValueList{}
}

// List implements rest.Lister interface
func (s *CustomMetrics) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	// populate the label selector, defaulting to all
	selector := labels.Everything()
	metricLabelSelector := labels.Everything()
	var kindSelector, nameSelector, nameSpaceSelector, metricsSelector string
	var kindFound, nameFound, nameSpaceFound, metricsFound bool
	if options != nil && options.LabelSelector != nil {
		kindSelector, kindFound = options.LabelSelector.RequiresExactMatch("kind")
		metricsSelector, metricsFound = options.LabelSelector.RequiresExactMatch("metrics")
		if !kindFound && !metricsFound {
			selector = options.LabelSelector
		}

	}

	metrics, err := s.handleMetricsList(ctx, selector, metricLabelSelector)
	if err != nil {
		return nil, err
	}
	if options != nil && options.FieldSelector != nil {
		nameSelector, nameFound = options.FieldSelector.RequiresExactMatch("metadata.name")
		nameSpaceSelector, nameSpaceFound = options.FieldSelector.RequiresExactMatch("metadata.namespace")
	}
	var items []custom_metrics.MetricValue
	if kindFound || nameFound || nameSpaceFound || metricsFound {
		for _, item := range metrics.Items {
			if kindFound && item.DescribedObject.Kind != kindSelector {
				continue
			}
			if nameFound && item.DescribedObject.Name != nameSelector {
				continue
			}
			if nameSpaceFound && item.DescribedObject.Namespace != nameSpaceSelector {
				continue
			}
			if metricsFound && item.Metric.Name != metricsSelector {
				continue
			}
			items = append(items, item)
		}
		if len(items) != 0 {
			metrics.Items = items
		} else {
			metrics = nil
		}

	}

	if metrics == nil {
		return nil, fmt.Errorf("resources not found")
	}
	return metrics, nil
}

func (s *CustomMetrics) handleMetricsList(ctx context.Context, selector labels.Selector, metricLabelSelector labels.Selector) (*custom_metrics.MetricValueList, error) {
	return s.CustomMetricsProvider.GetAllMetrics(ctx, selector, metricLabelSelector)

}

// Get implements rest.Getter interface
func (s *CustomMetrics) Get(ctx context.Context, name string, opts *metav1.GetOptions) (runtime.Object, error) {
	selector := labels.Everything()
	metricLabelSelector := labels.Everything()
	metrics, err := s.handleMetricsList(ctx, selector, metricLabelSelector)
	if err != nil {
		return nil, err
	}
	var items []custom_metrics.MetricValue
	for _, item := range metrics.Items {
		if item.Metric.Name != name {
			continue
		}
		items = append(items, item)
	}
	if len(items) != 0 {
		metrics.Items = items
	} else {
		metrics = nil
	}

	if metrics == nil {
		return nil, fmt.Errorf("resource not found")
	}
	return metrics, nil
}

// ConvertToTable implements rest.TableConvertor interface
func (s *CustomMetrics) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1beta1.Table, error) {
	var table metav1beta1.Table
	setTableColumnDefine(&table)
	resourceMapAll := make(map[custom_metrics.ObjectReference][]custom_metrics.MetricValue)
	kindMap := make(map[string][]custom_metrics.ObjectReference)
	switch t := object.(type) {
	case *custom_metrics.MetricValue:
		resourceMapAll[t.DescribedObject] = append(resourceMapAll[t.DescribedObject], *t)
		table.ResourceVersion = t.DescribedObject.ResourceVersion
		addCustomMetricsToTable(&table, true, resourceMapAll)
		sort.Slice(resourceMapAll[t.DescribedObject], func(i, j int) bool {
			return resourceMapAll[t.DescribedObject][i].Metric.Name < resourceMapAll[t.DescribedObject][j].Metric.Name
		})
	case *custom_metrics.MetricValueList:
		table.ResourceVersion = t.ResourceVersion
		table.Continue = t.Continue
		for _, item := range t.Items {
			resourceMapAll[item.DescribedObject] = append(resourceMapAll[item.DescribedObject], item)
		}
		for k := range resourceMapAll {
			kindMap[k.Kind] = append(kindMap[k.Kind], k)

		}
		var kind []string
		for k := range kindMap {
			kind = append(kind, k)
		}
		sort.Slice(kind, func(i, j int) bool {
			return kind[i] < kind[j]
		})
		for _, k := range kind {
			v := kindMap[k]
			if len(v) != 0 {
				var nameSpaced bool

				if v[0].Namespace != "" {
					nameSpaced = true
				}
				resourceMap := make(map[custom_metrics.ObjectReference][]custom_metrics.MetricValue)
				for _, r := range v {
					resourceMap[r] = resourceMapAll[r]
				}
				addCustomMetricsToTable(&table, nameSpaced, resourceMap)
			}
		}
	default:
	}
	return &table, nil
}

// NamespaceScoped implements rest.Scoper interface
func (c *CustomMetrics) NamespaceScoped() bool {
	return false
}
