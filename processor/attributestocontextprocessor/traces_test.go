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
			{Key: "service", FromResourceAttribute: "service.name"},
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
	rs.Resource().Attributes().PutStr("service.name", "my-service")

	ctx := client.NewContext(t.Context(), client.Info{})
	err := processor.ConsumeTraces(ctx, traces)

	assert.NoError(t, err)
	assert.False(t, processor.Capabilities().MutatesData)

	clientInfo := client.FromContext(capturedCtx)
	assert.Equal(t, []string{"my-service"}, clientInfo.Metadata.Get("service"))
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
