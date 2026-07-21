/*
Copyright 2022 The Kubernetes Authors.

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

package options

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	openapicommon "k8s.io/kube-openapi/pkg/common"

	"sigs.k8s.io/custom-metrics-apiserver/pkg/apiserver"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		testName  string
		args      []string
		shouldErr bool
	}{
		{
			testName:  "only-secure-port",
			args:      []string{"--secure-port=6443"}, // default is 443, which requires privileges
			shouldErr: false,
		},
		{
			testName:  "secure-port-0",
			args:      []string{"--secure-port=0"}, // means: "don't serve HTTPS at all"
			shouldErr: false,
		},
		{
			testName:  "invalid-secure-port",
			args:      []string{"--secure-port=-1"},
			shouldErr: true,
		},
		{
			testName:  "empty-header",
			args:      []string{"--secure-port=6443", "--requestheader-username-headers=\" \""},
			shouldErr: true,
		},
		{
			testName:  "invalid-audit-log-format",
			args:      []string{"--secure-port=6443", "--audit-log-path=file", "--audit-log-format=txt"},
			shouldErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			o := NewCustomMetricsAdapterServerOptions()

			flagSet := pflag.NewFlagSet("", pflag.PanicOnError)
			o.AddFlags(flagSet)
			err := flagSet.Parse(c.args)
			assert.NoErrorf(t, err, "Error while parsing flags")

			errList := o.Validate()
			err = utilerrors.NewAggregate(errList)
			if c.shouldErr {
				assert.Errorf(t, err, "Expected error while validating options")
			} else {
				assert.NoErrorf(t, err, "Error while validating options")
			}
		})
	}
}

func TestApplyTo(t *testing.T) {
	cases := []struct {
		testName  string
		args      []string
		shouldErr bool
	}{
		{
			testName: "only-secure-port",
			args:     []string{"--secure-port=6443"}, // default is 443, which requires privileges
		},
		{
			testName:  "secure-port-0",
			args:      []string{"--secure-port=0"}, // means: "don't serve HTTPS at all"
			shouldErr: false,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			o := NewCustomMetricsAdapterServerOptions()

			// Unit tests have no Kubernetes cluster access
			o.Authentication.RemoteKubeConfigFileOptional = true
			o.Authorization.RemoteKubeConfigFileOptional = true

			flagSet := pflag.NewFlagSet("", pflag.PanicOnError)
			o.AddFlags(flagSet)
			err := flagSet.Parse(c.args)
			assert.NoErrorf(t, err, "Error while parsing flags")

			serverConfig := genericapiserver.NewRecommendedConfig(apiserver.Codecs)
			serverConfig.ClientConfig = &rest.Config{}
			serverConfig.SharedInformerFactory = informers.NewSharedInformerFactory(nil, 0)
			err = o.ApplyTo(serverConfig)

			defer func() {
				// Close the listener, if any
				if serverConfig.SecureServing != nil && serverConfig.SecureServing.Listener != nil {
					err := serverConfig.SecureServing.Listener.Close()
					assert.NoError(t, err)
				}
			}()

			assert.NoErrorf(t, err, "Error while applying options")
		})
	}
}

func TestApplyToOpenAPISecurityDefinitions(t *testing.T) {
	// A kubeconfig is enough to get a token review client, which is what makes
	// the delegating authenticator publish its BearerToken security scheme.
	kubeconfig := filepath.Join(t.TempDir(), "kubeconfig")
	require.NoError(t, os.WriteFile(kubeconfig, []byte(`apiVersion: v1
kind: Config
clusters:
- cluster: {server: https://127.0.0.1:6443}
  name: test
contexts:
- context: {cluster: test, user: test}
  name: test
current-context: test
users:
- name: test
  user: {token: test}
`), 0600))

	o := NewCustomMetricsAdapterServerOptions()
	o.Authentication.RemoteKubeConfigFile = kubeconfig
	o.Authentication.SkipInClusterLookup = true
	o.Authorization.RemoteKubeConfigFileOptional = true
	o.OpenAPIConfig = &openapicommon.Config{}

	flagSet := pflag.NewFlagSet("", pflag.PanicOnError)
	o.AddFlags(flagSet)
	require.NoError(t, flagSet.Parse([]string{"--secure-port=0"}))

	serverConfig := genericapiserver.NewRecommendedConfig(apiserver.Codecs)
	serverConfig.ClientConfig = &rest.Config{}
	serverConfig.SharedInformerFactory = informers.NewSharedInformerFactory(nil, 0)
	require.NoError(t, o.ApplyTo(serverConfig))

	require.NotNil(t, o.OpenAPIConfig.SecurityDefinitions, "security definitions were not published to the OpenAPI config")
	assert.Contains(t, *o.OpenAPIConfig.SecurityDefinitions, "BearerToken")
	assert.Same(t, o.OpenAPIConfig, serverConfig.OpenAPIConfig)
}
