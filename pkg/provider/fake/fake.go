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

// Package fake provides a fake implementation of metrics providers.
package fake

import (
	"context"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/metrics/pkg/apis/custom_metrics"
	"k8s.io/metrics/pkg/apis/external_metrics"

	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider/defaults"
)

type fakeProvider struct {
	defaults.DefaultCustomMetricsProvider
	defaults.DefaultExternalMetricsProvider
}

func (*fakeProvider) GetMetricByName(_ context.Context, _ types.NamespacedName, _ provider.CustomMetricInfo, _ labels.Selector) (*custom_metrics.MetricValue, error) {
	return &custom_metrics.MetricValue{}, nil
}

func (*fakeProvider) GetMetricBySelector(_ context.Context, _ string, _ labels.Selector, _ provider.CustomMetricInfo, _ labels.Selector) (*custom_metrics.MetricValueList, error) {
	return &custom_metrics.MetricValueList{}, nil
}

func (*fakeProvider) GetExternalMetric(_ context.Context, _ string, _ labels.Selector, _ provider.ExternalMetricInfo) (*external_metrics.ExternalMetricValueList, error) {
	return &external_metrics.ExternalMetricValueList{}, nil
}

// NewProvider creates a fake implementation of MetricsProvider.
func NewProvider() provider.MetricsProvider {
	return &fakeProvider{}
}
