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
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestTracesProcessor(t *testing.T) {
	cfg := &Config{
		Actions: []actions.KeyValue{
			{Key: "key1", Action: actions.INSERT, FromResourceAttribute: "resource.attribute1"},
			{Key: "key2", Action: actions.INSERT, Value: "static-value"},
		},
	}

	var capturedCtx context.Context
	next := &mockTracesConsumer{
		consumeFunc: func(ctx context.Context, _ ptrace.Traces) error {
			capturedCtx = ctx
			return nil
		},
	}
	processor := newTracesProcessor(cfg, next)

	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	rs.Resource().Attributes().PutStr("resource.attribute1", "resource-value")

	ctx := client.NewContext(t.Context(), client.Info{})
	err := processor.ConsumeTraces(ctx, traces)

	assert.NoError(t, err)
	assert.False(t, processor.Capabilities().MutatesData)

	// Verify metadata was updated
	clientInfo := client.FromContext(capturedCtx)
	assert.Equal(t, []string{"resource-value"}, clientInfo.Metadata.Get("key1"))
	assert.Equal(t, []string{"static-value"}, clientInfo.Metadata.Get("key2"))
}

type mockTracesConsumer struct {
	consumeFunc func(ctx context.Context, td ptrace.Traces) error
}

func (m *mockTracesConsumer) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	return m.consumeFunc(ctx, td)
}

func (*mockTracesConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func TestTracesProcessorStart(t *testing.T) {
	processor := &tracesProcessor{}
	err := processor.Start(t.Context(), nil)
	assert.NoError(t, err)
}

func TestTracesProcessorShutdown(t *testing.T) {
	processor := &tracesProcessor{}
	err := processor.Shutdown(t.Context())
	assert.NoError(t, err)
}
