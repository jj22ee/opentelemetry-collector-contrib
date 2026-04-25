// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package awscloudwatchlogsprovisionerextension

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
)

func TestConfig_Defaults(t *testing.T) {
	cfg := createDefaultConfig().(*Config)

	assert.Equal(t, 10, cfg.LogsProvisionTimeoutSeconds)
	assert.Equal(t, 30, cfg.LogsProvisionFailureBackoffSeconds)
	assert.Empty(t, cfg.LogGroupContextKey)
	assert.Empty(t, cfg.LogStreamContextKey)
}

func TestConfig_WithContextKeys(t *testing.T) {
	authID := component.MustNewID("sigv4auth")
	cfg := &Config{
		AdditionalAuth:      &authID,
		Region:              "us-west-2",
		LogGroupContextKey:  "cwlogs.log_group",
		LogStreamContextKey: "cwlogs.log_stream",
	}

	assert.Equal(t, "sigv4auth", cfg.AdditionalAuth.String())
	assert.Equal(t, "cwlogs.log_group", cfg.LogGroupContextKey)
	assert.Equal(t, "cwlogs.log_stream", cfg.LogStreamContextKey)
}

func TestConfig_NoContextKeys(t *testing.T) {
	// Valid: no context keys — extension just provisions whatever headers are on the request
	cfg := &Config{}
	assert.Empty(t, cfg.LogGroupContextKey)
}

func TestFactory_Type(t *testing.T) {
	f := NewFactory()
	assert.Equal(t, component.MustNewType("awscloudwatchlogsprovisioner"), f.Type())
}
