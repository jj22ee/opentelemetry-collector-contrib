// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestProcessResource(t *testing.T) {
	a := NewActions([]KeyValue{
		{Key: "service", FromResourceAttribute: "service.name"},
		{Key: "host", FromResourceAttribute: "host.name"},
	})

	attrs := pcommon.NewMap()
	attrs.PutStr("service.name", "my-service")
	attrs.PutStr("host.name", "my-host")

	metadata := make(map[string][]string)
	a.ProcessResource(metadata, attrs)

	assert.Equal(t, []string{"my-service"}, metadata["service"])
	assert.Equal(t, []string{"my-host"}, metadata["host"])
}

func TestProcessResource_OverwritesExisting(t *testing.T) {
	a := NewActions([]KeyValue{
		{Key: "service", FromResourceAttribute: "service.name"},
	})

	attrs := pcommon.NewMap()
	attrs.PutStr("service.name", "new-service")

	metadata := map[string][]string{
		"service": {"old-service"},
	}
	a.ProcessResource(metadata, attrs)

	assert.Equal(t, []string{"new-service"}, metadata["service"])
}

func TestProcessResource_MissingAttribute(t *testing.T) {
	a := NewActions([]KeyValue{
		{Key: "missing", FromResourceAttribute: "not.found"},
	})

	attrs := pcommon.NewMap()
	metadata := make(map[string][]string)
	a.ProcessResource(metadata, attrs)

	assert.Empty(t, metadata)
}
