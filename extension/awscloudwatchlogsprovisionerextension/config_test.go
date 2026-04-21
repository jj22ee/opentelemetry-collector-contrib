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

	assert.Equal(t, "/aws/telemetry/{ServiceName}", cfg.LogGroupName)
	assert.Equal(t, "default", cfg.LogStreamName)
	assert.Equal(t, "undefined", cfg.DefaultPlaceholderValue)
	assert.Equal(t, 30, cfg.LogsProvisionFailureBackoffSeconds)
	assert.Nil(t, cfg.AdditionalAuth)
}

func TestConfig_WithAuth(t *testing.T) {
	authID := component.MustNewID("sigv4auth")
	cfg := &Config{
		AdditionalAuth:          &authID,
		Region:                  "us-west-2",
		LogGroupName:            "/custom/{service.name}",
		LogStreamName:           "{host.id}",
		DefaultPlaceholderValue: "fallback",
		LogsProvisionFailureBackoffSeconds:   60,
	}

	assert.Equal(t, "sigv4auth", cfg.AdditionalAuth.String())
	assert.Equal(t, "us-west-2", cfg.Region)
	assert.Equal(t, "/custom/{service.name}", cfg.LogGroupName)
	assert.Equal(t, "{host.id}", cfg.LogStreamName)
	assert.Equal(t, "fallback", cfg.DefaultPlaceholderValue)
	assert.Equal(t, 60, cfg.LogsProvisionFailureBackoffSeconds)
}

func TestFactory_Type(t *testing.T) {
	f := NewFactory()
	assert.Equal(t, component.MustNewType("awscloudwatchlogsprovisioner"), f.Type())
}
