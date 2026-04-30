// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package attributestocontextprocessor

import (
	"context"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributestocontextprocessor/internal/actions"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestMetricsProcessor(t *testing.T) {
	cfg := &Config{
		Actions: []actions.KeyValue{
			{Key: "service", FromResourceAttribute: "service.name"},
		},
	}

	var capturedCtx context.Context
	next := &mockMetricsConsumer{
		consumeFunc: func(ctx context.Context, _ pmetric.Metrics) error {
			capturedCtx = ctx
			return nil
		},
	}
	processor := newMetricsProcessor(cfg, next)

	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("service.name", "my-service")

	ctx := client.NewContext(t.Context(), client.Info{})
	err := processor.ConsumeMetrics(ctx, md)

	assert.NoError(t, err)
	assert.False(t, processor.Capabilities().MutatesData)

	clientInfo := client.FromContext(capturedCtx)
	assert.Equal(t, []string{"my-service"}, clientInfo.Metadata.Get("service"))
}

type mockMetricsConsumer struct {
	consumeFunc func(ctx context.Context, md pmetric.Metrics) error
}

func (m *mockMetricsConsumer) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	return m.consumeFunc(ctx, md)
}

func (*mockMetricsConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func TestMetricsProcessorStart(t *testing.T) {
	processor := &metricsProcessor{}
	err := processor.Start(t.Context(), nil)
	assert.NoError(t, err)
}

func TestMetricsProcessorShutdown(t *testing.T) {
	processor := &metricsProcessor{}
	err := processor.Shutdown(t.Context())
	assert.NoError(t, err)
}
