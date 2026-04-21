// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package attributestocontextprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributestocontextprocessor"

import (
	"context"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributestocontextprocessor/internal/actions"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
)

type tracesProcessor struct {
	actions actions.Actions
	next    consumer.Traces
}

func newTracesProcessor(cfg *Config, next consumer.Traces) processor.Traces {
	return &tracesProcessor{
		actions: actions.NewActions(cfg.Actions),
		next:    next,
	}
}

func (p *tracesProcessor) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	clientInfo := client.FromContext(ctx)
	metadataMap := make(map[string][]string)
	for key := range clientInfo.Metadata.Keys() {
		metadataMap[key] = clientInfo.Metadata.Get(key)
	}

	p.actions.ProcessStatic(metadataMap)

	if p.actions.HasResourceActions() || p.actions.HasAttributeActions() {
		resourceSpans := td.ResourceSpans()
		for i := 0; i < resourceSpans.Len(); i++ {
			rs := resourceSpans.At(i)

			if p.actions.HasResourceActions() {
				p.actions.ProcessResource(metadataMap, rs.Resource().Attributes())
			}

			if p.actions.HasAttributeActions() {
				scopeSpans := rs.ScopeSpans()
				for j := 0; j < scopeSpans.Len(); j++ {
					spans := scopeSpans.At(j).Spans()
					for k := 0; k < spans.Len(); k++ {
						p.actions.ProcessAttributes(metadataMap, spans.At(k).Attributes())
					}
				}
			}
		}
	}

	clientInfo.Metadata = client.NewMetadata(metadataMap)
	newCtx := client.NewContext(ctx, clientInfo)
	return p.next.ConsumeTraces(newCtx, td)
}

func (*tracesProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (*tracesProcessor) Start(context.Context, component.Host) error {
	return nil
}

func (*tracesProcessor) Shutdown(context.Context) error {
	return nil
}
