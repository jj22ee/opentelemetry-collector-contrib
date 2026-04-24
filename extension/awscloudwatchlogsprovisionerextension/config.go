// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package awscloudwatchlogsprovisionerextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/awscloudwatchlogsprovisionerextension"

import (
	"errors"

	"go.opentelemetry.io/collector/component"
)

// Config for the awscloudwatchlogsprovisioner extension.
//
// Supports two modes (mutually exclusive):
//
// Static mode: log_group_name and log_stream_name are set directly in the config.
// All requests use the same log group/stream.
//
// Dynamic mode: log_group_context_key and log_stream_context_key specify
// client.Metadata keys to read the full log group/stream names from at request
// time. The values should be set by an upstream processor (e.g., transform
// processor + attributestocontext processor). No placeholder resolution is done
// by the extension — the metadata value IS the log group/stream name.
type Config struct {
	// AdditionalAuth is a reference to the inner auth extension (typically sigv4auth)
	// that this extension chains with for request signing. Follows the same pattern
	// as headers_setter's additional_auth field.
	AdditionalAuth *component.ID `mapstructure:"additional_auth"`

	// Region overrides the AWS region for CreateLogGroup/CreateLogStream calls.
	// If empty, the region is extracted from the request URL.
	Region string `mapstructure:"region,omitempty"`

	// --- Static mode (mutually exclusive with context key mode) ---

	// LogGroupName is the static log group name. All requests use this value.
	LogGroupName string `mapstructure:"log_group_name,omitempty"`

	// LogStreamName is the static log stream name.
	LogStreamName string `mapstructure:"log_stream_name,omitempty"`

	// --- Dynamic mode (mutually exclusive with static mode) ---

	// LogGroupContextKey is the client.Metadata key to read the log group name from.
	// The value at this key should be the full log group name (no placeholders).
	LogGroupContextKey string `mapstructure:"log_group_context_key,omitempty"`

	// LogStreamContextKey is the client.Metadata key to read the log stream name from.
	// If the value is empty at request time, falls back to "default".
	LogStreamContextKey string `mapstructure:"log_stream_context_key,omitempty"`

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
	hasStatic := cfg.LogGroupName != ""
	hasContext := cfg.LogGroupContextKey != ""

	if !hasStatic && !hasContext {
		return errors.New("either log_group_name or log_group_context_key must be specified")
	}
	if hasStatic && hasContext {
		return errors.New("log_group_name and log_group_context_key are mutually exclusive")
	}

	if hasStatic {
		if cfg.LogStreamName == "" {
			return errors.New("log_stream_name is required when using log_group_name")
		}
		if cfg.LogStreamContextKey != "" {
			return errors.New("log_stream_context_key cannot be used with log_group_name (use log_stream_name instead)")
		}
	}

	if hasContext {
		if cfg.LogStreamContextKey == "" {
			return errors.New("log_stream_context_key is required when using log_group_context_key")
		}
		if cfg.LogStreamName != "" {
			return errors.New("log_stream_name cannot be used with log_group_context_key (use log_stream_context_key instead)")
		}
	}

	return nil
}
