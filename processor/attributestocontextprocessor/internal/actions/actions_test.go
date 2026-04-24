// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestNewActions(t *testing.T) {
	cfg := []KeyValue{
		{Key: "key1", Action: INSERT, FromResourceAttribute: "resource.attribute1"},
		{Key: "key2", Action: INSERT, FromAttribute: "attribute1"},
		{Key: "key3", Action: INSERT, Value: "static-value"},
		{Key: "key4", Action: DELETE},
	}

	a := NewActions(cfg)

	// Test that actions are properly categorized by testing their behavior
	metadata := map[string][]string{
		"key4": {"some", "value"},
	}

	// Test static actions
	a.ProcessStatic(metadata)
	assert.Equal(t, []string{"static-value"}, metadata["key3"])

	// Test resource actions
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("resource.attribute1", "resource-value")
	a.ProcessResource(metadata, resourceAttrs)
	assert.Equal(t, []string{"resource-value"}, metadata["key1"])

	// Test attribute actions
	attrs := pcommon.NewMap()
	attrs.PutStr("attribute1", "attr-value")
	a.ProcessAttributes(metadata, attrs)
	assert.Equal(t, []string{"attr-value"}, metadata["key2"])
}

func TestProcessStatic(t *testing.T) {
	testCases := []struct {
		name     string
		actions  []KeyValue
		initial  map[string][]string
		expected map[string][]string
	}{
		{
			name: "insert static value",
			actions: []KeyValue{
				{Key: "env", Action: INSERT, Value: "prod"},
			},
			initial:  make(map[string][]string),
			expected: map[string][]string{"env": {"prod"}},
		},
		{
			name: "update static value",
			actions: []KeyValue{
				{Key: "env", Action: UPDATE, Value: "beta"},
			},
			initial:  map[string][]string{"env": {"prod"}},
			expected: map[string][]string{"env": {"beta"}},
		},
		{
			name: "upsert static value",
			actions: []KeyValue{
				{Key: "env", Action: UPSERT, Value: "staging"},
			},
			initial:  make(map[string][]string),
			expected: map[string][]string{"env": {"staging"}},
		},
		{
			name: "delete key",
			actions: []KeyValue{
				{Key: "old-key", Action: DELETE},
			},
			initial:  map[string][]string{"old-key": {"old-value"}},
			expected: make(map[string][]string),
		},
		{
			name: "insert does not overwrite existing",
			actions: []KeyValue{
				{Key: "env", Action: INSERT, Value: "prod"},
			},
			initial:  map[string][]string{"env": {"existing"}},
			expected: map[string][]string{"env": {"existing"}},
		},
		{
			name: "update does not create new key",
			actions: []KeyValue{
				{Key: "env", Action: UPDATE, Value: "prod"},
			},
			initial:  make(map[string][]string),
			expected: make(map[string][]string),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewActions(tc.actions)
			metadata := make(map[string][]string)
			for k, v := range tc.initial {
				metadata[k] = v
			}

			a.ProcessStatic(metadata)

			assert.Equal(t, tc.expected, metadata)
		})
	}
}

func TestProcessResource(t *testing.T) {
	actions := []KeyValue{
		{Key: "service", Action: INSERT, FromResourceAttribute: "service.name"},
		{Key: "host", Action: UPDATE, FromResourceAttribute: "host.name"},
		{Key: "env", Action: UPSERT, FromResourceAttribute: "environment"},
	}
	a := NewActions(actions)

	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("service.name", "my-service")
	resourceAttrs.PutStr("host.name", "my-host")
	resourceAttrs.PutStr("environment", "prod")

	metadata := map[string][]string{
		"host":    {"old-host"},    // existing for UPDATE
		"service": {"old-service"}, // existing for INSERT (should not change)
	}

	a.ProcessResource(metadata, resourceAttrs)

	assert.Equal(t, []string{"old-service"}, metadata["service"]) // INSERT unchanged
	assert.Equal(t, []string{"my-host"}, metadata["host"])        // UPDATE changed
	assert.Equal(t, []string{"prod"}, metadata["env"])            // UPSERT added
}

func TestProcessAttributes(t *testing.T) {
	actions := []KeyValue{
		{Key: "span-id", Action: INSERT, FromAttribute: "span.id"},
		{Key: "trace-id", Action: UPSERT, FromAttribute: "trace.id"},
	}
	a := NewActions(actions)

	attrs := pcommon.NewMap()
	attrs.PutStr("span.id", "abc123")
	attrs.PutStr("trace.id", "def456")

	metadata := map[string][]string{
		"trace-id": {"old-trace"}, // existing for UPSERT
	}

	a.ProcessAttributes(metadata, attrs)

	assert.Equal(t, []string{"abc123"}, metadata["span-id"])  // INSERT added
	assert.Equal(t, []string{"def456"}, metadata["trace-id"]) // UPSERT changed
}

func TestProcessAttributesNotFound(t *testing.T) {
	actions := []KeyValue{
		{Key: "missing", Action: INSERT, FromAttribute: "not.found"},
	}
	a := NewActions(actions)

	attrs := pcommon.NewMap()
	metadata := make(map[string][]string)

	a.ProcessAttributes(metadata, attrs)

	assert.Empty(t, metadata) // No action should be taken for missing attributes
}
