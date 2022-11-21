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

package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// restMapper creates a RESTMapper with just the types we need for
// these tests.
func restMapper() apimeta.RESTMapper {
	mapper := apimeta.NewDefaultRESTMapper([]schema.GroupVersion{corev1.SchemeGroupVersion})
	mapper.Add(corev1.SchemeGroupVersion.WithKind("Pod"), apimeta.RESTScopeNamespace)

	return mapper
}

func TestNormalizeMetricInfoProducesSingularForm(t *testing.T) {
	pluralInfo := CustomMetricInfo{
		GroupResource: schema.GroupResource{Resource: "pods"},
		Namespaced:    true,
		Metric:        "cpu_usage",
	}

	_, singularRes, err := pluralInfo.Normalized(restMapper())
	require.NoError(t, err, "should not have returned an error while normalizing the plural MetricInfo")
	assert.Equal(t, "pod", singularRes, "should have produced a singular resource from the pural metric info")
}

func TestNormalizeMetricInfoDealsWithPluralization(t *testing.T) {
	singularInfo := CustomMetricInfo{
		GroupResource: schema.GroupResource{Resource: "pod"},
		Namespaced:    true,
		Metric:        "cpu_usage",
	}

	pluralInfo := CustomMetricInfo{
		GroupResource: schema.GroupResource{Resource: "pods"},
		Namespaced:    true,
		Metric:        "cpu_usage",
	}

	singularNormalized, singularRes, err := singularInfo.Normalized(restMapper())
	require.NoError(t, err, "should not have returned an error while normalizing the singular MetricInfo")
	pluralNormalized, pluralSingularRes, err := pluralInfo.Normalized(restMapper())
	require.NoError(t, err, "should not have returned an error while normalizing the plural MetricInfo")

	assert.Equal(t, singularRes, pluralSingularRes, "the plural and singular MetricInfo should have the same singularized resource")
	assert.Equal(t, singularNormalized, pluralNormalized, "the plural and singular MetricInfo should have the same normailzed form")
}
