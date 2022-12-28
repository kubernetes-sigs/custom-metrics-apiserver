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

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/kube-openapi/pkg/builder"

	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
	sampleprovider "sigs.k8s.io/custom-metrics-apiserver/test-adapter/provider"
)

func TestDefaultOpenAPIConfig(t *testing.T) {
	t.Run("no metric", func(t *testing.T) {
		adapter := &AdapterBase{}
		config := adapter.defaultOpenAPIConfig()

		_, err1 := builder.BuildOpenAPIDefinitionsForResources(config, "k8s.io/metrics/pkg/apis/custom_metrics/v1beta2.MetricValue")
		// Should err, because no provider is installed
		assert.Error(t, err1)
		_, err2 := builder.BuildOpenAPIDefinitionsForResources(config, "k8s.io/metrics/pkg/apis/external_metrics/v1beta1.ExternalMetricValue")
		assert.Error(t, err2)
	})

	t.Run("custom and external metrics", func(t *testing.T) {
		adapter := &AdapterBase{}

		prov := newFakeProvider()
		adapter.WithCustomMetrics(prov)
		adapter.WithExternalMetrics(prov)

		config := adapter.defaultOpenAPIConfig()

		_, err1 := builder.BuildOpenAPIDefinitionsForResources(config, "k8s.io/metrics/pkg/apis/custom_metrics/v1beta2.MetricValue")
		// Should NOT err
		assert.NoError(t, err1)
		_, err2 := builder.BuildOpenAPIDefinitionsForResources(config, "k8s.io/metrics/pkg/apis/external_metrics/v1beta1.ExternalMetricValue")
		assert.NoError(t, err2)
	})
}

func newFakeProvider() provider.MetricsProvider {
	dynClient := fake.NewSimpleDynamicClient(runtime.NewScheme())
	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{})
	prov, _ := sampleprovider.NewFakeProvider(dynClient, mapper)
	return prov
}
