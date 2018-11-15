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

package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/handlers"
	"k8s.io/apiserver/pkg/endpoints/metrics"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	utiltrace "k8s.io/utils/trace"

	cm_rest "github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/apiserver/registry/rest"
)

func ListResourceWithOptions(r cm_rest.ListerWithOptions, rw rest.Watcher, scope handlers.RequestScope, forceWatch bool, minRequestTimeout time.Duration) http.HandlerFunc {
	list := func(w http.ResponseWriter, req *http.Request, ctx context.Context, opts *metainternalversion.ListOptions) (runtime.Object, error) {
		extraOpts, subpath, subpathKey := r.NewListOptions()
		if err := getRequestOptions(req, scope, extraOpts, subpath, subpathKey, false); err != nil {
			err = errors.NewBadRequest(err.Error())
			return nil, err
		}
		return r.List(ctx, opts, extraOpts)
	}
	return listResource(list, rw, scope, forceWatch, minRequestTimeout)
}

// getRequestOptions parses out options and can include path information.  The path information shouldn't include the subresource.
func getRequestOptions(req *http.Request, scope handlers.RequestScope, into runtime.Object, subpath bool, subpathKey string, isSubresource bool) error {
	if into == nil {
		return nil
	}

	query := req.URL.Query()
	if subpath {
		newQuery := make(url.Values)
		for k, v := range query {
			newQuery[k] = v
		}

		ctx := req.Context()
		requestInfo, _ := request.RequestInfoFrom(ctx)
		startingIndex := 2
		if isSubresource {
			startingIndex = 3
		}

		p := strings.Join(requestInfo.Parts[startingIndex:], "/")

		// ensure non-empty subpaths correctly reflect a leading slash
		if len(p) > 0 && !strings.HasPrefix(p, "/") {
			p = "/" + p
		}

		// ensure subpaths correctly reflect the presence of a trailing slash on the original request
		if strings.HasSuffix(requestInfo.Path, "/") && !strings.HasSuffix(p, "/") {
			p += "/"
		}

		newQuery[subpathKey] = []string{p}
		query = newQuery
	}
	return scope.ParameterCodec.DecodeParameters(query, scope.Kind.GroupVersion(), into)
}

func listResource(list func(http.ResponseWriter, *http.Request, context.Context, *metainternalversion.ListOptions) (runtime.Object, error), rw rest.Watcher, scope handlers.RequestScope, forceWatch bool, minRequestTimeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// For performance tracking purposes.
		trace := utiltrace.New("List " + req.URL.Path)

		namespace, err := scope.Namer.Namespace(req)
		if err != nil {
			writeError(&scope, err, w, req)
			return
		}

		// Watches for single objects are routed to this function.
		// Treat a name parameter the same as a field selector entry.
		hasName := true
		_, name, err := scope.Namer.Name(req)
		if err != nil {
			hasName = false
		}

		ctx := req.Context()
		ctx = request.WithNamespace(ctx, namespace)

		opts := metainternalversion.ListOptions{}
		if err := metainternalversion.ParameterCodec.DecodeParameters(req.URL.Query(), scope.MetaGroupVersion, &opts); err != nil {
			err = errors.NewBadRequest(err.Error())
			writeError(&scope, err, w, req)
			return
		}

		// transform fields
		// TODO: DecodeParametersInto should do this.
		if opts.FieldSelector != nil {
			fn := func(label, value string) (newLabel, newValue string, err error) {
				return scope.Convertor.ConvertFieldLabel(scope.Kind, label, value)
			}
			if opts.FieldSelector, err = opts.FieldSelector.Transform(fn); err != nil {
				// TODO: allow bad request to set field causes based on query parameters
				err = errors.NewBadRequest(err.Error())
				writeError(&scope, err, w, req)
				return
			}
		}

		if hasName {
			// metadata.name is the canonical internal name.
			// SelectionPredicate will notice that this is a request for
			// a single object and optimize the storage query accordingly.
			nameSelector := fields.OneTermEqualSelector("metadata.name", name)

			// Note that fieldSelector setting explicitly the "metadata.name"
			// will result in reaching this branch (as the value of that field
			// is propagated to requestInfo as the name parameter.
			// That said, the allowed field selectors in this branch are:
			// nil, fields.Everything and field selector matching metadata.name
			// for our name.
			if opts.FieldSelector != nil && !opts.FieldSelector.Empty() {
				selectedName, ok := opts.FieldSelector.RequiresExactMatch("metadata.name")
				if !ok || name != selectedName {
					writeError(&scope, errors.NewBadRequest("fieldSelector metadata.name doesn't match requested name"), w, req)
					return
				}
			} else {
				opts.FieldSelector = nameSelector
			}
		}

		if opts.Watch || forceWatch {
			if rw == nil {
				writeError(&scope, errors.NewMethodNotSupported(scope.Resource.GroupResource(), "watch"), w, req)
				return
			}
			// TODO: Currently we explicitly ignore ?timeout= and use only ?timeoutSeconds=.
			timeout := time.Duration(0)
			if opts.TimeoutSeconds != nil {
				timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
			}
			if timeout == 0 && minRequestTimeout > 0 {
				timeout = time.Duration(float64(minRequestTimeout) * (rand.Float64() + 1.0))
			}
			glog.V(3).Infof("Starting watch for %s, rv=%s labels=%s fields=%s timeout=%s", req.URL.Path, opts.ResourceVersion, opts.LabelSelector, opts.FieldSelector, timeout)

			watcher, err := rw.Watch(ctx, &opts)
			if err != nil {
				writeError(&scope, err, w, req)
				return
			}
			requestInfo, _ := request.RequestInfoFrom(ctx)
			metrics.RecordLongRunning(req, requestInfo, func() {
				serveWatch(watcher, scope, req, w, timeout)
			})
			return
		}

		// Log only long List requests (ignore Watch).
		defer trace.LogIfLong(500 * time.Millisecond)
		trace.Step("About to List from storage")
		result, err := list(w, req, ctx, &opts)
		if err != nil {
			writeError(&scope, err, w, req)
			return
		}
		trace.Step("Listing from storage done")
		numberOfItems, err := setListSelfLink(result, ctx, req, scope.Namer)
		if err != nil {
			writeError(&scope, err, w, req)
			return
		}
		trace.Step("Self-linking done")
		// Ensure empty lists return a non-nil items slice
		if numberOfItems == 0 && meta.IsListType(result) {
			if err := meta.SetList(result, []runtime.Object{}); err != nil {
				writeError(&scope, err, w, req)
				return
			}
		}

		transformResponseObject(ctx, scope, req, w, http.StatusOK, result)
		trace.Step(fmt.Sprintf("Writing http response done (%d items)", numberOfItems))
	}
}
