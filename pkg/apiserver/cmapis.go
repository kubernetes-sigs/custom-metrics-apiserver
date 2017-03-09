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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/apimachinery"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericapi "k8s.io/apiserver/pkg/endpoints"

	"k8s.io/metrics/pkg/apis/custom_metrics"
	metricstorage "k8s.io/custom-metrics-boilerplate/pkg/registry/custom_metrics"
	"k8s.io/custom-metrics-boilerplate/pkg/provider"
)

func (s *CustomMetricsAdapterServer) InstallCustomMetricsAPI() error {

	groupMeta := registry.GroupOrDie(custom_metrics.GroupName)

	preferredVersionForDiscovery := metav1.GroupVersionForDiscovery{
		GroupVersion: groupMeta.GroupVersion.String(),
		Version:      groupMeta.GroupVersion.Version,
	}
    groupVersion := metav1.GroupVersionForDiscovery{
        GroupVersion: groupMeta.GroupVersion.String(),
        Version:      groupMeta.GroupVersion.Version,
    }
	apiGroup := metav1.APIGroup{
		Name:              groupMeta.GroupVersion.String(),
		Versions:          []metav1.GroupVersionForDiscovery{groupVersion},
		PreferredVersion:  preferredVersionForDiscovery,
	}

	cmAPI := s.cmAPI(groupMeta, &groupMeta.GroupVersion)

	if err := cmAPI.InstallREST(s.GenericAPIServer.HandlerContainer.Container); err != nil {
		return err
	}

	path := genericapiserver.APIGroupPrefix+"/"+groupMeta.GroupVersion.Group
	s.GenericAPIServer.HandlerContainer.Add(genericapi.NewGroupWebService(s.GenericAPIServer.Serializer, path, apiGroup))

	return nil
}
func (s *CustomMetricsAdapterServer) cmAPI(groupMeta *apimachinery.GroupMeta, groupVersion *schema.GroupVersion) *genericapi.APIGroupVersion {
	resourceStorage := metricstorage.NewREST(s.Provider)

	storage := map[string]rest.Storage{
		// TODO: make this non-returning storage
		"*": resourceStorage,
		"*/*": resourceStorage,
	}

	return &genericapi.APIGroupVersion{
		Root: genericapiserver.APIGroupPrefix,
		GroupVersion: *groupVersion,

		ParameterCodec: metav1.ParameterCodec,
		Serializer: Codecs,
		Creater: Scheme,
		Convertor: Scheme,
		UnsafeConvertor: runtime.UnsafeObjectConvertor(Scheme),
		Copier: Scheme,
		Typer: Scheme,
		SubresourceGroupVersionKind: nil, // TODO: do we need this?
		Linker: groupMeta.SelfLinker,
		Mapper: groupMeta.RESTMapper,
		Storage: storage,

		// TODO: Admit?
		Context: s.GenericAPIServer.RequestContextMapper(),
		MinRequestTimeout: s.GenericAPIServer.MinRequestTimeout(),
		OptionsExternalVersion: &schema.GroupVersion{Version: "v1"},

		ResourceLister: provider.NewResourceLister(s.Provider),
	}
}
