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

package context

import (
	"k8s.io/apiserver/pkg/endpoints/request"
)

// resourceInformation holds the resource and subresource for a request in the context.
type resourceInformation struct {
	resource    string
	subresource string
}

// contextKey is the type of the keys for the context in this file.
// It's private to avoid conflicts across packages.
type contextKey int

const resourceKey contextKey = iota

// WithResourceInformation returns a copy of parent in which the resource and subresource values are set
func WithResourceInformation(parent request.Context, resource, subresource string) request.Context {
	return request.WithValue(parent, resourceKey, resourceInformation{resource, subresource})
}

// ResourceInformationFrom returns resource and subresource on the ctx
func ResourceInformationFrom(ctx request.Context) (resource string, subresource string, ok bool) {
	resourceInfo, ok := ctx.Value(resourceKey).(resourceInformation)
	if !ok {
		return "", "", ok
	}

	return resourceInfo.resource, resourceInfo.subresource, ok
}
