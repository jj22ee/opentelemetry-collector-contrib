// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package awscloudwatchlogsprovisionerextension

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
)

func TestConfig_Defaults(t *testing.T) {
	cfg := createDefaultConfig().(*Config)

	assert.Equal(t, 10, cfg.LogsProvisionTimeoutSeconds)
	assert.Equal(t, 30, cfg.LogsProvisionFailureBackoffSeconds)
	assert.Empty(t, cfg.LogGroupName)
	assert.Empty(t, cfg.LogGroupContextKey)
}

func TestConfig_ValidateStaticMode(t *testing.T) {
	cfg := &Config{
		LogGroupName:  "/aws/telemetry/my-service",
		LogStreamName: "default",
	}
	assert.NoError(t, cfg.Validate())
}

func TestConfig_ValidateDynamicMode(t *testing.T) {
	cfg := &Config{
		LogGroupContextKey:  "cwlogs.log_group",
		LogStreamContextKey: "cwlogs.log_stream",
	}
	assert.NoError(t, cfg.Validate())
}

func TestConfig_ValidateRejectsEmpty(t *testing.T) {
	cfg := &Config{}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "either log_group_name or log_group_context_key must be specified")
}

func TestConfig_ValidateRejectsBothModes(t *testing.T) {
	cfg := &Config{
		LogGroupName:       "/static/group",
		LogGroupContextKey: "cwlogs.log_group",
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestConfig_ValidateStaticRequiresStreamName(t *testing.T) {
	cfg := &Config{
		LogGroupName: "/static/group",
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "log_stream_name is required")
}

func TestConfig_ValidateDynamicRequiresStreamContextKey(t *testing.T) {
	cfg := &Config{
		LogGroupContextKey: "cwlogs.log_group",
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "log_stream_context_key is required")
}

func TestConfig_ValidateRejectsMixedStaticAndContext(t *testing.T) {
	cfg := &Config{
		LogGroupName:        "/static/group",
		LogStreamName:       "default",
		LogStreamContextKey: "cwlogs.log_stream",
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "log_stream_context_key cannot be used with log_group_name")
}

func TestFactory_Type(t *testing.T) {
	f := NewFactory()
	assert.Equal(t, component.MustNewType("awscloudwatchlogsprovisioner"), f.Type())
}
