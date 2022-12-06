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

package metrics

import (
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/component-base/metrics/testutil"
	"k8s.io/metrics/pkg/apis/custom_metrics"
	"k8s.io/metrics/pkg/apis/external_metrics"
	clocktesting "k8s.io/utils/clock/testing"
)

func TestFreshness(t *testing.T) {
	now := time.Now()

	metricFreshness.Create(nil)
	metricFreshness.Reset()

	externalMetricsList := external_metrics.ExternalMetricValueList{
		Items: []external_metrics.ExternalMetricValue{
			{Timestamp: metav1.NewTime(now.Add(-10 * time.Second))},
			{Timestamp: metav1.NewTime(now.Add(-10 * time.Second))},
			{Timestamp: metav1.NewTime(now.Add(-2 * time.Second))},
		},
	}
	externalObserver := NewFreshnessObserver("external.metrics.k8s.io")
	externalObserver.(*freshnessObserver).clock = clocktesting.NewFakeClock(now)
	for _, m := range externalMetricsList.Items {
		externalObserver.Observe(m.Timestamp)
	}

	customMetricsList := custom_metrics.MetricValueList{
		Items: []custom_metrics.MetricValue{
			{Timestamp: metav1.NewTime(now.Add(-5 * time.Second))},
			{Timestamp: metav1.NewTime(now.Add(-10 * time.Second))},
			{Timestamp: metav1.NewTime(now.Add(-25 * time.Second))},
		},
	}
	customObserver := NewFreshnessObserver("custom.metrics.k8s.io")
	customObserver.(*freshnessObserver).clock = clocktesting.NewFakeClock(now)
	for _, m := range customMetricsList.Items {
		customObserver.Observe(m.Timestamp)
	}

	err := testutil.CollectAndCompare(metricFreshness, strings.NewReader(`
	# HELP metrics_apiserver_metric_freshness_seconds [ALPHA] Freshness of metrics exported
	# TYPE metrics_apiserver_metric_freshness_seconds histogram
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="1"} 0
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="1.364"} 0
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="1.8604960000000004"} 0
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="2.5377165440000007"} 0
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="3.4614453660160014"} 0
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="4.721411479245826"} 0
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="6.440005257691307"} 1
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="8.784167171490942"} 1
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="11.981604021913647"} 2
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="16.342907885890217"} 2
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="22.291726356354257"} 2
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="30.405914750067208"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="41.47366771909167"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="56.57008276884105"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="77.16159289669919"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="105.2484127110977"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="143.55883493793726"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="195.81425085534644"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="267.09063816669254"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="364.31163045936864"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="custom.metrics.k8s.io",le="+Inf"} 3
	metrics_apiserver_metric_freshness_seconds_sum{group="custom.metrics.k8s.io"} 40
	metrics_apiserver_metric_freshness_seconds_count{group="custom.metrics.k8s.io"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="1"} 0
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="1.364"} 0
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="1.8604960000000004"} 0
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="2.5377165440000007"} 1
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="3.4614453660160014"} 1
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="4.721411479245826"} 1
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="6.440005257691307"} 1
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="8.784167171490942"} 1
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="11.981604021913647"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="16.342907885890217"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="22.291726356354257"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="30.405914750067208"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="41.47366771909167"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="56.57008276884105"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="77.16159289669919"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="105.2484127110977"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="143.55883493793726"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="195.81425085534644"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="267.09063816669254"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="364.31163045936864"} 3
	metrics_apiserver_metric_freshness_seconds_bucket{group="external.metrics.k8s.io",le="+Inf"} 3
	metrics_apiserver_metric_freshness_seconds_sum{group="external.metrics.k8s.io"} 22
	metrics_apiserver_metric_freshness_seconds_count{group="external.metrics.k8s.io"} 3
	`))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
