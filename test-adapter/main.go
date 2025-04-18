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

package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/component-base/logs"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog/v2"

	"sigs.k8s.io/custom-metrics-apiserver/pkg/apiserver/metrics"
	basecmd "sigs.k8s.io/custom-metrics-apiserver/pkg/cmd"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
	fakeprov "sigs.k8s.io/custom-metrics-apiserver/test-adapter/provider"
)

type SampleAdapter struct {
	basecmd.AdapterBase

	// Message is printed on successful startup
	Message string
}

func (a *SampleAdapter) makeProviderOrDie() (provider.MetricsProvider, *restful.WebService) {
	client, err := a.DynamicClient()
	if err != nil {
		klog.Fatalf("unable to construct dynamic client: %v", err)
	}

	mapper, err := a.RESTMapper()
	if err != nil {
		klog.Fatalf("unable to construct discovery REST mapper: %v", err)
	}

	return fakeprov.NewFakeProvider(client, mapper)
}

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	cmd := &SampleAdapter{}
	cmd.Name = "test-adapter"

	cmd.Flags().StringVar(&cmd.Message, "msg", "starting adapter...", "startup message")
	logs.AddFlags(cmd.Flags())
	if err := cmd.Flags().Parse(os.Args); err != nil {
		klog.Fatalf("unable to parse flags: %v", err)
	}

	testProvider, webService := cmd.makeProviderOrDie()
	cmd.WithCustomMetrics(testProvider)
	cmd.WithExternalMetrics(testProvider)

	if err := metrics.RegisterMetrics(legacyregistry.Register); err != nil {
		klog.Fatalf("unable to register metrics: %v", err)
	}

	klog.Infof("%s", cmd.Message)
	// Set up POST endpoint for writing fake metric values
	restful.DefaultContainer.Add(webService)
	go func() {
		// Open port for POSTing fake metrics
		server := &http.Server{
			Addr:              ":8080",
			ReadHeaderTimeout: 3 * time.Second,
		}
		klog.Fatal(server.ListenAndServe())
	}()
	if err := cmd.Run(context.Background()); err != nil {
		klog.Fatalf("unable to run custom metrics adapter: %v", err)
	}
}
