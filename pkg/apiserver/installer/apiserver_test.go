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

package installer

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	genericapi "k8s.io/apiserver/pkg/endpoints"
	genericapifilters "k8s.io/apiserver/pkg/endpoints/filters"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/dynamic/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/metrics/pkg/apis/custom_metrics"
	installcm "k8s.io/metrics/pkg/apis/custom_metrics/install"
	cmv1beta1 "k8s.io/metrics/pkg/apis/custom_metrics/v1beta1"
	"k8s.io/metrics/pkg/apis/external_metrics"
	installem "k8s.io/metrics/pkg/apis/external_metrics/install"
	emv1beta1 "k8s.io/metrics/pkg/apis/external_metrics/v1beta1"

	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
	custommetricstorage "sigs.k8s.io/custom-metrics-apiserver/pkg/registry/custom_metrics"
	externalmetricstorage "sigs.k8s.io/custom-metrics-apiserver/pkg/registry/external_metrics"
	sampleprovider "sigs.k8s.io/custom-metrics-apiserver/test-adapter/provider"
)

// defaultAPIServer exposes nested objects for testability.
type defaultAPIServer struct {
	http.Handler
	container *restful.Container
}

var (
	Scheme                      = runtime.NewScheme()
	Codecs                      = serializer.NewCodecFactory(Scheme)
	prefix                      = genericapiserver.APIGroupPrefix
	customMetricsGroupVersion   schema.GroupVersion
	customMetricsGroupInfo      genericapiserver.APIGroupInfo
	externalMetricsGroupVersion schema.GroupVersion
	externalMetricsGroupInfo    genericapiserver.APIGroupInfo
	codec                       = Codecs.LegacyCodec()
	emptySet                    = labels.Set{}
	matchingSet                 = labels.Set{"foo": "bar"}
)

func init() {
	installcm.Install(Scheme)
	installem.Install(Scheme)

	// we need custom conversion functions to list resources with options
	utilruntime.Must(RegisterConversions(Scheme))

	// we need to add the options to empty v1
	// TODO fix the server code to avoid this
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})

	// TODO: keep the generic API server from wanting this
	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)

	customMetricsGroupInfo = genericapiserver.NewDefaultAPIGroupInfo(custom_metrics.GroupName, Scheme, runtime.NewParameterCodec(Scheme), Codecs)
	customMetricsGroupVersion = customMetricsGroupInfo.PrioritizedVersions[0]
	externalMetricsGroupInfo = genericapiserver.NewDefaultAPIGroupInfo(external_metrics.GroupName, Scheme, runtime.NewParameterCodec(Scheme), Codecs)
	externalMetricsGroupVersion = externalMetricsGroupInfo.PrioritizedVersions[0]
}

func extractBody(response *http.Response, object runtime.Object) error {
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	return runtime.DecodeInto(codec, body, object)
}

func extractBodyString(response *http.Response) (string, error) {
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(body), err
}

func apiGroupVersion(gv schema.GroupVersion, groupInfo genericapiserver.APIGroupInfo) *genericapi.APIGroupVersion {
	return &genericapi.APIGroupVersion{
		Root:             prefix,
		GroupVersion:     gv,
		MetaGroupVersion: groupInfo.MetaGroupVersion,
		ParameterCodec:   groupInfo.ParameterCodec,
		Serializer:       groupInfo.NegotiatedSerializer,
		Creater:          groupInfo.Scheme,
		Convertor:        groupInfo.Scheme,
		UnsafeConvertor:  runtime.UnsafeObjectConvertor(groupInfo.Scheme),
		Typer:            groupInfo.Scheme,
		Namer:            runtime.Namer(meta.NewAccessor()),
	}
}

func handleCustomMetrics(prov provider.CustomMetricsProvider) http.Handler {
	container := restful.NewContainer()
	container.Router(restful.CurlyRouter{})
	mux := container.ServeMux
	resourceStorage := custommetricstorage.NewREST(prov)
	group := &MetricsAPIGroupVersion{
		DynamicStorage:  resourceStorage,
		APIGroupVersion: apiGroupVersion(customMetricsGroupVersion, customMetricsGroupInfo),
		ResourceLister:  provider.NewCustomMetricResourceLister(prov),
		Handlers:        &CMHandlers{},
	}

	if err := group.InstallREST(container); err != nil {
		panic(fmt.Sprintf("unable to install container %s: %v", group.GroupVersion, err))
	}

	var handler http.Handler = &defaultAPIServer{mux, container}
	reqInfoResolver := genericapiserver.NewRequestInfoResolver(&genericapiserver.Config{})
	handler = genericapifilters.WithRequestInfo(handler, reqInfoResolver)
	return handler
}

