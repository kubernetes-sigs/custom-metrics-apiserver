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

package apiserver

import (
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/metrics/pkg/apis/custom_metrics"
)

func setTableColumnDefine(table *metav1beta1.Table) {
	table.ColumnDefinitions = []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: "Name of the resource"},
	}
	table.ColumnDefinitions = append(table.ColumnDefinitions, metav1beta1.TableColumnDefinition{
		Name:   "MetricsName",
		Type:   "string",
		Format: "",
	})
	table.ColumnDefinitions = append(table.ColumnDefinitions, metav1beta1.TableColumnDefinition{
		Name:   "MetricsValue",
		Type:   "string",
		Format: "",
	})
	table.ColumnDefinitions = append(table.ColumnDefinitions, metav1beta1.TableColumnDefinition{
		Name:   "TimeStamp",
		Type:   "string",
		Format: "",
	})

}

func addCustomMetricsToTable(table *metav1beta1.Table, nameSpaced bool, metrics map[custom_metrics.ObjectReference][]custom_metrics.MetricValue) {
	var obj []custom_metrics.ObjectReference
	for k := range metrics {
		obj = append(obj, k)
	}
	if nameSpaced {
		sort.Slice(obj, func(i, j int) bool {
			return obj[i].Namespace < obj[j].Namespace
		})
	}
	sort.Slice(obj, func(i, j int) bool {
		return obj[i].Name < obj[j].Name
	})
	for _, key := range obj {
		metric := metrics[key]
		sort.Slice(metric, func(i, j int) bool {
			return metric[i].Metric.Name < metric[j].Metric.Name
		})
		for _, item := range metric {
			row := make([]interface{}, 0, 5)
			if (nameSpaced && key.Namespace == "") || key.Kind == "" || key.Name == "" {
				klog.InfoS("bad resource record", "ObjectReference", key)
				continue
			}
			if item.Metric.Name == "" || item.Value.String() == "" {
				klog.InfoS("bad metrics record", "ObjectReference", key, "metrics", item)
				continue
			}
			if nameSpaced {
				row = append(row, fmt.Sprintf("%v/%v/%v", key.Namespace, key.Kind, key.Name))
			} else {
				row = append(row, fmt.Sprintf("%v/%v", key.Kind, key.Name))
			}

			row = append(row, item.Metric.Name)
			row = append(row, item.Value.String())
			if !item.Timestamp.IsZero() {
				row = append(row, item.Timestamp)
			} else {
				row = append(row, "  ")
			}
			table.Rows = append(table.Rows, metav1beta1.TableRow{
				Cells:  row,
				Object: runtime.RawExtension{Object: &item},
			})
		}
	}
	row := make([]interface{}, 0, 5)
	row = append(row, " ")
	row = append(row, " ")
	row = append(row, " ")
	row = append(row, " ")
	item := custom_metrics.MetricValue{}
	table.Rows = append(table.Rows, metav1.TableRow{Cells: row, Object: runtime.RawExtension{Object: &item}})
}
