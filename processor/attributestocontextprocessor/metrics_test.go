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
	testMetrics := pmetric.NewMetrics()
	rm := testMetrics.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("resource.attribute1", "resource-value")

	sm := rm.ScopeMetrics().AppendEmpty()
	metric := sm.Metrics().AppendEmpty()
	metric.SetName("test_gauge")
	gauge := metric.SetEmptyGauge()
	dp := gauge.DataPoints().AppendEmpty()
	dp.Attributes().PutStr("attribute1", "attribute-value")

	testCases := []struct {
		name     string
		config   *Config
		expected map[string][]string
	}{
		{
			name: "resource attribute extraction",
			config: &Config{
				Actions: []actions.KeyValue{
					{Key: "key1", Action: actions.INSERT, FromResourceAttribute: "resource.attribute1"},
				},
			},
			expected: map[string][]string{
				"key1": {"resource-value"},
			},
		},
		{
			name: "static value",
			config: &Config{
				Actions: []actions.KeyValue{
					{Key: "key2", Action: actions.INSERT, Value: "static-value"},
				},
			},
			expected: map[string][]string{
				"key2": {"static-value"},
			},
		},
		{
			name: "metric attribute extraction",
			config: &Config{
				Actions: []actions.KeyValue{
					{Key: "key3", Action: actions.INSERT, FromAttribute: "attribute1"},
				},
			},
			expected: map[string][]string{
				"key3": {"attribute-value"},
			},
		},
		{
			name: "multiple extractions",
			config: &Config{
				Actions: []actions.KeyValue{
					{Key: "key1", Action: actions.INSERT, FromResourceAttribute: "resource.attribute1"},
					{Key: "key2", Action: actions.INSERT, Value: "static-value"},
					{Key: "key3", Action: actions.INSERT, FromAttribute: "attribute1"},
				},
			},
			expected: map[string][]string{
				"key1": {"resource-value"},
				"key2": {"static-value"},
				"key3": {"attribute-value"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedCtx context.Context
			next := &mockMetricsConsumer{
				consumeFunc: func(ctx context.Context, _ pmetric.Metrics) error {
					capturedCtx = ctx
					return nil
				},
			}
			processor := newMetricsProcessor(tc.config, next)

			ctx := client.NewContext(t.Context(), client.Info{})
			err := processor.ConsumeMetrics(ctx, testMetrics)

			assert.NoError(t, err)

			// Verify metadata was updated
			clientInfo := client.FromContext(capturedCtx)
			for key, expectedValues := range tc.expected {
				assert.Equal(t, expectedValues, clientInfo.Metadata.Get(key))
			}
		})
	}
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
