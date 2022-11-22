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

package dynamicmapper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/fake"
	core "k8s.io/client-go/testing"
)

const testingMapperRefreshInterval = 1 * time.Second

func setupMapper(t *testing.T, stopChan <-chan struct{}) (*RegeneratingDiscoveryRESTMapper, *fake.FakeDiscovery) {
	fakeDiscovery := &fake.FakeDiscovery{Fake: &core.Fake{}}
	mapper, err := NewRESTMapper(fakeDiscovery, testingMapperRefreshInterval)
	require.NoError(t, err, "constructing the rest mapper shouldn't have produced an error")

	fakeDiscovery.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Namespaced: true, Kind: "Pod"},
			},
		},
	}

	if stopChan != nil {
		mapper.RunUntil(stopChan)
	}

	return mapper, fakeDiscovery
}

func TestRegeneratingUpdatesMapper(t *testing.T) {
	mapper, fakeDiscovery := setupMapper(t, nil)
	require.NoError(t, mapper.RegenerateMappings(), "regenerating the mappings the first time should not have yielded an error")

	// add first, to ensure we don't update before regen
	fakeDiscovery.Resources[0].APIResources = append(fakeDiscovery.Resources[0].APIResources, metav1.APIResource{
		Name: "services", Namespaced: true, Kind: "Service",
	})
	fakeDiscovery.Resources = append(fakeDiscovery.Resources, &metav1.APIResourceList{
		GroupVersion: "wardle/v1alpha1",
		APIResources: []metav1.APIResource{
			{Name: "flunders", Namespaced: true, Kind: "Flunder"},
		},
	})

	// fetch before regen
	podsGVK, err := mapper.KindFor(schema.GroupVersionResource{Resource: "pods"})
	require.NoError(t, err, "should have been able to fetch the kind for 'pods' the first time")
	assert.Equal(t, schema.GroupVersionKind{Version: "v1", Kind: "Pod"}, podsGVK, "should have correctly fetched the kind for 'pods' the first time")
	_, err = mapper.KindFor(schema.GroupVersionResource{Resource: "services"})
	assert.Error(t, err, "should not have been able to fetch the kind for 'services' the first time")
	_, err = mapper.KindFor(schema.GroupVersionResource{Resource: "flunders", Group: "wardle"})
	assert.Error(t, err, "should not have been able to fetch the kind for 'flunders.wardle' the first time")

	// regen and check again
	require.NoError(t, mapper.RegenerateMappings(), "regenerating the mappings the second time should not have yielded an error")

	podsGVK, err = mapper.KindFor(schema.GroupVersionResource{Resource: "pods"})
	if assert.NoError(t, err, "should have been able to fetch the kind for 'pods' the second time") {
		assert.Equal(t, schema.GroupVersionKind{Version: "v1", Kind: "Pod"}, podsGVK, "should have correctly fetched the kind for 'pods' the first time")
	}
	servicesGVK, err := mapper.KindFor(schema.GroupVersionResource{Resource: "services"})
	if assert.NoError(t, err, "should have been able to fetch the kind for 'services' the second time") {
		assert.Equal(t, schema.GroupVersionKind{Version: "v1", Kind: "Service"}, servicesGVK, "should have correctly fetched the kind for 'services' the second time")
	}
	flundersGVK, err := mapper.KindFor(schema.GroupVersionResource{Resource: "flunders", Group: "wardle"})
	if assert.NoError(t, err, "should have been able to fetch the kind for 'flunders.wardle' the second time") {
		assert.Equal(t, schema.GroupVersionKind{Version: "v1alpha1", Kind: "Flunder", Group: "wardle"}, flundersGVK, "should have correctly fetched the kind for 'flunders.wardle' the second time")
	}
}
