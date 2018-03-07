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

package installer

import (
	"net/http"
	gpath "path"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/endpoints/handlers"
	"k8s.io/apiserver/pkg/endpoints/handlers/negotiation"
	"k8s.io/apiserver/pkg/endpoints/metrics"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"

	"github.com/emicklei/go-restful"
)

type EMHandlers struct {
}

// registerResourceHandlers registers the resource handlers for external metrics.
// The implementation is based on corresponding registerResourceHandlers for Custom Metrics API
func (ch *EMHandlers) registerResourceHandlers(a *MetricsAPIInstaller, ws *restful.WebService) error {
	context := a.group.Context

	optionsExternalVersion := a.group.GroupVersion
	if a.group.OptionsExternalVersion != nil {
		optionsExternalVersion = *a.group.OptionsExternalVersion
	}

	mapping, err := a.restMapping()
	if err != nil {
		return err
	}

	fqKindToRegister, err := a.getResourceKind(a.group.DynamicStorage)
	if err != nil {
		return err
	}

	kind := fqKindToRegister.Kind

	lister := a.group.DynamicStorage.(rest.Lister)
	list := lister.NewList()
	listGVKs, _, err := a.group.Typer.ObjectKinds(list)
	if err != nil {
		return err
	}
	versionedListPtr, err := a.group.Creater.New(a.group.GroupVersion.WithKind(listGVKs[0].Kind))
	if err != nil {
		return err
	}
	versionedList := indirectArbitraryPointer(versionedListPtr)

	versionedListOptions, err := a.group.Creater.New(optionsExternalVersion.WithKind("ListOptions"))
	if err != nil {
		return err
	}

	ctxFn := func(req *http.Request) request.Context {
		if ctx, ok := context.Get(req); ok {
			return request.WithUserAgent(ctx, req.Header.Get("User-Agent"))
		}
		return request.WithUserAgent(request.NewContext(), req.Header.Get("User-Agent"))
	}

	scope := mapping.Scope
	namespaceParam := ws.PathParameter(scope.ArgumentName(), scope.ParamDescription()).DataType("string")
	nameParam := ws.PathParameter("name", "name of the described resource").DataType("string")

	externalMetricParams := []*restful.Parameter{
		namespaceParam,
		nameParam,
	}
	externalMetricPath := scope.ParamName() + "/{" + scope.ArgumentName() + "}/{resource}"

	mediaTypes, streamMediaTypes := negotiation.MediaTypesForSerializer(a.group.Serializer)
	allMediaTypes := append(mediaTypes, streamMediaTypes...)
	ws.Produces(allMediaTypes...)

	reqScope := handlers.RequestScope{
		ContextFunc:     ctxFn,
		Serializer:      a.group.Serializer,
		ParameterCodec:  a.group.ParameterCodec,
		Creater:         a.group.Creater,
		Convertor:       a.group.Convertor,
		Copier:          a.group.Copier,
		Typer:           a.group.Typer,
		UnsafeConvertor: a.group.UnsafeConvertor,

		// TODO: This seems wrong for cross-group subresources. It makes an assumption that a subresource and its parent are in the same group version. Revisit this.
		Resource:    a.group.GroupVersion.WithResource("*"),
		Subresource: "*",
		Kind:        fqKindToRegister,

		MetaGroupVersion: metav1.SchemeGroupVersion,
	}
	if a.group.MetaGroupVersion != nil {
		reqScope.MetaGroupVersion = *a.group.MetaGroupVersion
	}

	doc := "list external metrics"
	reqScope.Namer = MetricsNaming{
		handlers.ContextBasedNaming{
			GetContext:         ctxFn,
			SelfLinker:         a.group.Linker,
			ClusterScoped:      false,
			SelfLinkPathPrefix: gpath.Join(a.prefix, scope.ParamName()) + "/",
		},
	}

	externalMetricHandler := metrics.InstrumentRouteFunc("LIST", "external-metrics", "", "", restfulListResource(lister, nil, reqScope, false, a.minRequestTimeout))

	externalMetricRoute := ws.GET(externalMetricPath).To(externalMetricHandler).
		Doc(doc).
		Param(ws.QueryParameter("pretty", "If 'true', then the output is pretty printed.")).
		Operation("list" + kind).
		Produces(allMediaTypes...).
		Returns(http.StatusOK, "OK", versionedList).
		Writes(versionedList)
	if err := addObjectParams(ws, externalMetricRoute, versionedListOptions); err != nil {
		return err
	}
	addParams(externalMetricRoute, externalMetricParams)
	ws.Route(externalMetricRoute)

	return nil
}
