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
	assert.Nil(t, cfg.AdditionalAuth)
}

func TestConfig_WithAdditionalAuth(t *testing.T) {
	authID := component.MustNewID("sigv4auth")
	cfg := &Config{
		AdditionalAuth: &authID,
	}

	assert.Equal(t, "sigv4auth", cfg.AdditionalAuth.String())
}

func TestFactory_Type(t *testing.T) {
	f := NewFactory()
	assert.Equal(t, component.MustNewType("awscloudwatchlogsprovisioner"), f.Type())
}