func handleExternalMetrics(prov provider.ExternalMetricsProvider) http.Handler {
	container := restful.NewContainer()
	container.Router(restful.CurlyRouter{})
	mux := container.ServeMux
	resourceStorage := externalmetricstorage.NewREST(prov)

	group := &MetricsAPIGroupVersion{
		DynamicStorage:  resourceStorage,
		APIGroupVersion: apiGroupVersion(externalMetricsGroupVersion, externalMetricsGroupInfo),
		ResourceLister:  provider.NewExternalMetricResourceLister(prov),
		Handlers:        &EMHandlers{},
	}

	if err := group.InstallREST(container); err != nil {
		panic(fmt.Sprintf("unable to install container %s: %v", group.GroupVersion, err))
	}

	var handler http.Handler = &defaultAPIServer{mux, container}
	reqInfoResolver := genericapiserver.NewRequestInfoResolver(&genericapiserver.Config{})
	handler = genericapifilters.WithRequestInfo(handler, reqInfoResolver)
	return handler
}

type T struct {
	Method         string
	Path           string
	Status         int
	ExpectedCount  int
	IsResourceList bool
}

func TestCustomMetricsAPI(t *testing.T) {
	totalNodesCount := 4
	totalPodsCount := 16
	matchingNodesCount := 2
	matchingPodsCount := 8

	cases := map[string]T{
		// checks which should fail
		"GET long prefix": {"GET", prefix + "/", http.StatusNotFound, 0, false},

		"root GET missing storage": {"GET", prefix + "/" + customMetricsGroupVersion.Group + "/" + customMetricsGroupVersion.Version + "/blah", http.StatusNotFound, 0, false},

		"GET at root resource leaf":       {"GET", prefix + "/" + customMetricsGroupVersion.Group + "/" + customMetricsGroupVersion.Version + "/nodes/foo", http.StatusNotFound, 0, false},
		"GET at namespaced resource leaf": {"GET", prefix + "/" + customMetricsGroupVersion.Group + "/" + customMetricsGroupVersion.Version + "/namespaces/ns/pods/bar", http.StatusNotFound, 0, false},

		// Positive checks to make sure everything is wired correctly
		"GET for all nodes (root)":                 {"GET", prefix + "/" + customMetricsGroupVersion.Group + "/" + customMetricsGroupVersion.Version + "/nodes/*/some-metric", http.StatusOK, totalNodesCount, false},
		"GET for all pods (namespaced)":            {"GET", prefix + "/" + customMetricsGroupVersion.Group + "/" + customMetricsGroupVersion.Version + "/namespaces/ns/pods/*/some-metric", http.StatusOK, totalPodsCount, false},
		"GET for namespace":                        {"GET", prefix + "/" + customMetricsGroupVersion.Group + "/" + customMetricsGroupVersion.Version + "/namespaces/ns/metrics/some-metric", http.StatusOK, 1, false},
		"GET for label selected nodes (root)":      {"GET", prefix + "/" + customMetricsGroupVersion.Group + "/" + customMetricsGroupVersion.Version + "/nodes/*/some-metric?labelSelector=foo%3Dbar", http.StatusOK, matchingNodesCount, false},
		"GET for label selected pods (namespaced)": {"GET", prefix + "/" + customMetricsGroupVersion.Group + "/" + customMetricsGroupVersion.Version + "/namespaces/ns/pods/*/some-metric?labelSelector=foo%3Dbar", http.StatusOK, matchingPodsCount, false},
		"GET for single node (root)":               {"GET", prefix + "/" + customMetricsGroupVersion.Group + "/" + customMetricsGroupVersion.Version + "/nodes/foo/some-metric", http.StatusOK, 1, false},
		"GET for single pod (namespaced)":          {"GET", prefix + "/" + customMetricsGroupVersion.Group + "/" + customMetricsGroupVersion.Version + "/namespaces/ns/pods/foo/some-metric", http.StatusOK, 1, false},

		"GET all metrics": {"GET", prefix + "/" + customMetricsGroupVersion.Group + "/" + customMetricsGroupVersion.Version, http.StatusOK, 3, true},
	}

	scheme := runtime.NewScheme()
	err := corev1.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}
	dynClient := fake.NewSimpleDynamicClient(scheme)
	dynClient.PrependReactor("list", "*", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		var items []unstructured.Unstructured
		switch action.GetResource().Resource {
		case "nodes":
			items = make([]unstructured.Unstructured, totalNodesCount)
			for i := 0; i < totalNodesCount; i++ {
				items[i].SetName("*")
				if i < matchingNodesCount {
					items[i].SetLabels(matchingSet)
				}
			}
		case "pods":
			items = make([]unstructured.Unstructured, totalPodsCount)
			for i := 0; i < totalPodsCount; i++ {
				items[i].SetName("*")
				if i < matchingPodsCount {
					items[i].SetLabels(matchingSet)
				}
			}
		default:
			return false, nil, nil
		}
		return true, &unstructured.UnstructuredList{Items: items}, nil
	})
	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{})
	mapper.Add(schema.FromAPIVersionAndKind("v1", "Node"), meta.RESTScopeRoot)
	mapper.Add(schema.FromAPIVersionAndKind("v1", "Pod"), meta.RESTScopeNamespace)
	mapper.Add(schema.FromAPIVersionAndKind("v1", "Namespace"), meta.RESTScopeRoot)

	prov, _ := sampleprovider.NewFakeProvider(dynClient, mapper)
	prov.UpdateMetric("", "nodes", "foo", "some-metric", resource.NewMilliQuantity(300, resource.DecimalSI), emptySet)
	prov.UpdateMetric("", "nodes", "*", "some-metric", resource.NewMilliQuantity(300, resource.DecimalSI), emptySet)
	prov.UpdateMetric("ns", "pods", "foo", "some-metric", resource.NewMilliQuantity(300, resource.DecimalSI), emptySet)
	prov.UpdateMetric("ns", "pods", "*", "some-metric", resource.NewMilliQuantity(300, resource.DecimalSI), emptySet)
	prov.UpdateMetric("", "namespaces", "ns", "some-metric", resource.NewMilliQuantity(300, resource.DecimalSI), emptySet)

	server := httptest.NewServer(handleCustomMetrics(prov))
	defer server.Close()
	client := http.Client{}
	for k, v := range cases {
		response, err := executeRequest(t, k, v, server, &client)
		if assert.NoError(t, err) && v.ExpectedCount > 0 {
			if v.IsResourceList {
				lst := &metav1.APIResourceList{}
				err := extractBody(response, lst)
				if assert.NoErrorf(t, err, "unexpected error (%s)", k) {
					assert.Equalf(t, v.ExpectedCount, len(lst.APIResources), "(%s)", k)
				}
			} else {
				lst := &cmv1beta1.MetricValueList{}
				err := extractBody(response, lst)
				if assert.NoErrorf(t, err, "unexpected error (%s)", k) {
					assert.Equalf(t, v.ExpectedCount, len(lst.Items), "(%s)", k)
				}
			}
		}
	}
}

