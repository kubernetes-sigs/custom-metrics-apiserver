/*
Copyright 2014 The Kubernetes Authors.

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
	"fmt"
	"net/http"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/endpoints/handlers"
	"k8s.io/apiserver/pkg/endpoints/handlers/negotiation"
	"k8s.io/apiserver/pkg/endpoints/request"
)

// nothing will ever be sent down this channel
var neverExitWatch <-chan time.Time = make(chan time.Time)

// timeoutFactory abstracts watch timeout logic for testing
type TimeoutFactory interface {
	TimeoutCh() (<-chan time.Time, func() bool)
}

// realTimeoutFactory implements timeoutFactory
type realTimeoutFactory struct {
	timeout time.Duration
}

// TimeoutCh returns a channel which will receive something when the watch times out,
// and a cleanup function to call when this happens.
func (w *realTimeoutFactory) TimeoutCh() (<-chan time.Time, func() bool) {
	if w.timeout == 0 {
		return neverExitWatch, func() bool { return false }
	}
	t := time.NewTimer(w.timeout)
	return t.C, t.Stop
}

// serveWatch handles serving requests to the server
// TODO: the functionality in this method and in WatchServer.Serve is not cleanly decoupled.
func serveWatch(watcher watch.Interface, scope handlers.RequestScope, req *http.Request, w http.ResponseWriter, timeout time.Duration) {
	// negotiate for the stream serializer
	serializer, err := negotiation.NegotiateOutputStreamSerializer(req, scope.Serializer)
	if err != nil {
		writeError(&scope, err, w, req)
		return
	}
	framer := serializer.StreamSerializer.Framer
	streamSerializer := serializer.StreamSerializer.Serializer
	embedded := serializer.Serializer
	if framer == nil {
		writeError(&scope, fmt.Errorf("no framer defined for %q available for embedded encoding", serializer.MediaType), w, req)
		return
	}
	encoder := scope.Serializer.EncoderForVersion(streamSerializer, scope.Kind.GroupVersion())

	useTextFraming := serializer.EncodesAsText

	// find the embedded serializer matching the media type
	embeddedEncoder := scope.Serializer.EncoderForVersion(embedded, scope.Kind.GroupVersion())

	// TODO: next step, get back mediaTypeOptions from negotiate and return the exact value here
	mediaType := serializer.MediaType
	if mediaType != runtime.ContentTypeJSON {
		mediaType += ";stream=watch"
	}

	ctx := req.Context()
	requestInfo, ok := request.RequestInfoFrom(ctx)
	if !ok {
		writeError(&scope, fmt.Errorf("missing requestInfo"), w, req)
		return
	}

	server := &handlers.WatchServer{
		Watching: watcher,
		Scope:    scope,

		UseTextFraming:  useTextFraming,
		MediaType:       mediaType,
		Framer:          framer,
		Encoder:         encoder,
		EmbeddedEncoder: embeddedEncoder,
		Fixup: func(obj runtime.Object) {
			if err := setSelfLink(obj, requestInfo, scope.Namer); err != nil {
				utilruntime.HandleError(fmt.Errorf("failed to set link for object %v: %v", reflect.TypeOf(obj), err))
			}
		},

		TimeoutFactory: &realTimeoutFactory{timeout},
	}

	server.ServeHTTP(w, req)
}
