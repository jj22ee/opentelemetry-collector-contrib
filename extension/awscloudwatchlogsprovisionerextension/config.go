// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package awscloudwatchlogsprovisionerextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/awscloudwatchlogsprovisionerextension"

import (
	"errors"

	"go.opentelemetry.io/collector/component"
)

// Config for the awscloudwatchlogsprovisioner extension.
type Config struct {
	// AdditionalAuth is a reference to the inner auth extension (typically sigv4auth)
	// that this extension chains with for request signing. Follows the same pattern
	// as headers_setter's additional_auth field.
	AdditionalAuth *component.ID `mapstructure:"additional_auth"`

	// LogGroupName is the log group name template (required). Placeholders like
	// {service.name}, {k8s.pod.name}, etc. are resolved from client.Metadata.
	// Example: "/aws/telemetry/{service.name}"
	LogGroupName string `mapstructure:"log_group_name"`

	// LogStreamName is the log stream name template.
	// Default: "default"
	LogStreamName string `mapstructure:"log_stream_name"`

	// DefaultPlaceholderValue is the fallback value when a placeholder cannot
	// be resolved from client.Metadata.
	// Default: "undefined"
	DefaultPlaceholderValue string `mapstructure:"default_placeholder_value,omitempty"`

	// LogsProvisionTimeoutSeconds is the HTTP timeout for each CreateLogGroup/CreateLogStream
	// API call (including SDK retries). Bounds how long singleflight waiters block.
	// Default: 10 seconds.
	LogsProvisionTimeoutSeconds int `mapstructure:"logs_provision_timeout_seconds,omitempty"`

	// LogsProvisionFailureBackoffSeconds is the TTL for negative cache entries.
	// During this period, the extension won't retry creation for the same (group, stream) pair.
	// Default: 30 seconds.
	LogsProvisionFailureBackoffSeconds int `mapstructure:"logs_provision_failure_backoff_seconds,omitempty"`
}

var _ component.Config = (*Config)(nil)

func (cfg *Config) Validate() error {
	if cfg.LogGroupName == "" {
		return errors.New("log_group_name is required")
	}
	if cfg.LogStreamName == "" {
		return errors.New("log_stream_name must not be empty")
	}
	return nil
}