func TestExternalMetricsAPI(t *testing.T) {
	cases := map[string]T{
		// checks which should fail
		"GET long prefix":             {"GET", prefix + "/", http.StatusNotFound, 0, false},
		"GET at root scope":           {"GET", prefix + "/" + externalMetricsGroupVersion.Group + "/" + externalMetricsGroupVersion.Version + "/nonexistent-metric", http.StatusNotFound, 0, false},
		"GET without metric name":     {"GET", prefix + "/" + externalMetricsGroupVersion.Group + "/" + externalMetricsGroupVersion.Version + "/namespaces/foo", http.StatusNotFound, 0, false},
		"GET for metric with slashes": {"GET", prefix + "/" + externalMetricsGroupVersion.Group + "/" + externalMetricsGroupVersion.Version + "/namespaces/foo/group/metric", http.StatusNotFound, 0, false},

		// Positive checks to make sure everything is wired correctly
		"GET for external metric":               {"GET", prefix + "/" + externalMetricsGroupVersion.Group + "/" + externalMetricsGroupVersion.Version + "/namespaces/default/my-external-metric", http.StatusOK, 2, false},
		"GET for external metric with selector": {"GET", prefix + "/" + externalMetricsGroupVersion.Group + "/" + externalMetricsGroupVersion.Version + "/namespaces/default/my-external-metric?labelSelector=foo%3Dbar", http.StatusOK, 1, false},
		"GET for nonexistent metric":            {"GET", prefix + "/" + externalMetricsGroupVersion.Group + "/" + externalMetricsGroupVersion.Version + "/namespaces/foo/nonexistent-metric", http.StatusOK, 0, false},

		"GET all metrics": {"GET", prefix + "/" + externalMetricsGroupVersion.Group + "/" + externalMetricsGroupVersion.Version, http.StatusOK, 3, true},
	}

	// "real" fake provider implementation can be used in test, because it doesn't have any dependencies.
	// Note: this provider has a hardcoded list of external metrics.
	prov, _ := sampleprovider.NewFakeProvider(nil, nil)

	server := httptest.NewServer(handleExternalMetrics(prov))
	defer server.Close()
	client := http.Client{}
	for k, v := range cases {
		response, err := executeRequest(t, k, v, server, &client)
		if assert.NoError(t, err) && v.ExpectedCount > 0 {
			if v.IsResourceList {
				lst := &metav1.APIResourceList{}
				err := extractBody(response, lst)
				if assert.NoErrorf(t, err, "unexpected error (%s)", k) {
					assert.Equalf(t, v.ExpectedCount, len(lst.APIResources), "(%s)", k)
				}
			} else {
				lst := &emv1beta1.ExternalMetricValueList{}
				err := extractBody(response, lst)
				if assert.NoErrorf(t, err, "unexpected error (%s)", k) {
					assert.Equalf(t, v.ExpectedCount, len(lst.Items), "(%s)", k)
				}
			}
		}
	}
}

func executeRequest(t *testing.T, k string, v T, server *httptest.Server, client *http.Client) (*http.Response, error) {
	request, err := http.NewRequest(v.Method, server.URL+v.Path, nil)
	if err != nil {
		t.Fatalf("unexpected error (%s): %v", k, err)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unexpected error (%s): %v", k, err)
	}

	if response.StatusCode != v.Status {
		body, err := extractBodyString(response)
		bodyPart := body
		if err != nil {
			bodyPart = fmt.Sprintf("[error extracting body: %v]", err)
		}
		return nil, fmt.Errorf("Expected %d for %s (%s), Got %#v -- %s", v.Status, v.Method, k, response, bodyPart)
	}
	return response, nil
}
