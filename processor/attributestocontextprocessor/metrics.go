// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package attributestocontextprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributestocontextprocessor"

import (
	"context"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributestocontextprocessor/internal/actions"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
)

type metricsProcessor struct {
	actions actions.Actions
	next    consumer.Metrics
}

func newMetricsProcessor(cfg *Config, next consumer.Metrics) processor.Metrics {
	return &metricsProcessor{
		actions: actions.NewActions(cfg.Actions),
		next:    next,
	}
}

func (p *metricsProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	clientInfo := client.FromContext(ctx)
	metadataMap := make(map[string][]string)
	for key := range clientInfo.Metadata.Keys() {
		metadataMap[key] = clientInfo.Metadata.Get(key)
	}

	p.actions.ProcessStatic(metadataMap)

	if p.actions.HasResourceActions() || p.actions.HasAttributeActions() {
		resourceMetrics := md.ResourceMetrics()
		for i := 0; i < resourceMetrics.Len(); i++ {
			rm := resourceMetrics.At(i)

			if p.actions.HasResourceActions() {
				p.actions.ProcessResource(metadataMap, rm.Resource().Attributes())
			}

			if p.actions.HasAttributeActions() {
				scopeMetrics := rm.ScopeMetrics()
				for j := 0; j < scopeMetrics.Len(); j++ {
					metrics := scopeMetrics.At(j).Metrics()
					for k := 0; k < metrics.Len(); k++ {
						p.extractFromMetric(metrics.At(k), metadataMap)
					}
				}
			}
		}
	}

	clientInfo.Metadata = client.NewMetadata(metadataMap)
	newCtx := client.NewContext(ctx, clientInfo)
	return p.next.ConsumeMetrics(newCtx, md)
}

func (p *metricsProcessor) extractFromMetric(metric pmetric.Metric, metadataMap map[string][]string) {
	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		dataPoints := metric.Gauge().DataPoints()
		for i := 0; i < dataPoints.Len(); i++ {
			p.actions.ProcessAttributes(metadataMap, dataPoints.At(i).Attributes())
		}
	case pmetric.MetricTypeSum:
		dataPoints := metric.Sum().DataPoints()
		for i := 0; i < dataPoints.Len(); i++ {
			p.actions.ProcessAttributes(metadataMap, dataPoints.At(i).Attributes())
		}
	case pmetric.MetricTypeHistogram:
		dataPoints := metric.Histogram().DataPoints()
		for i := 0; i < dataPoints.Len(); i++ {
			p.actions.ProcessAttributes(metadataMap, dataPoints.At(i).Attributes())
		}
	case pmetric.MetricTypeExponentialHistogram:
		dataPoints := metric.ExponentialHistogram().DataPoints()
		for i := 0; i < dataPoints.Len(); i++ {
			p.actions.ProcessAttributes(metadataMap, dataPoints.At(i).Attributes())
		}
	case pmetric.MetricTypeSummary:
		dataPoints := metric.Summary().DataPoints()
		for i := 0; i < dataPoints.Len(); i++ {
			p.actions.ProcessAttributes(metadataMap, dataPoints.At(i).Attributes())
		}
	}
}

func (*metricsProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (*metricsProcessor) Start(context.Context, component.Host) error {
	return nil
}

func (*metricsProcessor) Shutdown(context.Context) error {
	return nil
}
