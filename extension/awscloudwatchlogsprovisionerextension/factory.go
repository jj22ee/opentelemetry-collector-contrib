// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package awscloudwatchlogsprovisionerextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/awscloudwatchlogsprovisionerextension"

import (
	"context"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/awscloudwatchlogsprovisionerextension/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

func NewFactory() extension.Factory {
	return extension.NewFactory(
		metadata.Type,
		createDefaultConfig,
		createExtension,
		metadata.ExtensionStability,
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		// TODO: Update default log group name before PR is merged.
		LogGroupName:                       "/aws/telemetry/{ServiceName}",
		LogStreamName:                      "default",
		DefaultPlaceholderValue:            "undefined",
		LogsProvisionTimeoutSeconds:        10,
		LogsProvisionFailureBackoffSeconds: 30,
	}
}

func createExtension(_ context.Context, settings extension.Settings, cfg component.Config) (extension.Extension, error) {
	return newExtension(settings.Logger, cfg.(*Config)), nil
}
