// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package awscloudwatchlogsprovisionerextension

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/extension/extensiontest"
)

func TestNewFactory(t *testing.T) {
	f := NewFactory()
	assert.NotNil(t, f)
	assert.Equal(t, "awscloudwatchlogsprovisioner", f.Type().String())
}

func TestCreateDefaultConfig(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	assert.Empty(t, cfg.LogGroupName)
	assert.Empty(t, cfg.LogGroupContextKey)
	assert.Equal(t, 10, cfg.LogsProvisionTimeoutSeconds)
	assert.Equal(t, 30, cfg.LogsProvisionFailureBackoffSeconds)
	assert.Nil(t, cfg.AdditionalAuth)
	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
}

func TestCreateExtension(t *testing.T) {
	f := NewFactory()
	cfg := createDefaultConfig()
	ext, err := f.Create(t.Context(), extensiontest.NewNopSettings(f.Type()), cfg)
	require.NoError(t, err)
	assert.NotNil(t, ext)
}
