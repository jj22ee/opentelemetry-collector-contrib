// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package attributestocontextprocessor

import (
	"path/filepath"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributestocontextprocessor/internal/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/confmap/xconfmap"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id       component.ID
		expected component.Config
	}{
		{
			id: component.NewIDWithName(component.MustNewType("attributestocontext"), "insert"),
			expected: &Config{
				Actions: []actions.KeyValue{
					{Key: "key1", Action: actions.INSERT, Value: "static-value"},
					{Key: "key2", Action: actions.INSERT, FromResourceAttribute: "resource.attribute1"},
				},
			},
		},
		{
			id: component.NewIDWithName(component.MustNewType("attributestocontext"), "update"),
			expected: &Config{
				Actions: []actions.KeyValue{
					{Key: "key1", Action: actions.UPDATE, FromAttribute: "attribute1"},
					{Key: "key2", Action: actions.UPDATE, Value: "updated-value"},
				},
			},
		},
		{
			id: component.NewIDWithName(component.MustNewType("attributestocontext"), "upsert"),
			expected: &Config{
				Actions: []actions.KeyValue{
					{Key: "key1", Action: actions.UPSERT, FromResourceAttribute: "service.name"},
					{Key: "key2", Action: actions.UPSERT, Value: "upserted-value"},
				},
			},
		},
		{
			id: component.NewIDWithName(component.MustNewType("attributestocontext"), "delete"),
			expected: &Config{
				Actions: []actions.KeyValue{
					{Key: "old-key", Action: actions.DELETE},
				},
			},
		},
		{
			id: component.NewIDWithName(component.MustNewType("attributestocontext"), "mixed"),
			expected: &Config{
				Actions: []actions.KeyValue{
					{Key: "static-key", Action: actions.INSERT, Value: "static-value"},
					{Key: "resource-key", Action: actions.UPDATE, FromResourceAttribute: "host.name"},
					{Key: "attribute-key", Action: actions.UPSERT, FromAttribute: "span.id"},
					{Key: "remove-key", Action: actions.DELETE},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
			require.NoError(t, err)

			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, sub.Unmarshal(cfg))

			assert.NoError(t, xconfmap.Validate(cfg))
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		config  *Config
		wantErr string
	}{
		{
			name: "valid delete action",
			config: &Config{
				Actions: []actions.KeyValue{
					{Key: "key1", Action: actions.DELETE},
				},
			},
			wantErr: "",
		},
		{
			name: "empty actions",
			config: &Config{
				Actions: []actions.KeyValue{},
			},
			wantErr: "missing required field \"actions\"",
		},
		{
			name: "missing key",
			config: &Config{
				Actions: []actions.KeyValue{
					{Action: actions.INSERT, Value: "test"},
				},
			},
			wantErr: "action 0: missing required field \"key\"",
		},
		{
			name: "missing action",
			config: &Config{
				Actions: []actions.KeyValue{
					{Key: "test", Value: "test"},
				},
			},
			wantErr: "action 0: missing required field \"action\"",
		},
		{
			name: "no source specified",
			config: &Config{
				Actions: []actions.KeyValue{
					{Key: "test", Action: actions.INSERT},
				},
			},
			wantErr: "action 0: exactly one of \"value\", \"from_attribute\", or \"from_resource_attribute\" must be specified",
		},
		{
			name: "multiple sources specified",
			config: &Config{
				Actions: []actions.KeyValue{
					{Key: "test", Action: actions.INSERT, Value: "test", FromAttribute: "attr"},
				},
			},
			wantErr: "action 0: exactly one of \"value\", \"from_attribute\", or \"from_resource_attribute\" must be specified",
		},
		{
			name: "delete action with value source",
			config: &Config{
				Actions: []actions.KeyValue{
					{Key: "test", Action: actions.DELETE, Value: "test"},
				},
			},
			wantErr: "action 0: DELETE action should not specify value sources",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}
