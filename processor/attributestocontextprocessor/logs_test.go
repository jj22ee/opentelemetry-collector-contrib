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
			{Key: "cwlogs.log_group", FromResourceAttribute: "cwlogs.log_group"},
			{Key: "cwlogs.log_stream", FromResourceAttribute: "cwlogs.log_stream"},
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
	rl.Resource().Attributes().PutStr("cwlogs.log_group", "/aws/telemetry/my-service")
	rl.Resource().Attributes().PutStr("cwlogs.log_stream", "default")

	ctx := client.NewContext(t.Context(), client.Info{})
	err := processor.ConsumeLogs(ctx, logs)

	assert.NoError(t, err)
	assert.False(t, processor.Capabilities().MutatesData)

	clientInfo := client.FromContext(capturedCtx)
	assert.Equal(t, []string{"/aws/telemetry/my-service"}, clientInfo.Metadata.Get("cwlogs.log_group"))
	assert.Equal(t, []string{"default"}, clientInfo.Metadata.Get("cwlogs.log_stream"))
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
