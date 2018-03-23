/*
Copyright 2018 The Kubernetes Authors.

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
	"k8s.io/apimachinery/pkg/apimachinery"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	genericapi "k8s.io/apiserver/pkg/endpoints"
	"k8s.io/apiserver/pkg/endpoints/discovery"
	genericapiserver "k8s.io/apiserver/pkg/server"

	specificapi "github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/apiserver/installer"
	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
	metricstorage "github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/registry/external_metrics"
	"k8s.io/metrics/pkg/apis/external_metrics"
)

// InstallExternalMetricsAPI registers the api server in Kube Aggregator
func (s *CustomMetricsAdapterServer) InstallExternalMetricsAPI() error {

	groupMeta := registry.GroupOrDie(external_metrics.GroupName)

	preferredVersionForDiscovery := metav1.GroupVersionForDiscovery{
		GroupVersion: groupMeta.GroupVersion.String(),
		Version:      groupMeta.GroupVersion.Version,
	}
	groupVersion := metav1.GroupVersionForDiscovery{
		GroupVersion: groupMeta.GroupVersion.String(),
		Version:      groupMeta.GroupVersion.Version,
	}
	apiGroup := metav1.APIGroup{
		Name:             groupMeta.GroupVersion.Group,
		Versions:         []metav1.GroupVersionForDiscovery{groupVersion},
		PreferredVersion: preferredVersionForDiscovery,
	}

	emAPI := s.emAPI(groupMeta, &groupMeta.GroupVersion)

	if err := emAPI.InstallREST(s.GenericAPIServer.Handler.GoRestfulContainer); err != nil {
		return err
	}

	s.GenericAPIServer.DiscoveryGroupManager.AddGroup(apiGroup)
	s.GenericAPIServer.Handler.GoRestfulContainer.Add(discovery.NewAPIGroupHandler(s.GenericAPIServer.Serializer, apiGroup, s.GenericAPIServer.RequestContextMapper()).WebService())

	return nil
}

func (s *CustomMetricsAdapterServer) emAPI(groupMeta *apimachinery.GroupMeta, groupVersion *schema.GroupVersion) *specificapi.MetricsAPIGroupVersion {
	resourceStorage := metricstorage.NewREST(s.externalMetricsProvider)

	return &specificapi.MetricsAPIGroupVersion{
		DynamicStorage: resourceStorage,
		APIGroupVersion: &genericapi.APIGroupVersion{
			Root:         genericapiserver.APIGroupPrefix,
			GroupVersion: *groupVersion,

			ParameterCodec:  metav1.ParameterCodec,
			Serializer:      Codecs,
			Creater:         Scheme,
			Convertor:       Scheme,
			UnsafeConvertor: runtime.UnsafeObjectConvertor(Scheme),
			Typer:           Scheme,
			Linker:          groupMeta.SelfLinker,
			Mapper:          groupMeta.RESTMapper,

			Context:                s.GenericAPIServer.RequestContextMapper(),
			OptionsExternalVersion: &schema.GroupVersion{Version: "v1"},
		},
		ResourceLister: provider.NewExternalMetricResourceLister(s.externalMetricsProvider),
		Handlers:       &specificapi.EMHandlers{},
	}
}
