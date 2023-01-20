// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

//go:build !windows

package metric_value_benchmark

import (
	_ "embed"
	"fmt"
	"github.com/aws/amazon-cloudwatch-agent-test/internal/common"
	"github.com/aws/amazon-cloudwatch-agent-test/test/metric"
	"github.com/aws/amazon-cloudwatch-agent-test/test/metric/dimension"
	"github.com/aws/amazon-cloudwatch-agent-test/test/status"
	"github.com/aws/amazon-cloudwatch-agent-test/test/test_runner"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type PrometheusTestRunner struct {
	test_runner.BaseTestRunner
}

var _ test_runner.ITestRunner = (*PrometheusTestRunner)(nil)

//go:embed agent_configs/prometheus.yaml

const (
	prometheusConfigPathIn   = "resources/prometheus.yaml"
	prometheusConfigPathOut  = "/opt/aws/amazon-cloudwatch-agent/bin/prometheus_config.yaml"
	prometheusMetricsPathIn  = "resources/prometheus_metrics.txt"
	prometheusMetricsPathOut = "/tmp/metrics"
)

func (t *PrometheusTestRunner) Validate() status.TestGroupResult {
	metricsToFetch := t.GetMeasuredMetrics()
	testResults := make([]status.TestResult, len(metricsToFetch))
	for i, metricName := range metricsToFetch {
		testResults[i] = t.validatePrometheusMetric(metricName)
	}

	return status.TestGroupResult{
		Name:        t.GetTestName(),
		TestResults: testResults,
	}
}

func (t *PrometheusTestRunner) GetTestName() string {
	return "Prometheus"
}

func (t *PrometheusTestRunner) GetAgentConfigFileName() string {
	return "prometheus_config.json"
}

func (t *PrometheusTestRunner) SetupBeforeAgentRun() error {
	common.CopyFile(prometheusConfigPathIn, prometheusConfigPathOut)
	common.CopyFile(prometheusMetricsPathIn, prometheusMetricsPathOut)
	startPrometheusCommands := []string{
		"sudo python3 -m http.server 8101 --directory /tmp &> /dev/null &",
	}

	return common.RunCommands(startPrometheusCommands)
}

func (t *PrometheusTestRunner) GetMeasuredMetrics() []string {
	return []string{
		"prometheus_test_counter",
		"prometheus_test_gauge",
		"prometheus_test_summary_count",
		"prometheus_test_summary_sum",
		"prometheus_test_summary",
	}
}

func (t *PrometheusTestRunner) validatePrometheusMetric(metricName string) status.TestResult {
	testResult := status.TestResult{
		Name:   metricName,
		Status: status.FAILED,
	}

	var dims []types.Dimension
	var failed []dimension.Instruction

	switch metricName {
	case "prometheus_test_counter":
		dims, failed = t.DimensionFactory.GetDimensions([]dimension.Instruction{
			{
				Key:   "prom_metric_type",
				Value: dimension.ExpectedDimensionValue{aws.String("counter")},
			},
		})
	case "prometheus_test_gauge":
		dims, failed = t.DimensionFactory.GetDimensions([]dimension.Instruction{
			{
				Key:   "prom_metric_type",
				Value: dimension.ExpectedDimensionValue{aws.String("gauge")},
			},
		})
	case "prometheus_test_summary_count":
		dims, failed = t.DimensionFactory.GetDimensions([]dimension.Instruction{
			{
				Key:   "prom_metric_type",
				Value: dimension.ExpectedDimensionValue{aws.String("summary")},
			},
		})
	case "prometheus_test_summary_sum":
		dims, failed = t.DimensionFactory.GetDimensions([]dimension.Instruction{
			{
				Key:   "prom_metric_type",
				Value: dimension.ExpectedDimensionValue{aws.String("summary")},
			},
		})
	case "prometheus_test_summary":
		dims, failed = t.DimensionFactory.GetDimensions([]dimension.Instruction{
			{
				Key:   "prom_metric_type",
				Value: dimension.ExpectedDimensionValue{aws.String("summary")},
			},
			{
				Key:   "quantile",
				Value: dimension.ExpectedDimensionValue{aws.String("0.5")},
			},
		})
	default:
		dims, failed = t.DimensionFactory.GetDimensions([]dimension.Instruction{})
	}

	if len(failed) > 0 {
		return testResult
	}

	fetcher := metric.MetricValueFetcher{}
	values, err := fetcher.Fetch(namespace, metricName, dims, metric.AVERAGE)
	if err != nil {
		return testResult
	}

	if !isAllValuesGreaterThanOrEqualToZero(metricName, values) {
		return testResult
	}

	testResult.Status = status.SUCCESSFUL
	return testResult
}
