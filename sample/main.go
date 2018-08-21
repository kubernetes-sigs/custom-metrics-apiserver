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
	"flag"
	"os"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/util/logs"

	basecmd "github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/cmd"
	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
	fakeprov "github.com/kubernetes-incubator/custom-metrics-apiserver/sample/provider"
)

type SampleAdapter struct {
	basecmd.AdapterBase

	// Message is printed on succesful startup
	Message string
}

func (a *SampleAdapter) makeProviderOrDie() provider.MetricsProvider {
	client, err := a.DynamicClient()
	if err != nil {
		glog.Fatalf("unable to construct dynamic client: %v", err)
	}

	mapper, err := a.RESTMapper()
	if err != nil {
		glog.Fatalf("unable to construct discovery REST mapper: %v", err)
	}

	return fakeprov.NewFakeProvider(client, mapper)
}

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	cmd := &SampleAdapter{}
	cmd.Flags().StringVar(&cmd.Message, "msg", "starting adapter...", "startup message")
	cmd.Flags().AddGoFlagSet(flag.CommandLine) // make sure we get the glog flags
	cmd.Flags().Parse(os.Args)

	provider := cmd.makeProviderOrDie()
	cmd.WithCustomMetrics(provider)
	cmd.WithExternalMetrics(provider)

	glog.Infof(cmd.Message)
	if err := cmd.Run(wait.NeverStop); err != nil {
		glog.Fatalf("unable to run custom metrics adapter: %v", err)
	}
}
