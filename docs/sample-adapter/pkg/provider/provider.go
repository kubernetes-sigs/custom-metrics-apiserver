// Copyright 2022 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"
	"time"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/metrics/pkg/apis/custom_metrics"

	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider/defaults"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider/helpers"
)

type yourProvider struct {
	defaults.DefaultCustomMetricsProvider
	defaults.DefaultExternalMetricsProvider

	client dynamic.Interface
	mapper apimeta.RESTMapper

	// just increment values when they're requested
	values map[provider.CustomMetricInfo]int64
}

func NewProvider(client dynamic.Interface, mapper apimeta.RESTMapper) provider.CustomMetricsProvider {
	return &yourProvider{
		client: client,
		mapper: mapper,
		values: make(map[provider.CustomMetricInfo]int64),
	}
}

// valueFor fetches a value from the fake list and increments it.
func (p *yourProvider) valueFor(info provider.CustomMetricInfo) (int64, error) {
	// normalize the value so that you treat plural resources and singular
	// resources the same (e.g. pods vs pod)
	info, _, err := info.Normalized(p.mapper)
	if err != nil {
		return 0, err
	}

	value := p.values[info]
	value++
	p.values[info] = value

	return value, nil
}

// metricFor constructs a result for a single metric value.
func (p *yourProvider) metricFor(value int64, name types.NamespacedName, info provider.CustomMetricInfo) (*custom_metrics.MetricValue, error) {
	// construct a reference referring to the described object
	objRef, err := helpers.ReferenceFor(p.mapper, name, info)
	if err != nil {
		return nil, err
	}

	return &custom_metrics.MetricValue{
		DescribedObject: objRef,
		Metric: custom_metrics.MetricIdentifier{
			Name: info.Metric,
		},
		// you'll want to use the actual timestamp in a real adapter
		Timestamp: metav1.Time{Time: time.Now()},
		Value:     *resource.NewMilliQuantity(value*100, resource.DecimalSI),
	}, nil
}

func (p *yourProvider) GetMetricByName(_ context.Context, name types.NamespacedName, info provider.CustomMetricInfo, _ labels.Selector) (*custom_metrics.MetricValue, error) {
	value, err := p.valueFor(info)
	if err != nil {
		return nil, err
	}
	return p.metricFor(value, name, info)
}

func (p *yourProvider) GetMetricBySelector(_ context.Context, namespace string, selector labels.Selector, info provider.CustomMetricInfo, _ labels.Selector) (*custom_metrics.MetricValueList, error) {
	totalValue, err := p.valueFor(info)
	if err != nil {
		return nil, err
	}

	names, err := helpers.ListObjectNames(p.mapper, p.client, namespace, selector, info)
	if err != nil {
		return nil, err
	}

	res := make([]custom_metrics.MetricValue, len(names))
	for i, name := range names {
		// in a real adapter, you might want to consider pre-computing the
		// object reference created in metricFor, instead of recomputing it
		// for each object.
		value, err := p.metricFor(100*totalValue/int64(len(res)), types.NamespacedName{Namespace: namespace, Name: name}, info)
		if err != nil {
			return nil, err
		}
		res[i] = *value
	}

	return &custom_metrics.MetricValueList{
		Items: res,
	}, nil
}
