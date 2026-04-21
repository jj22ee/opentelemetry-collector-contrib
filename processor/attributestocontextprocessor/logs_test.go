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
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestLogsProcessor(t *testing.T) {
	cfg := &Config{
		Actions: []actions.KeyValue{
			{Key: "key1", Action: actions.INSERT, FromResourceAttribute: "resource.attribute1"},
			{Key: "key2", Action: actions.INSERT, Value: "static-value"},
		},
	}

	var capturedCtx context.Context
	next := &mockLogsConsumer{
		consumeFunc: func(ctx context.Context, _ plog.Logs) error {
			capturedCtx = ctx
			return nil
		},
	}
	processor := newLogsProcessor(cfg, next)

	logs := plog.NewLogs()
	rl := logs.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr("resource.attribute1", "resource-value")

	ctx := client.NewContext(t.Context(), client.Info{})
	err := processor.ConsumeLogs(ctx, logs)

	assert.NoError(t, err)
	assert.False(t, processor.Capabilities().MutatesData)

	// Verify metadata was updated
	clientInfo := client.FromContext(capturedCtx)
	assert.Equal(t, []string{"resource-value"}, clientInfo.Metadata.Get("key1"))
	assert.Equal(t, []string{"static-value"}, clientInfo.Metadata.Get("key2"))
}

type mockLogsConsumer struct {
	consumeFunc func(ctx context.Context, ld plog.Logs) error
}

func (m *mockLogsConsumer) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	return m.consumeFunc(ctx, ld)
}

func (*mockLogsConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func TestLogsProcessorStart(t *testing.T) {
	processor := &logsProcessor{}
	err := processor.Start(t.Context(), nil)
	assert.NoError(t, err)
}

func TestLogsProcessorShutdown(t *testing.T) {
	processor := &logsProcessor{}
	err := processor.Shutdown(t.Context())
	assert.NoError(t, err)
}
