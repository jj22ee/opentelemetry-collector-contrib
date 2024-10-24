// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package awsxrayexporter // import "github.com/jj22ee/opentelemetry-collector-contrib/exporter/awsxrayexporter"

import (
	"go.opentelemetry.io/collector/component"

	"github.com/jj22ee/opentelemetry-collector-contrib/internal/aws/awsutil"
	"github.com/jj22ee/opentelemetry-collector-contrib/internal/aws/xray/telemetry"
)

// Config defines configuration for AWS X-Ray exporter.
type Config struct {
	// AWSSessionSettings contains the common configuration options
	// for creating AWS session to communicate with backend
	awsutil.AWSSessionSettings `mapstructure:",squash"`
	// By default, OpenTelemetry attributes are converted to X-Ray metadata, which are not indexed.
	// Specify a list of attribute names to be converted to X-Ray annotations instead, which will be indexed.
	// See annotation vs. metadata: https://docs.aws.amazon.com/xray/latest/devguide/xray-concepts.html#xray-concepts-annotations
	IndexedAttributes []string `mapstructure:"indexed_attributes"`
	// Set to true to convert all OpenTelemetry attributes to X-Ray annotation (indexed) ignoring the IndexedAttributes option.
	// Default value: false
	IndexAllAttributes bool `mapstructure:"index_all_attributes"`

	LogGroupNames []string `mapstructure:"aws_log_groups"`
	// TelemetryConfig contains the options for telemetry collection.
	TelemetryConfig telemetry.Config `mapstructure:"telemetry,omitempty"`
	// MiddlewareID is an ID for an extension that can be used to configure the
	// AWS client.
	MiddlewareID *component.ID `mapstructure:"middleware,omitempty"`

	// X-Ray Export sends spans in its original otlp format to X-Ray Service when this flag is on
	TransitSpansInOtlpFormat bool `mapstructure:"transit_spans_in_otlp_format,omitempty"`

	// skipTimestampValidation if enabled, will skip timestamp validation logic on the trace ID
	skipTimestampValidation bool
}
