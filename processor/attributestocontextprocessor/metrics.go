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

	resourceMetrics := md.ResourceMetrics()
	for i := 0; i < resourceMetrics.Len(); i++ {
		p.actions.ProcessResource(metadataMap, resourceMetrics.At(i).Resource().Attributes())
	}

	clientInfo.Metadata = client.NewMetadata(metadataMap)
	newCtx := client.NewContext(ctx, clientInfo)
	return p.next.ConsumeMetrics(newCtx, md)
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
