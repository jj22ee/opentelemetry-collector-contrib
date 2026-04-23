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

	assert.Equal(t, "", cfg.LogGroupName)
	assert.Equal(t, "default", cfg.LogStreamName)
	assert.Equal(t, "undefined", cfg.DefaultPlaceholderValue)
	assert.Equal(t, 10, cfg.LogsProvisionTimeoutSeconds)
	assert.Equal(t, 30, cfg.LogsProvisionFailureBackoffSeconds)
	assert.Nil(t, cfg.AdditionalAuth)
}

func TestConfig_WithAuth(t *testing.T) {
	authID := component.MustNewID("sigv4auth")
	cfg := &Config{
		AdditionalAuth:                     &authID,
		LogGroupName:                       "/custom/{service.name}",
		LogStreamName:                      "{host.id}",
		DefaultPlaceholderValue:            "fallback",
		LogsProvisionFailureBackoffSeconds: 60,
	}

	assert.Equal(t, "sigv4auth", cfg.AdditionalAuth.String())
	assert.Equal(t, "/custom/{service.name}", cfg.LogGroupName)
	assert.Equal(t, "{host.id}", cfg.LogStreamName)
	assert.Equal(t, "fallback", cfg.DefaultPlaceholderValue)
	assert.Equal(t, 60, cfg.LogsProvisionFailureBackoffSeconds)
}

func TestConfig_Validate_MissingLogGroupName(t *testing.T) {
	cfg := &Config{}
	assert.EqualError(t, cfg.Validate(), "log_group_name is required")
}

func TestConfig_Validate_EmptyLogStreamName(t *testing.T) {
	cfg := &Config{LogGroupName: "/test/{service.name}", LogStreamName: ""}
	assert.EqualError(t, cfg.Validate(), "log_stream_name must not be empty")
}

func TestConfig_Validate_EmptyDefaultPlaceholderValue(t *testing.T) {
	cfg := &Config{LogGroupName: "/test/{service.name}", LogStreamName: "default", DefaultPlaceholderValue: ""}
	assert.EqualError(t, cfg.Validate(), "default_placeholder_value must not be empty")
}

func TestConfig_Validate_Valid(t *testing.T) {
	cfg := &Config{LogGroupName: "/test/{service.name}", LogStreamName: "default", DefaultPlaceholderValue: "undefined"}
	assert.NoError(t, cfg.Validate())
}

func TestFactory_Type(t *testing.T) {
	f := NewFactory()
	assert.Equal(t, component.MustNewType("awscloudwatchlogsprovisioner"), f.Type())
}
