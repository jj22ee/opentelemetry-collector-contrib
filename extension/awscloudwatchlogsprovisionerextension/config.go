// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package awscloudwatchlogsprovisionerextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/awscloudwatchlogsprovisionerextension"

import (
	"go.opentelemetry.io/collector/component"
)

// Config for the awscloudwatchlogsprovisioner extension.
//
// The extension reads x-aws-log-group and x-aws-log-stream headers from outgoing
// requests and creates the corresponding CloudWatch log groups and streams.
//
// Header resolution priority:
//  1. Start with whatever x-aws-log-group/x-aws-log-stream headers are already
//     on the request (e.g., set by otlphttp exporter static headers or headers_setter).
//  2. If log_group_context_key is set, override x-aws-log-group with the value
//     from client.Metadata at that key.
//  3. If log_stream_context_key is set, override x-aws-log-stream with the value
//     from client.Metadata at that key.
//
// This means:
//   - Static case: No context keys needed. Configure x-aws-log-group in the
//     otlphttp exporter headers. The extension just provisions whatever it sees.
//   - Dynamic case: Set context keys. An upstream processor (e.g., transform +
//     attributestocontext) populates client.Metadata with full log group/stream
//     names, and the extension overrides the headers and provisions.
type Config struct {
	// AdditionalAuth is a reference to the inner auth extension (typically sigv4auth)
	// that this extension chains with for request signing. Follows the same pattern
	// as headers_setter's additional_auth field.
	AdditionalAuth *component.ID `mapstructure:"additional_auth"`

	// LogGroupContextKey is an optional client.Metadata key to read the log group
	// name from. If set, overrides whatever x-aws-log-group header is on the request.
	LogGroupContextKey string `mapstructure:"log_group_context_key,omitempty"`

	// LogStreamContextKey is an optional client.Metadata key to read the log stream
	// name from. If set, overrides whatever x-aws-log-stream header is on the request.
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
