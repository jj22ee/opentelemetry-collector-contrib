// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package attributestocontextprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributestocontextprocessor"

import (
	"context"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributestocontextprocessor/internal/actions"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributestocontextprocessor/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithLogs(createLogsProcessor, metadata.LogsStability),
		processor.WithTraces(createTracesProcessor, metadata.TracesStability),
		processor.WithMetrics(createMetricsProcessor, metadata.MetricsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		Actions: []actions.KeyValue{},
	}
}

func createLogsProcessor(
	_ context.Context,
	_ processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (processor.Logs, error) {
	return newLogsProcessor(cfg.(*Config), nextConsumer), nil
}

func createTracesProcessor(
	_ context.Context,
	_ processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (processor.Traces, error) {
	return newTracesProcessor(cfg.(*Config), nextConsumer), nil
}

func createMetricsProcessor(
	_ context.Context,
	_ processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	return newMetricsProcessor(cfg.(*Config), nextConsumer), nil
}
