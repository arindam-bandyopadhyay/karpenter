/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package fakes

import (
	"fmt"
	"net/http"
	"testing"

	prometheusmodel "github.com/prometheus/client_model/go"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	crmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	MetricsNamespace              = "karpenter"
	MetricsCloudProviderSubsystem = "cloudprovider"
)

type TestResponse struct {
	StatusCode int
	Count      int
}

func (r *TestResponse) HTTPResponse() *http.Response {
	return &http.Response{
		StatusCode: r.StatusCode,
	}
}

func getMetricFullName(namespace string, subSystem string, name string) string {
	return fmt.Sprintf("%s_%s_%s", namespace, subSystem, name)
}

func FindMetricWithLabelValues(t *testing.T, name string,
	labelValues map[string]string) (*prometheusmodel.Metric, bool) {
	fullName := getMetricFullName(MetricsNamespace, MetricsCloudProviderSubsystem, name)

	metrics, err := crmetrics.Registry.Gather()
	assert.NoError(t, err)

	mf, found := lo.Find(metrics, func(mf *prometheusmodel.MetricFamily) bool {
		return mf.GetName() == fullName
	})
	if !found {
		return nil, false
	}
	for _, m := range mf.Metric {
		temp := lo.Assign(labelValues)
		for _, labelPair := range m.Label {
			if v, ok := temp[labelPair.GetName()]; ok && v == labelPair.GetValue() {
				delete(temp, labelPair.GetName())
			}
		}
		if len(temp) == 0 {
			return m, true
		}
	}
	return nil, false
}
