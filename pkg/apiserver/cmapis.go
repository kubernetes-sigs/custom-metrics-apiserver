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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/metrics/pkg/apis/custom_metrics"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
	apiserver "sigs.k8s.io/custom-metrics-apiserver/pkg/registry/custom_metrics"
)

func (s *CustomMetricsAdapterServer) InstallCustomMetricsAPI() error {
	groupInfo := genericapiserver.NewDefaultAPIGroupInfo(custom_metrics.GroupName, Scheme, runtime.NewParameterCodec(Scheme), Codecs)
	custommetrics := newCustomMetrics(s.customMetricsProvider)
	customMetricsResources := map[string]rest.Storage{
		"custommetrics": custommetrics,
	}
	for _, mainGroupVer := range groupInfo.PrioritizedVersions {
		groupInfo.VersionedResourcesStorageMap[mainGroupVer.Version] = customMetricsResources
	}
	if err := s.GenericAPIServer.InstallAPIGroup(&groupInfo); err != nil {
		return err
	}
	return nil
}

func newCustomMetrics(p provider.CustomMetricsProvider) *apiserver.CustomMetrics {
	return &apiserver.CustomMetrics{CustomMetricsProvider: p}
}
