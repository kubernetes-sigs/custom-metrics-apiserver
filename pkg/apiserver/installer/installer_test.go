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
	"net/http"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/handlers"
	"k8s.io/apiserver/pkg/endpoints/request"
)

type setTestSelfLinker struct {
	t           *testing.T
	expectedSet string
	name        string
	namespace   string
	called      bool
	err         error
}

func (s *setTestSelfLinker) Namespace(runtime.Object) (string, error) { return s.namespace, s.err }
func (s *setTestSelfLinker) Name(runtime.Object) (string, error)      { return s.name, s.err }
func (s *setTestSelfLinker) SelfLink(runtime.Object) (string, error)  { return "", s.err }
func (s *setTestSelfLinker) SetSelfLink(obj runtime.Object, selfLink string) error {
	if e, a := s.expectedSet, selfLink; e != a {
		s.t.Errorf("expected '%v', got '%v'", e, a)
	}
	s.called = true
	return s.err
}

func TestScopeNamingGenerateLink(t *testing.T) {
	selfLinker := &setTestSelfLinker{
		t:           t,
		expectedSet: "/api/v1/namespaces/other/services/foo",
		name:        "foo",
		namespace:   "other",
	}
	reqInfo := &request.RequestInfo{
		Resource: "services",
	}
	ctxFn := func(req *http.Request) request.Context {
		return request.NewContext()
	}
	s := MetricsNaming{
		handlers.ContextBasedNaming{
			GetContext:         ctxFn,
			SelfLinker:         selfLinker,
			ClusterScoped:      false,
			SelfLinkPathPrefix: "/api/v1/",
		},
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "other",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Service",
		},
	}
	_, err := s.GenerateLink(reqInfo, service)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
