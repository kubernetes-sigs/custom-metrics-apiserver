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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/pkg/api"

	// install in order to make the types available for lookup
	_ "k8s.io/client-go/pkg/api/install"
)

func TestNormalizeMetricInfoProducesSingularForm(t *testing.T) {
	pluralInfo := MetricInfo{
		GroupResource: schema.GroupResource{Resource: "pods"},
		Namespaced: true,
		Metric: "cpu_usage",
	}

	_, singularRes, err := pluralInfo.Normalized(api.Registry.RESTMapper())
	require.NoError(t, err, "should not have returned an error while normalizing the plural MetricInfo")
	assert.Equal(t, "pod", singularRes, "should have produced a singular resource from the pural metric info")
}

func TestNormalizeMetricInfoDealsWithPluralization(t *testing.T) {
	singularInfo := MetricInfo{
		GroupResource: schema.GroupResource{Resource: "pod"},
		Namespaced: true,
		Metric: "cpu_usage",
	}

	pluralInfo := MetricInfo{
		GroupResource: schema.GroupResource{Resource: "pods"},
		Namespaced: true,
		Metric: "cpu_usage",
	}

	singularNormalized, singularRes, err := singularInfo.Normalized(api.Registry.RESTMapper())
	require.NoError(t, err, "should not have returned an error while normalizing the singular MetricInfo")
	pluralNormalized, pluralSingularRes, err := pluralInfo.Normalized(api.Registry.RESTMapper())
	require.NoError(t, err, "should not have returned an error while normalizing the plural MetricInfo")

	assert.Equal(t, singularRes, pluralSingularRes, "the plural and singular MetricInfo should have the same singularized resource")
	assert.Equal(t, singularNormalized, pluralNormalized, "the plural and singular MetricInfo should have the same normailzed form")
}
